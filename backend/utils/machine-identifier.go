package utils

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/jaypipes/ghw"
	"golang.org/x/sys/windows/registry"
)

type MachineInfo struct {
	Hostname   string         `json:"Hostname"`
	OS         string         `json:"OS"`
	Version    string         `json:"Version"`
	HWID       string         `json:"HWID"`
	ExternalIP string         `json:"ExternalIP"`
	CPU        ghw.CPUInfo    `json:"CPU"`
	RAM        ghw.MemoryInfo `json:"RAM"`
	GPU        ghw.GPUInfo    `json:"GPU"`
}

type ExtendedMachineInfo struct {
	MachineInfo
	InternalIP string     `json:"InternalIP"`
	ADDomain   string     `json:"ADDomain"`
	EMLyConfig EMLyConfig `json:"EMLyConfig"`
}

func GetMachineInfo() (*MachineInfo, error) {
	info := &MachineInfo{}

	// 1. Get Hostname
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}
	info.Hostname = hostname

	// 2. Get OS Info
	info.OS = fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)

	// 3. Get Version Info
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

	// 3. Get HWID (Windows specific via wmic)
	// Fallback or different implementation needed for Linux/Mac if required
	if runtime.GOOS == "windows" {
		out, err := exec.Command("wmic", "csproduct", "get", "uuid").Output()
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

		// Fallback to registry MachineGuid if wmic fails or empty
		if info.HWID == "" {
			// Simplified registry read attempt using reg query command to avoid cgo/syscall complexity for now
			// HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Cryptography -> MachineGuid
			out, err := exec.Command("reg", "query", `HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Cryptography`, "/v", "MachineGuid").Output()
			if err == nil {
				// Parse output
				content := string(out)
				if idx := strings.Index(content, "REG_SZ"); idx != -1 {
					info.HWID = strings.TrimSpace(content[idx+6:])
				}
			}
		}
	} else {
		info.HWID = "Not implemented for " + runtime.GOOS
	}

	// 4. Get External IP
	ip, err := getExternalIP()
	if err == nil {
		info.ExternalIP = ip
	} else {
		info.ExternalIP = "Unavailable"
	}

	// 5. Get CPU Info
	cpuInfo, err := getCPUInfo()
	if err == nil {
		info.CPU = *cpuInfo
	}

	// 6. Get GPU Info
	gpuInfo, err := getGPUInfo()
	if err == nil {
		info.GPU = *gpuInfo
	}

	// 7. Get RAM Info
	ramInfo, err := getRAMInfo()
	if err == nil {
		info.RAM = *ramInfo
	}

	return info, nil
}

func getExternalIP() (string, error) {
	resp, err := http.Get("https://api.ipify.org?format=text")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func getCPUInfo() (*ghw.CPUInfo, error) {
	cpuInfo, _ := ghw.CPU()
	if cpuInfo == nil {
		return nil, fmt.Errorf("failed to get CPU info")
	}
	return cpuInfo, nil
}

func getGPUInfo() (*ghw.GPUInfo, error) {
	gpuInfo, err := ghw.GPU()
	if gpuInfo == nil {
		return nil, fmt.Errorf("failed to get GPU info: %w", err)
	}
	return gpuInfo, nil
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
			return ip.String(), nil
		}
	}
	return "", fmt.Errorf("no internal IP found")
}

func getADDomain() (string, error) {
	// This is a Windows-specific implementation to get the Active Directory domain
	if runtime.GOOS != "windows" {
		return "", fmt.Errorf("AD domain retrieval not implemented for %s", runtime.GOOS)
	}
	out, err := exec.Command("powershell", "-Command", "(Get-WmiObject Win32_ComputerSystem).Domain").Output()
	if err != nil {
		return "", fmt.Errorf("failed to get AD domain: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func GetExtendedMachineInfo() (*ExtendedMachineInfo, error) {
	baseInfo, err := GetMachineInfo()
	if err != nil {
		return nil, err
	}
	internalIP, err := getInternalIP()
	if err != nil {
		internalIP = "Unavailable"
	}
	adDomain, err := getADDomain()
	if err != nil {
		adDomain = "Unavailable"
	}
	cfgPath := DefaultConfigPath()
	rawConfig, err := LoadConfig(cfgPath)
	var emlyConfig EMLyConfig
	if err == nil {
		emlyConfig = rawConfig.EMLy
	} else {
		emlyConfig = EMLyConfig{} // Return empty config if there's an error
	}
	return &ExtendedMachineInfo{
		MachineInfo: *baseInfo,
		InternalIP:  internalIP,
		ADDomain:    adDomain,
		EMLyConfig:  emlyConfig,
	}, nil
}
