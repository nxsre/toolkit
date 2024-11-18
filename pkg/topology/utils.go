package topology

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/golang/glog"
	"github.com/jochenvg/go-udev"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

// ExecCommand is generic wrapper around cmd.Run. It executes the exec.Cmd arriving as an input parameters, and either returns an error, or the stdout of the command to the caller
// Used to interrogate CPU topology and cpusets directly from the host OS
func ExecCommand(cmd *exec.Cmd) (string, error) {
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return string(stdout.Bytes()), nil
}

func listAndParseCores(attribute string) map[int]int {
	glog.Infoln("listAndParseCores::", attribute)
	coreMap := make(map[int]int)
	outStr, err := ExecCommand(exec.Command("lscpu", "-p=cpu,"+attribute))
	if err != nil {
		log.Println("ERROR: could not interrogate the CPU topology of the node with lscpu, because:" + err.Error())
		return coreMap
	}
	//Here be dragons: we need to manually parse the stdout into a CPU core map line-by-line
	//lscpu -p and -J options are mutually exclusive :(
	for _, lsLine := range strings.Split(strings.TrimSuffix(outStr, "\n"), "\n") {
		cpuInfoStr := strings.Split(lsLine, ",")
		if len(cpuInfoStr) != 2 {
			continue
		}
		cpuInt, cpuErr := strconv.Atoi(cpuInfoStr[0])
		attributeInt, numaErr := strconv.Atoi(cpuInfoStr[1])
		if cpuErr != nil || numaErr != nil {
			continue
		}
		coreMap[cpuInt] = attributeInt
	}
	return coreMap
}

// 使用 udev 库从设备读取 numa 信息
func numa(path string) (int, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		return -1, err
	}
	var devType uint8
	var fm = fi.Mode()
	switch fm & (os.ModeType | os.ModeCharDevice) {
	case 0:
		fmt.Println("文件")
	case os.ModeDir:
		fmt.Println("目录")
	case os.ModeSymlink:
		fmt.Println("软连接")
	case os.ModeNamedPipe:
		fmt.Println("管道")
	case os.ModeSocket:
		fmt.Println("socket")
	case os.ModeDevice:
		fmt.Println("设备")
	case os.ModeDevice | os.ModeCharDevice:
		devType = os.ModeCharDevice.String()[0]
	default:
		return -1, errors.New("unknow device type")
	}

	if sys, ok := fi.Sys().(*syscall.Stat_t); ok {
		major, minor := devModesSplit(sys.Rdev)

		// 根据 major minor 查询设备信息（sysfs 路径等信息）
		u := udev.Udev{}
		d := u.NewDeviceFromDevnum(devType, udev.MkDev(int(major), int(minor)))
		if d == nil {
			return -1, errors.New("new device failed")
		}
		bs, err := os.ReadFile(strings.Replace(d.Syspath(), filepath.Join(d.Subsystem(), d.Sysname()), "numa_node", -1))
		if err != nil {
			return -1, err
		}
		return strconv.Atoi(TrimSpaceString(string(bs)))
	}
	return -1, errors.New("unknown device")
}

func devModesSplit(rdev uint64) (major int64, minor int64) {
	// Constants herein are not a joy: they're a workaround for https://github.com/golang/go/issues/8106
	return int64((rdev >> 8) & 0xff), int64((rdev & 0xff) | ((rdev >> 12) & 0xfff00))
}

var newlineBytes = []byte{'\n'}

// 去掉 src 开头和结尾的空白, 如果 src 包括换行, 去掉换行和这个换行符两边的空白
//
//	NOTE: 根据 '\n' 来分行的, 某些系统或软件用 '\r' 来分行, 则不能正常工作.
func TrimSpace(src []byte) []byte {
	bytesArr := bytes.Split(src, newlineBytes)
	for i := 0; i < len(bytesArr); i++ {
		bytesArr[i] = bytes.TrimSpace(bytesArr[i])
	}
	return bytes.Join(bytesArr, nil)
}

// 去掉 src 开头和结尾的空白, 如果 src 包括换行, 去掉换行和这个换行符两边的空白
//
//	NOTE: 根据 '\n' 来分行的, 某些系统或软件用 '\r' 来分行, 则不能正常工作.
func TrimSpaceString(src string) string {
	strs := strings.Split(src, string(newlineBytes))
	for i := 0; i < len(strs); i++ {
		strs[i] = strings.TrimSpace(strs[i])
	}
	return strings.Join(strs, "")
}

// ConvertWeight 转换字符串形式的权重描述到 map 结构，如输入为: "0:22,1:26,2:26,3:26"
func ConvertWeight(weight string) (map[int]int, error) {
	pattern := regexp.MustCompile("(([0-9]+):([0-9]+),?)")
	numaWeight := map[int]int{}
	if pattern.MatchString(weight) {
		for _, v := range pattern.FindAllStringSubmatch(weight, -1) {
			if len(v) == 4 {
				n2, err := strconv.Atoi(v[2])
				if err != nil {
					fmt.Println(v[2], "convert error:", err)
					return nil, err
				}
				n3, err := strconv.Atoi(v[3])
				if err != nil {
					fmt.Println(v[3], "convert error:", err)
					return nil, err
				}
				numaWeight[n2] = n3
			}
		}
	}
	return numaWeight, nil
}
