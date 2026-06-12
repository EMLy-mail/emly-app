package utils

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"emly/backend/logger"

	"github.com/denisbrodbeck/machineid"
	"github.com/jaypipes/ghw"
	"golang.org/x/sys/windows/registry"
)

type MachineInfo struct {
	Hostname string         `json:"Hostname"`
	OS       string         `json:"OS"`
	Version  string         `json:"Version"`
	HWID     string         `json:"HWID"`
	CPU      ghw.CPUInfo    `json:"CPU"`
	RAM      ghw.MemoryInfo `json:"RAM"`
}

type ExtendedMachineInfo struct {
	MachineInfo
	InternalIP string     `json:"InternalIP"`
	ADDomain   string     `json:"ADDomain"`
	EMLyConfig EMLyConfig `json:"EMLyConfig"`
}

func GetMachineInfo() (*MachineInfo, error) {
	logger.Debug("GetMachineInfo: starting metadata collection")
	start := time.Now()

	info := &MachineInfo{}

	// 1. Get Hostname
	logger.Debug("GetMachineInfo: fetching hostname")
	t1 := time.Now()
	hostname, err := os.Hostname()
	logger.Debug("GetMachineInfo: fetched hostname", "duration_ms", time.Since(t1).Milliseconds())
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}
	info.Hostname = hostname

	// 2. Get OS Info
	logger.Debug("GetMachineInfo: fetching OS info")
	t2 := time.Now()
	info.OS = fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)
	logger.Debug("GetMachineInfo: fetched OS info", "duration_ms", time.Since(t2).Milliseconds())

	// 3. Get Version Info
	logger.Debug("GetMachineInfo: fetching Windows version info")
	t3 := time.Now()
	k, _ := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows NT\CurrentVersion`,
		registry.QUERY_VALUE,
	)
	defer k.Close()

	product, _, _ := k.GetStringValue("ProductName")
	build, _, _ := k.GetStringValue("CurrentBuild")
	ubr, _, _ := k.GetIntegerValue("UBR")
	display, _, _ := k.GetStringValue("DisplayVersion")
	edition, _, _ := k.GetStringValue("EditionID")

	// Append edition if available
	if edition != "" {
		product = fmt.Sprintf("%s %s", product, edition)
	}

	// Split display versione via H (like 23H2, 24H2, 25H2), if its => 23, then its Windows 11, not 10
	if strings.HasPrefix(display, "2") {
		parts := strings.SplitN(display, "H", 2)
		if len(parts) > 0 {
			yearPart := parts[0]
			if yearPartInt := strings.TrimSpace(yearPart); yearPartInt >= "23" {
				product = "Windows 11"
			}
		}
	}

	info.Version = fmt.Sprintf("%s %s %s (Build %s.%d)", product, display, edition, build, ubr)
	logger.Debug("GetMachineInfo: fetched Windows version info", "duration_ms", time.Since(t3).Milliseconds())

	// 3. Get HWID (Windows specific via wmic)
	logger.Debug("GetMachineInfo: fetching HWID")
	t4 := time.Now()
	// Fallback or different implementation needed for Linux/Mac if required
	if runtime.GOOS == "windows" {
		wmicCmd := exec.Command("wmic", "csproduct", "get", "uuid")
		wmicCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
		out, err := wmicCmd.Output()
		if err == nil {
			// Parse output which looks like "UUID \n <UUID> \n\n"
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if trimmed != "" && trimmed != "UUID" {
					info.HWID = trimmed
					break
				}
			}
		}

		// Fallback to machineid library if wmic fails or returns empty, as it has multiple strategies for Windows
		if info.HWID == "" {
			logger.Debug("WMIC failed or was empty, fallback to machineid.ProtectedID")
			out, err := machineid.ProtectedID("emly-machine-id")
			if err != nil {
				logger.Warn("GetMachineInfo: machineid.ProtectedID failed", "error", err)
				info.HWID = "N/A for " + runtime.GOOS
				return info, nil
			}
			info.HWID = out
		}
	} else {
		id, err := machineid.ProtectedID("emly-machine-id")
		if err != nil {
			logger.Warn("GetMachineInfo: machineid.ProtectedID failed", "error", err)
			info.HWID = "N/A for " + runtime.GOOS
		}
		info.HWID = id
	}
	info.HWID = strings.TrimSpace(info.HWID)

	logger.Debug("GetMachineInfo: fetched HWID", "duration_ms", time.Since(t4).Milliseconds())

	// 5. Get CPU Info
	logger.Debug("GetMachineInfo: fetching CPU info")
	t6 := time.Now()
	cpuInfo, err := getCPUInfo()
	elapsed6 := time.Since(t6).Milliseconds()
	if err == nil {
		logger.Debug("GetMachineInfo: fetched CPU info", "duration_ms", elapsed6)
		info.CPU = *cpuInfo
	} else {
		logger.Debug("GetMachineInfo: CPU info failed", "error", err, "duration_ms", elapsed6)
	}

	// 7. Get RAM Info
	logger.Debug("GetMachineInfo: fetching RAM info")
	t8 := time.Now()
	ramInfo, err := getRAMInfo()
	elapsed8 := time.Since(t8).Milliseconds()
	if err == nil {
		logger.Debug("GetMachineInfo: fetched RAM info", "duration_ms", elapsed8)
		info.RAM = *ramInfo
	} else {
		logger.Debug("GetMachineInfo: RAM info failed", "error", err, "duration_ms", elapsed8)
	}

	logger.Debug("GetMachineInfo: completed", "duration_ms", time.Since(start).Milliseconds())
	return info, nil
}

func getCPUInfo() (*ghw.CPUInfo, error) {
	cpuInfo, _ := ghw.CPU()
	if cpuInfo == nil {
		return nil, fmt.Errorf("failed to get CPU info")
	}
	return cpuInfo, nil
}

func getRAMInfo() (*ghw.MemoryInfo, error) {
	memory, err := ghw.Memory()
	if memory == nil {
		return nil, fmt.Errorf("failed to get RAM info: %w", err)
	}
	return memory, nil
}

func getInternalIP() (string, error) {
	// This is a simplified method to get the internal IP by checking the network interfaces
	// For a more robust solution, consider using a library like "github.com/vishvananda/netlink"
	ifaces, err := net.Interfaces()
	if err != nil {
		logger.Error("getInternalIP: failed to list interfaces", "error", err)
		return "", fmt.Errorf("failed to get network interfaces: %w", err)
	}
	for _, iface := range ifaces {
		// ignore down or loopback interfaces
		if (iface.Flags&net.FlagUp) == 0 || (iface.Flags&net.FlagLoopback) != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			var ip net.IP
			switch v := a.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue
			}
			logger.Debug("getInternalIP: found valid IP", "interface", iface.Name, "ip", ip.String())
			return ip.String(), nil
		}
	}
	return "", fmt.Errorf("no internal IP found")
}

func getADDomain() (string, error) {
	// This is a Windows-specific implementation to get the Active Directory domain
	if runtime.GOOS != "windows" {
		logger.Debug("getADDomain: skipped on non-windows OS")
		return "", fmt.Errorf("AD domain retrieval not implemented for %s", runtime.GOOS)
	}
	logger.Debug("getADDomain: querying Win32_ComputerSystem")
	psCmd := exec.Command("powershell", "-Command", "(Get-WmiObject Win32_ComputerSystem).Domain")
	psCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	out, err := psCmd.Output()
	if err != nil {
		logger.Error("getADDomain: powershell query failed", "error", err)
		return "", fmt.Errorf("failed to get AD domain: %w", err)
	}
	domain := strings.TrimSpace(string(out))
	logger.Debug("getADDomain: success", "domain", domain)
	return domain, nil
}

func GetExtendedMachineInfo() (*ExtendedMachineInfo, error) {
	logger.Debug("GetExtendedMachineInfo: starting extended info collection")
	start := time.Now()

	t1 := time.Now()
	baseInfo, err := GetMachineInfo()
	elapsed1 := time.Since(t1).Milliseconds()
	if err != nil {
		logger.Error("GetExtendedMachineInfo: base info failed", "error", err, "duration_ms", elapsed1)
		return nil, err
	}
	logger.Debug("GetExtendedMachineInfo: fetched base info", "duration_ms", elapsed1)

	logger.Debug("GetExtendedMachineInfo: fetching internal IP")
	t2 := time.Now()
	internalIP, err := getInternalIP()
	elapsed2 := time.Since(t2).Milliseconds()
	if err != nil {
		logger.Debug("GetExtendedMachineInfo: internal IP failed", "error", err, "duration_ms", elapsed2)
		internalIP = "Unavailable"
	} else {
		logger.Debug("GetExtendedMachineInfo: fetched internal IP", "duration_ms", elapsed2)
	}

	logger.Debug("GetExtendedMachineInfo: fetching AD domain")
	t3 := time.Now()
	adDomain, err := getADDomain()
	elapsed3 := time.Since(t3).Milliseconds()
	if err != nil {
		logger.Debug("GetExtendedMachineInfo: AD domain failed", "error", err, "duration_ms", elapsed3)
		adDomain = "Unavailable"
	} else {
		logger.Debug("GetExtendedMachineInfo: fetched AD domain", "duration_ms", elapsed3)
	}

	logger.Debug("GetExtendedMachineInfo: loading config")
	t4 := time.Now()
	cfgPath := DefaultConfigPath()
	rawConfig, err := LoadConfig(cfgPath)
	var emlyConfig EMLyConfig
	if err == nil {
		emlyConfig = rawConfig.EMLy
	} else {
		logger.Warn("GetExtendedMachineInfo: config loading failed, using defaults", "error", err)
		emlyConfig = EMLyConfig{} // Return empty config if there's an error
	}
	logger.Debug("GetExtendedMachineInfo: loaded config", "duration_ms", time.Since(t4).Milliseconds())

	logger.Debug("GetExtendedMachineInfo: completed", "duration_ms", time.Since(start).Milliseconds())
	return &ExtendedMachineInfo{
		MachineInfo: *baseInfo,
		InternalIP:  internalIP,
		ADDomain:    adDomain,
		EMLyConfig:  emlyConfig,
	}, nil
}
