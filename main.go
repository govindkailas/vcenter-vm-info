package main

import (
	"context"
	"crypto/sha1"
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"os"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type DiskInfo struct {
	FileName   string
	CapacityGB float64
}

type VMInfo struct {
	UUID       string
	CPUCount   int32
	MemoryGB   float64
	BootOption string
	Disks      []DiskInfo
}

func getThumbprint(host string) (string, error) {
	conn, err := net.Dial("tcp", host+":443")
	if err != nil {
		return "", fmt.Errorf("failed to connect to %s: %v", host, err)
	}
	defer conn.Close()

	tlsConn := tls.Client(conn, &tls.Config{InsecureSkipVerify: true})
	if err := tlsConn.Handshake(); err != nil {
		return "", fmt.Errorf("TLS handshake failed: %v", err)
	}
	defer tlsConn.Close()

	cert := tlsConn.ConnectionState().PeerCertificates[0]
	thumbprint := sha1.Sum(cert.Raw)
	thumbprintStr := fmt.Sprintf("%X", thumbprint)
	// Format thumbprint with colons (e.g., A1:B2:C3:...)
	formatted := ""
	for i := 0; i < len(thumbprintStr); i += 2 {
		formatted += thumbprintStr[i : i+2]
		if i+2 < len(thumbprintStr) {
			formatted += ":"
		}
	}
	return formatted, nil
}

func getVMInfo(ctx context.Context, client *govmomi.Client, vmName string) (*VMInfo, error) {
	finder := find.NewFinder(client.Client, true)

	// Find the first datacenter and set it as the context for the finder
	dc, err := finder.DefaultDatacenter(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find default datacenter: %v", err)
	}
	finder.SetDatacenter(dc)

	// Search for the VM in the selected datacenter
	vm, err := finder.VirtualMachine(ctx, vmName)
	if err != nil {
		return nil, fmt.Errorf("failed to find VM %q: %v", vmName, err)
	}

	var moVM mo.VirtualMachine
	err = vm.Properties(ctx, vm.Reference(), []string{"config"}, &moVM)
	if err != nil {
		return nil, fmt.Errorf("failed to get VM properties: %v", err)
	}

	// Get CPU and Memory
	cpuCount := moVM.Config.Hardware.NumCPU
	memoryGB := float64(moVM.Config.Hardware.MemoryMB) / 1024

	// Get Boot Option
	bootOption := "BIOS"
	if moVM.Config.Firmware == string(types.GuestOsDescriptorFirmwareTypeEfi) {
		bootOption = "UEFI"
	}

	// Get Disk Info
	disks := []DiskInfo{}
	for _, device := range moVM.Config.Hardware.Device {
		if disk, ok := device.(*types.VirtualDisk); ok {
			if backing, ok := disk.Backing.(*types.VirtualDiskFlatVer2BackingInfo); ok {
				capacityGB := float64(disk.CapacityInKB) / 1024 / 1024
				disks = append(disks, DiskInfo{
					FileName:   backing.FileName,
					CapacityGB: capacityGB,
				})
			}
		}
	}

	return &VMInfo{
		UUID:       moVM.Config.Uuid,
		CPUCount:   cpuCount,
		MemoryGB:   memoryGB,
		BootOption: bootOption,
		Disks:      disks,
	}, nil
}

func main() {
	// Read environment variables
	vcenterHost := os.Getenv("VCENTER_HOST")
	username := os.Getenv("VCENTER_USERNAME")
	password := os.Getenv("VCENTER_PASSWORD")
	vmName := os.Getenv("VM_NAME")

	if vcenterHost == "" || username == "" || password == "" || vmName == "" {
		fmt.Println("Missing required environment variables: VCENTER_HOST, VCENTER_USERNAME, VCENTER_PASSWORD, VM_NAME")
		os.Exit(1)
	}

	// Connect to vCenter
	ctx := context.Background()
	u := &url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf("%s:443", vcenterHost),
		Path:   "/sdk",
		User:   url.UserPassword(username, password),
	}
	fmt.Printf("Connecting to vCenter with URL: %s\n", u.String())
	client, err := govmomi.NewClient(ctx, u, true)
	if err != nil {
		fmt.Printf("Failed to connect to vCenter Server: %v\n", err)
		os.Exit(1)
	}
	defer client.Logout(ctx)

	// Get VM info
	vmInfo, err := getVMInfo(ctx, client, vmName)
	if err != nil {
		fmt.Printf("Error retrieving VM info: %v\n", err)
		os.Exit(1)
	}

	if vmInfo != nil {
		fmt.Printf("uuid: %q\n", vmInfo.UUID)
		fmt.Printf("CPU count: %d\n", vmInfo.CPUCount)
		fmt.Printf("Memory: %.2f GB\n", vmInfo.MemoryGB)
		fmt.Printf("Boot Option: %s\n", vmInfo.BootOption)
		for i, disk := range vmInfo.Disks {
			fmt.Printf("Disk %d:\n", i+1)
			fmt.Printf("  Backing File: %q\n", disk.FileName)
			fmt.Printf("  Capacity: %.2f GB\n", disk.CapacityGB)
		}
	} else {
		fmt.Printf("VM %q not found\n", vmName)
	}

	// Get thumbprint
	thumbprint, err := getThumbprint(vcenterHost)
	if err != nil {
		fmt.Printf("Failed to retrieve thumbprint: %v\n", err)
	} else {
		fmt.Printf("thumbprint: %q\n", thumbprint)
	}

	// Print URL
	fmt.Printf("url: %q\n", fmt.Sprintf("https://%s", vcenterHost))
}
