package topology

import (
	"k8s.io/utils/cpuset"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

// CPU,Core,Socket,Node
type Topology struct {
	CPU     int
	Core    int
	Socket  int
	Node    int
	Devices []string
}

var (
	// 每 socket 维度
	vastaitechPerSocket = map[int]*RoundRobinBalance{}
	extDevices          map[DeviceType]map[int][]Topology
)

func GetAllExtDevices() map[DeviceType]map[int][]Topology {
	return extDevices
}

func ExistsExtDevices(devType DeviceType) bool {
	devs, ok := extDevices[devType]
	return ok && len(devs) > 0
}

func GetVastaitechBySocket(socket int) []string {
	return vastaitechPerSocket[socket].Next()
}

// GetCGPUTopology  获取 CPU 和 GPU 的拓扑
func GetCGPUTopology() map[int]Topology {
	coreMap := make(map[int]Topology)
	outStr, err := ExecCommand(exec.Command("lscpu", "-p=cpu,core,socket,node"))
	if err != nil {
		log.Println("ERROR: could not interrogate the CPU topology of the node with lscpu, because:" + err.Error())
		return coreMap
	}
	//Here be dragons: we need to manually parse the stdout into a CPU core map line-by-line
	//lscpu -p and -J options are mutually exclusive :(
	for _, lsLine := range strings.Split(strings.TrimSuffix(outStr, "\n"), "\n") {
		cpuInfoStr := strings.Split(lsLine, ",")
		if len(cpuInfoStr) != 4 {
			continue
		}
		cpuInt, cpuErr := strconv.Atoi(cpuInfoStr[0])
		coreInt, coreErr := strconv.Atoi(cpuInfoStr[1])
		socketInt, socketErr := strconv.Atoi(cpuInfoStr[2])
		nodeInt, nodeErr := strconv.Atoi(cpuInfoStr[3])
		if cpuErr != nil || coreErr != nil || socketErr != nil || nodeErr != nil {
			continue
		}
		coreMap[cpuInt] = Topology{
			CPU:    cpuInt,
			Core:   coreInt,
			Socket: socketInt,
			Node:   nodeInt,
		}
	}

	// 多卡情况先计算每张卡应分配次数，然后优先取当前 numa 节点的卡，不够的情况下取同 socket 下使用次数较少的卡
	// 每 socket 卡数量 map[socket]map[numa]
	perSocketGPUs := map[int]map[int][]Topology{}
	// map[socket]map[numa]
	perSocketNumas := map[int]map[int][]Topology{}

	for _, v := range coreMap {
		if _, ok := perSocketNumas[v.Socket]; !ok {
			perSocketNumas[v.Socket] = map[int][]Topology{}
		}
		perSocketNumas[v.Socket][v.Node] = append(perSocketNumas[v.Socket][v.Node], v)
	}

	// 根据 CPU numa 布局获取扩展设备（amdgpu、netint、vastaitech）的信息，
	extDevices = GetExtDevices(coreMap)

	for _, gpu := range extDevices[Vastaitech] {
		for _, gv := range gpu {
			if _, ok := perSocketGPUs[gv.Socket]; !ok {
				perSocketGPUs[gv.Socket] = map[int][]Topology{}
			}
			perSocketGPUs[gv.Socket][gv.Node] = append(perSocketGPUs[gv.Socket][gv.Node], gv)
		}
	}

	for socket, gpuTopol := range perSocketGPUs {
		//fmt.Println("    ", "Socket:", socket, "GPU 数:", len(gpuTopol), "detail: ", gpuTopol)
		if _, ok := vastaitechPerSocket[socket]; !ok {
			vastaitechPerSocket[socket] = &RoundRobinBalance{}
		}
		for _, gpu := range gpuTopol {
			for _, gv := range gpu {
				vastaitechPerSocket[socket].Add(gv.Devices)
			}
		}
	}

	// netint 镕铭微电子
	var netintNum int
	if _, ok := extDevices[Netint]; ok && len(extDevices[Netint]) > 0 {
		netintNum = len(coreMap) / len(extDevices[Netint][0])
	}

	// 标配 2 * AMD W6800  + 1 * NETINT 可以在这里关联到每个 CPU Core，瀚博2卡或多卡无法分配到每个核心，需要单独处理
	for core, cpuTopol := range coreMap {
		// netint 镕铭微电子
		if netintNum > 0 {
			for _, d := range extDevices[Netint][0][core/netintNum].Devices {
				cpuTopol.Devices = append(cpuTopol.Devices, d)
			}
		}

		// amdgpu
		for _, gpu := range extDevices[AMDGPU] {
			for _, gv := range gpu {
				if cpuTopol.Socket == gv.Socket {
					cpuTopol.Devices = append(cpuTopol.Devices, gv.Devices...)
					break
				}
			}
		}
		coreMap[core] = cpuTopol
	}

	return coreMap
}

// GetNodeTopology inspects the node's CPU architecture with lscpu, and returns a map of coreID-NUMA node ID associations
func GetNodeTopology() map[int]int {
	return listAndParseCores("node")
}

// GetHTTopology inspects the node's CPU architecture with lscpu, and returns a map of physical coreID-list of logical coreIDs associations
func GetHTTopology() map[int]string {
	coreMap := listAndParseCores("core")
	htMap := make(map[int]string)
	for logicalCoreID, physicalCoreID := range coreMap {
		//We don't want to duplicate the physical core itself into the logical core ID list
		if physicalCoreID != logicalCoreID {
			logicalCoreIDStr := strconv.Itoa(logicalCoreID)
			if htMap[physicalCoreID] != "" {
				htMap[physicalCoreID] += ","
			}
			htMap[physicalCoreID] += logicalCoreIDStr
		}
	}
	return htMap
}

// AddHTSiblingsToCPUSet takes an allocated exclusive CPU set and expands it with all the sibling threads belonging to the allocated physical cores
func AddHTSiblingsToCPUSet(exclusiveCPUSet cpuset.CPUSet, coreMap map[int]string) cpuset.CPUSet {
	tempSet := exclusiveCPUSet
	for _, coreID := range exclusiveCPUSet.List() {
		if siblings, exists := coreMap[coreID]; exists {
			siblingSet, err := cpuset.Parse(siblings)
			if err != nil {
				log.Println("ERROR: could not parse the HT siblings list of assigned exclusive cores because:" + err.Error())
				return exclusiveCPUSet
			}
			tempSet = tempSet.Union(siblingSet)
		}
	}
	return tempSet
}
