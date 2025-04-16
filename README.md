# vcenter-vm-info

A command-line tool to fetch and display information about VM from a VMware vCenter server. This is particularly useful when we are planning to migrate VM from vCenter to `KubeVirt` as [explained here](https://github.com/govindkailas/kubevirt-examples/blob/main/vm/demo-vm-from-vcenter.yaml) 

## Features

- Connects to a vCenter server using environment variables for credentials.
- Fetches and displays detailed information about a specified virtual machine, including:
    - UUID
    - CPU count
    - Memory size
    - Boot option
    - Disk details (backing file, capacity)
    - vCenter thumbprint
    - vCenter URL

## Usage

Set the required environment variables and run the program:

```sh
export VCENTER_HOST=<vcenter-server>
export VCENTER_USERNAME=<username>
export VCENTER_PASSWORD=<password>
export VM_NAME=<virtual-machine-name>
go run main.go
```
Or you can use the binary from the go build output `./vcenter-vm-info`

## Output Format

Example output:

```
uuid: "564dcfb1-2cc4-9636-4c55-b110d4a81846"
CPU count: 1
Memory: 2.00 GB
Boot Option: BIOS
Disk 1:
    Backing File: "[vm-datastore] Win10-2/Win10-2.vmdk"
    Capacity: 127.00 GB
thumbprint: "70:7D:02:A1:11:F3:F8:K9:BB:38:72:46:57:F4:1B:6F:3B:41:3C:FE"
url: "https://10.10.10.10"
```

## Requirements

- Go 1.18 or later
- Network access to the vCenter server

