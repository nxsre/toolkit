package topology

import (
	"fmt"
	"github.com/jaypipes/pcidb"
	"github.com/smallnest/weighted"
	"github.com/u-root/u-root/pkg/pci"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

//func listGPU(cpu map[int]Topology) map[int]Topology {
//	var gpu = make(map[int]Topology)
//	driPath := "/dev/dri"
//	entries, _ := os.ReadDir(driPath)
//	gpuPattern := regexp.MustCompile("^renderD([0-9]+)$")
//	for _, entry := range entries {
//		if gpuPattern.MatchString(entry.Name()) {
//			gpuPath := filepath.Join(driPath, entry.Name())
//			if node, err := numa(gpuPath); err == nil {
//				var socket int
//				var core int
//				for _, v := range cpu {
//					if node == v.Node {
//						socket = v.Socket
//						core = v.Core
//						break
//					}
//				}
//				gpu[node] = Topology{
//					CPU:    core,
//					Node:   node,
//					GPU:    gpuPath,
//					Socket: socket,
//				}
//
//			}
//		}
//	}
//	return gpu
//}

// netint 编码卡不考虑 numa 绑定问题
//func listNvme() []string {
//	nvmes := []string{}
//	devPath := "/dev"
//	entries, _ := os.ReadDir(devPath)
//	gpuPattern := regexp.MustCompile("^nvme([0-9]+)$")
//	for _, entry := range entries {
//		if gpuPattern.MatchString(entry.Name()) {
//			nvmes = append(nvmes, filepath.Join(devPath, entry.Name()))
//		}
//	}
//	return nvmes
//}

type DeviceType string

const (
	Netint     DeviceType = "netint"
	Vastaitech            = "vastaitech"
	AMDGPU                = "amdgpu"
)

func NewDevsPoller(numaws map[int]int) (map[int]Topology, weighted.W, map[int]*weighted.SW) {
	cpuTopol := GetCGPUTopology()

	cpuTopolByNuma := map[int][]Topology{}
	for _, cpu := range cpuTopol {
		cpuTopolByNuma[cpu.Node] = append(cpuTopolByNuma[cpu.Node], cpu)
	}

	cpuNumaSW := &weighted.RRW{}
	for node, cpu := range cpuTopolByNuma {
		if numaws != nil {
			if w, ok := numaws[node]; ok {
				cpuNumaSW.Add(cpu, w)
				continue
			}
		}
		cpuNumaSW.Add(cpu, 1)
	}

	// 设备节点设置轮询策略
	devNumaSW := map[int]*weighted.SW{}

	if ExistsExtDevices(Vastaitech) {
		numas := map[int][]Topology{}
		for _, n := range GetNodeTopology() {
			numas[n] = []Topology{}
		}

		all := GetAllExtDevices()[Vastaitech]
		for _, devs := range all {
			for _, dev := range devs {
				numas[dev.Node] = append(numas[dev.Node], dev)
			}
		}

		// numa数为4 时，需要重新计算同socket下不同 numa 节点的负载均衡系数，numa 节点数为 2 时不需要考虑同 socket 不同 numa 的均衡问题
		if len(numas) == 4 {
			// socket 0
			ought0 := (len(numas[0]) + len(numas[1])) / 2
			if len(numas[0]) > ought0 && len(numas[1]) < ought0 {
				for _, dev := range numas[0][ought0:] {
					numas[1] = append(numas[1], dev)
					numas[0] = deleteDev(numas[0], dev)
				}
			}

			// socket 1
			ought1 := (len(numas[2]) + len(numas[3])) / 2
			if len(numas[2]) > ought0 && len(numas[3]) < ought1 {
				for _, dev := range numas[2][ought1:] {
					numas[3] = append(numas[3], dev)
					numas[2] = deleteDev(numas[2], dev)
				}
			}
		}

		for node, devs := range numas {
			devNumaSW[node] = &weighted.SW{}
			for _, dev := range devs {
				devNumaSW[node].Add(&dev, 1)
			}
		}
	}
	return cpuTopol, cpuNumaSW, devNumaSW
}

// deleteDev 用于从 numa 节点下删除多余的设备，迁移到其他 numa 节点
func deleteDev(s []Topology, elem Topology) []Topology {
	j := 0
	for _, v := range s {
		if v.Devices[0] != elem.Devices[0] {
			s[j] = v
			j++
		}
	}
	return s[:j]
}

func GetExtDevices(cpu map[int]Topology) map[DeviceType]map[int][]Topology {
	extDevices := map[DeviceType]map[int][]Topology{
		Netint:     {},
		Vastaitech: {},
		AMDGPU:     {},
	}
	// 遍历 pci 设备信息，用于识别编码卡
	// pcidb 升级数据库：
	// wget https://pci-ids.ucw.cz/v2.2/pci.ids
	// mv pci.ids /usr/share/hwdata/pci.ids
	// 初始化 pcidb
	db, err := pcidb.New()
	if err != nil {
		fmt.Printf("Error getting PCI info: %v", err)
	}

	// 遍历方法
	br, err := pci.NewBusReader()
	if err != nil {
		fmt.Println("读取 pcie", err.Error())
		return nil
	}

	var devices []*pci.PCI
	devices, err = br.Read(func(p *pci.PCI) bool {
		filters := map[string]string{
			// netint 编码卡标识: VendorID: "1d82" DeviceID: "0401"
			"1d82": "0401",
			// 瀚博编码卡标识 VendorID: "1f4f" DeviceID: "0200"
			"1f4f": "0200",
			// AMDGPU 标识 VendorID: "1002" DeviceID: "73a3"
			"1002": "73a3",
		}

		for k, v := range filters {
			vendorID, err := strconv.ParseUint(k, 16, 16)
			if err != nil {
				return false
			}
			deviceID, err := strconv.ParseUint(v, 16, 16)
			if err != nil {
				return false
			}

			if p.Vendor == uint16(vendorID) && p.Device == uint16(deviceID) {
				return true
			}
		}
		return false
	})
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	if len(devices) == 0 {
		fmt.Println("got 0 devices, want at least 1")
		return nil
	}

	for _, device := range devices {
		// pcidb 需要根据 16 进制的 vendor 查询
		hexVendorID := strconv.FormatUint(uint64(device.Vendor), 16)
		hexDeviceID := fmt.Sprintf("%04x", device.Device)

		// netint 编码卡
		if hexVendorID == "1d82" && hexDeviceID == "0401" {
			fmt.Println("netint")
			path, _ := filepath.EvalSymlinks(fmt.Sprintf("/dev/dri/by-path/pci-%s-render", device.Addr))
			fmt.Println("numa_node:", getNumaNode(device.FullPath), "renderDevice:", path)
			items, _ := os.ReadDir(filepath.Join(device.FullPath, "nvme"))
			numaDev := items[0]
			masterDev := filepath.Join("/dev", numaDev.Name())
			fmt.Println("numa_node:", getNumaNode(filepath.Join(device.FullPath, "nvme", numaDev.Name())), "devices:", masterDev+"n1")
			/* TODO: 容器内读不到 nvme ，暂时忽略
			// 读取 nvme 设备信息（型号、固件版本号等）
			netint := nvme.NewNVMeDevice(masterDev)
			err := netint.Open()
			if err != nil {
				log.Fatalf("open %s failed: %v", masterDev, err)
			}
			defer netint.Close()

			// IdentifyController 会写到 io.Writer 同时返回 NVMeController
			c, err := netint.IdentifyController(io.Discard)
			if err != nil {
				fmt.Println(err)
				return nil
			}
			fmt.Println("VendorID:", strconv.FormatInt(int64(c.VendorID), 16), "SN:", c.SerialNumber, "FirmwareVersion:", c.FirmwareVersion)
			*/
			// netint 编码卡是虚拟 nvme 设备，没有 SMART 信息
			//if err := netint.PrintSMART(os.Stdout); err != nil {
			//	panic(err)
			//}

			// 获取设备信息
			socket, node := getSocketNode(cpu, device.FullPath)

			extDevices[Netint][node] = append(extDevices[Netint][node], Topology{
				Core:    -1,
				CPU:     -1,
				Node:    node,
				Socket:  socket,
				Devices: []string{masterDev, masterDev + "n1"},
			})
		}

		// 瀚博 GPU
		if hexVendorID == "1f4f" && hexDeviceID == "0200" {
			//fmt.Println("瀚博")
			var id string
			pattern := regexp.MustCompile(`va_card_hw.(\d+)`)
			vaDeviceTmpls := []string{"/dev/va%s_ctl", "/dev/va_video%s", "/dev/vacc%s"}
			vaDevices := []string{}

			items, err := os.ReadDir(device.FullPath)
			if err != nil {
				log.Println(err)
			} else {
				for _, item := range items {
					if pattern.MatchString(item.Name()) {
						id = pattern.FindAllStringSubmatch(item.Name(), 1)[0][1]
						path, err := filepath.EvalSymlinks(fmt.Sprintf("/dev/dri/by-path/pci-%s-platform-%s-render", device.Addr, item.Name()))
						if err == nil {
							vaDevices = append(vaDevices, path)
						}
						break
					}
				}
			}
			if id != "" {
				for _, v := range vaDeviceTmpls {
					vaDevices = append(vaDevices, fmt.Sprintf(v, id))
				}
			}

			vaDevices = append(vaDevices, "/dev/va_sync", "/dev/vatools")

			// 获取设备信息
			socket, node := getSocketNode(cpu, device.FullPath)

			extDevices[Vastaitech][node] = append(extDevices[Vastaitech][node], Topology{
				Core:    -1,
				CPU:     -1,
				Node:    node,
				Socket:  socket,
				Devices: vaDevices,
			})
			fmt.Println("numaNode", getNumaNode(device.FullPath), "vaDevices:", vaDevices)
		}

		// AMD GPU
		if hexVendorID == "1002" && hexDeviceID == "73a3" {
			//fmt.Println("amdgpu")
			renderDev := fmt.Sprintf("/dev/dri/by-path/pci-%s-render", device.Addr)
			path, err := filepath.EvalSymlinks(renderDev)
			if err != nil {
				fmt.Println("renderDevice Error:", renderDev, err)
			}
			fmt.Println("numa_node:", getNumaNode(device.FullPath), "renderDevice:", path)
			// 获取设备信息
			socket, node := getSocketNode(cpu, device.FullPath)

			extDevices[AMDGPU][node] = append(extDevices[AMDGPU][node], Topology{
				Core:    -1,
				CPU:     -1,
				Node:    node,
				Socket:  socket,
				Devices: []string{path},
			})
		}
		fmt.Printf("PCI Device: %s\n", device.DeviceName)

		// 瀚博没有进 pci-sig 数据库
		if v, ok := db.Vendors[hexVendorID]; ok {
			fmt.Printf("\tPCI Vendor: %s\n", v.Name)
			for _, p := range v.Products {
				if p.ID == device.DeviceName && p.VendorID == hexVendorID {
					fmt.Printf("\tDevice %s ProductName: %s\n", device.DeviceName, p.Name)
				}
			}
		}
	}

	return extDevices
}

func getNumaNode(devicePath string) int {
	numa, err := os.ReadFile(filepath.Join(devicePath, "numa_node"))
	if err != nil {
		fmt.Println(err)
	}

	numa_node, _ := strconv.ParseUint(string(TrimSpace(numa)), 10, 10)
	return int(numa_node)
}

// 返回 numa node
func getSocketNode(cpu map[int]Topology, devPath string) (int, int) {
	for _, v := range cpu {
		if getNumaNode(devPath) == v.Node {
			return v.Socket, v.Node
		}
	}
	return -1, -1
}
