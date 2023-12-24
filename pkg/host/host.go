/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package host

import (
	"fmt"
	"os"
	pathlib "path"
	"path/filepath"
	"strings"

	"github.com/coreos/go-systemd/v22/unit"
	"github.com/jaypipes/ghw"
	"github.com/vishvananda/netlink"
	"sigs.k8s.io/controller-runtime/pkg/log"

	sriovnetworkv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/consts"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/utils"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/vars"
)

const (
	hostPathFromDaemon    = consts.Host
	redhatReleaseFile     = "/etc/redhat-release"
	rhelRDMAConditionFile = "/usr/libexec/rdma-init-kernel"
	rhelRDMAServiceName   = "rdma"
	rhelPackageManager    = "yum"

	ubuntuRDMAConditionFile = "/usr/sbin/rdma-ndd"
	ubuntuRDMAServiceName   = "rdma-ndd"
	ubuntuPackageManager    = "apt-get"

	genericOSReleaseFile = "/etc/os-release"
)

// Contains all the host manipulation functions
//
//go:generate ../../bin/mockgen -destination mock/mock_host.go -source host.go
type HostManagerInterface interface {
	// TryEnableTun load the tun kernel module
	TryEnableTun()
	// TryEnableVhostNet load the vhost-net kernel module
	TryEnableVhostNet()
	// TryEnableRdma tries to enable RDMA on the machine base on the operating system
	// if the package doesn't exist it will also will try to install it
	// supported operating systems are RHEL RHCOS and ubuntu
	TryEnableRdma() (bool, error)
	// TryToGetVirtualInterfaceName tries to find the virtio interface name base on pci address
	// used for virtual environment where we pass SR-IOV virtual function into the system
	// supported platform openstack
	TryToGetVirtualInterfaceName(string) string
	// TryGetInterfaceName tries to find the SR-IOV virtual interface name base on pci address
	TryGetInterfaceName(string) string
	// GetNicSriovMode returns the interface mode
	// supported modes SR-IOV legacy and switchdev
	GetNicSriovMode(string) (string, error)
	// GetPhysSwitchID returns the physical switch ID for a specific pci address
	GetPhysSwitchID(string) (string, error)
	// GetPhysPortName returns the physical port name for a specific pci address
	GetPhysPortName(string) (string, error)
	// IsSwitchdev returns true of the pci address is on switchdev mode
	IsSwitchdev(string) bool
	// IsKernelLockdownMode returns true if the kernel is in lockdown mode
	IsKernelLockdownMode() bool
	// GetNetdevMTU returns the interface MTU for devices attached to kernel drivers
	GetNetdevMTU(string) int
	// SetNetdevMTU sets the MTU for a request interface
	SetNetdevMTU(string, int) error
	// SetSriovNumVfs changes the number of virtual functions allocated for a specific
	// physical function base on pci address
	SetSriovNumVfs(string, int) error
	// GetNetDevMac returns the network interface mac address
	GetNetDevMac(string) string
	// GetNetDevLinkSpeed returns the network interface link speed
	GetNetDevLinkSpeed(string) string

	// GetVfInfo returns the virtual function information is the operator struct from the host information
	GetVfInfo(string, []*ghw.PCIDevice) sriovnetworkv1.VirtualFunction
	// SetVfGUID sets the GUID for a virtual function
	SetVfGUID(string, netlink.Link) error
	// VFIsReady returns the interface virtual function if the device is ready
	VFIsReady(string) (netlink.Link, error)
	// SetVfAdminMac sets the virtual function administrative mac address via the physical function
	SetVfAdminMac(string, netlink.Link, netlink.Link) error

	// GetLinkType return the link type
	// supported types are ethernet and infiniband
	GetLinkType(sriovnetworkv1.InterfaceExt) string
	// ResetSriovDevice resets the number of virtual function for the specific physical function to zero
	ResetSriovDevice(sriovnetworkv1.InterfaceExt) error
	// DiscoverSriovDevices returns a list of all the available SR-IOV capable network interfaces on the system
	DiscoverSriovDevices(StoreManagerInterface) ([]sriovnetworkv1.InterfaceExt, error)
	// ConfigSriovDevice configure the request SR-IOV device with the desired configuration
	ConfigSriovDevice(iface *sriovnetworkv1.Interface, ifaceStatus *sriovnetworkv1.InterfaceExt) error
	// ConfigSriovInterfaces configure multiple SR-IOV devices with the desired configuration
	ConfigSriovInterfaces(StoreManagerInterface, []sriovnetworkv1.Interface, []sriovnetworkv1.InterfaceExt, map[string]bool) error
	// ConfigSriovInterfaces configure virtual functions for virtual environments with the desired configuration
	ConfigSriovDeviceVirtual(iface *sriovnetworkv1.Interface) error

	// Unbind unbinds a virtual function from is current driver
	Unbind(string) error
	// BindDpdkDriver binds the virtual function to a DPDK driver
	BindDpdkDriver(string, string) error
	// BindDefaultDriver binds the virtual function to is default driver
	BindDefaultDriver(string) error
	// HasDriver returns try if the virtual function is bind to a driver
	HasDriver(string) (bool, string)
	// RebindVfToDefaultDriver rebinds the virtual function to is default driver
	RebindVfToDefaultDriver(string) error
	// UnbindDriverIfNeeded unbinds the virtual function from a driver if needed
	UnbindDriverIfNeeded(string, bool) error

	// WriteSwitchdevConfFile writes the needed switchdev configuration files for HW offload support
	WriteSwitchdevConfFile(*sriovnetworkv1.SriovNetworkNodeState, map[string]bool) (bool, error)
	// PrepareNMUdevRule creates the needed udev rules to disable NetworkManager from
	// our managed SR-IOV virtual functions
	PrepareNMUdevRule([]string) error
	// AddUdevRule adds a specific udev rule to the system
	AddUdevRule(string) error
	// RemoveUdevRule removes a udev rule from the system
	RemoveUdevRule(string) error

	// GetCurrentKernelArgs reads the /proc/cmdline to check the current kernel arguments
	GetCurrentKernelArgs() (string, error)
	// IsKernelArgsSet check is the requested kernel arguments are set
	IsKernelArgsSet(string, string) bool

	// IsServiceExist checks if the requested systemd service exist on the system
	IsServiceExist(string) (bool, error)
	// IsServiceEnabled checks if the requested systemd service is enabled on the system
	IsServiceEnabled(string) (bool, error)
	// ReadService reads a systemd servers and return it as a struct
	ReadService(string) (*Service, error)
	// EnableService enables a systemd server on the host
	EnableService(service *Service) error
	// ReadServiceManifestFile reads the systemd manifest for a specific service
	ReadServiceManifestFile(path string) (*Service, error)
	// RemoveFromService removes a systemd service from the host
	RemoveFromService(service *Service, options ...*unit.UnitOption) (*Service, error)
	// ReadScriptManifestFile reads the script manifest from a systemd service
	ReadScriptManifestFile(path string) (*ScriptManifestFile, error)
	// ReadServiceInjectionManifestFile reads the injection manifest file for the systemd service
	ReadServiceInjectionManifestFile(path string) (*Service, error)
	// CompareServices compare two servers and return true if they are equal
	CompareServices(serviceA, serviceB *Service) (bool, error)

	// private functions
	// part of the interface for the mock generation
	// LoadKernelModule loads a kernel module to the host
	LoadKernelModule(name string, args ...string) error
	// IsKernelModuleLoaded returns try if the requested kernel module is loaded
	IsKernelModuleLoaded(string) (bool, error)
	// IsRHELSystem returns try if the system is a RHEL base
	IsRHELSystem() (bool, error)
	// IsUbuntuSystem returns try if the system is an ubuntu base
	IsUbuntuSystem() (bool, error)
	// IsCoreOS returns true if the system is a CoreOS or RHCOS base
	IsCoreOS() (bool, error)
	// RdmaIsLoaded returns try if RDMA kernel modules are loaded
	RdmaIsLoaded() (bool, error)
	// EnableRDMA enable RDMA on the system
	EnableRDMA(string, string, string) (bool, error)
	// InstallRDMA install RDMA packages on the system
	InstallRDMA(string) error
	// TriggerUdevEvent triggers a udev event
	TriggerUdevEvent() error
	// ReloadDriver reloads a requested driver
	ReloadDriver(string) error
	// EnableRDMAOnRHELMachine enable RDMA on a RHEL base system
	EnableRDMAOnRHELMachine() (bool, error)
	// GetOSPrettyName returns OS name
	GetOSPrettyName() (string, error)
}

type hostManager struct {
	utilsHelper utils.CmdInterface
}

func NewHostManager(utilsInterface utils.CmdInterface) HostManagerInterface {
	return &hostManager{
		utilsHelper: utilsInterface,
	}
}

func (h *hostManager) LoadKernelModule(name string, args ...string) error {
	log.Log.Info("LoadKernelModule(): try to load kernel module", "name", name, "args", args)
	chrootDefinition := utils.GetChrootExtension()
	cmdArgs := strings.Join(args, " ")

	// check if the driver is already loaded in to the system
	isLoaded, err := h.IsKernelModuleLoaded(name)
	if err != nil {
		log.Log.Error(err, "LoadKernelModule(): failed to check if kernel module is already loaded", "name", name)
	}
	if isLoaded {
		log.Log.Info("LoadKernelModule(): kernel module already loaded", "name", name)
		return nil
	}

	_, _, err = h.utilsHelper.RunCommand("/bin/sh", "-c", fmt.Sprintf("%s modprobe %s %s", chrootDefinition, name, cmdArgs))
	if err != nil {
		log.Log.Error(err, "LoadKernelModule(): failed to load kernel module with arguments", "name", name, "args", args)
		return err
	}
	return nil
}

func (h *hostManager) IsKernelModuleLoaded(kernelModuleName string) (bool, error) {
	log.Log.Info("IsKernelModuleLoaded(): check if kernel module is loaded", "name", kernelModuleName)
	chrootDefinition := utils.GetChrootExtension()

	stdout, stderr, err := h.utilsHelper.RunCommand("/bin/sh", "-c", fmt.Sprintf("%s lsmod | grep \"^%s\"", chrootDefinition, kernelModuleName))
	if err != nil && len(stderr) != 0 {
		log.Log.Error(err, "IsKernelModuleLoaded(): failed to check if kernel module is loaded",
			"name", kernelModuleName, "stderr", stderr)
		return false, err
	}
	log.Log.V(2).Info("IsKernelModuleLoaded():", "stdout", stdout)
	if len(stderr) != 0 {
		log.Log.Error(err, "IsKernelModuleLoaded(): failed to check if kernel module is loaded", "name", kernelModuleName, "stderr", stderr)
		return false, fmt.Errorf(stderr)
	}

	if len(stdout) != 0 {
		log.Log.Info("IsKernelModuleLoaded(): kernel module already loaded", "name", kernelModuleName)
		return true, nil
	}

	return false, nil
}

func (h *hostManager) TryEnableTun() {
	if err := h.LoadKernelModule("tun"); err != nil {
		log.Log.Error(err, "tryEnableTun(): TUN kernel module not loaded")
	}
}

func (h *hostManager) TryEnableVhostNet() {
	if err := h.LoadKernelModule("vhost_net"); err != nil {
		log.Log.Error(err, "tryEnableVhostNet(): VHOST_NET kernel module not loaded")
	}
}

func (h *hostManager) TryEnableRdma() (bool, error) {
	log.Log.V(2).Info("tryEnableRdma()")
	chrootDefinition := utils.GetChrootExtension()

	// check if the driver is already loaded in to the system
	_, stderr, mlx4Err := h.utilsHelper.RunCommand("/bin/sh", "-c", fmt.Sprintf("grep --quiet 'mlx4_en' <(%s lsmod)", chrootDefinition))
	if mlx4Err != nil && len(stderr) != 0 {
		log.Log.Error(mlx4Err, "tryEnableRdma(): failed to check for kernel module 'mlx4_en'", "stderr", stderr)
		return false, fmt.Errorf(stderr)
	}

	_, stderr, mlx5Err := h.utilsHelper.RunCommand("/bin/sh", "-c", fmt.Sprintf("grep --quiet 'mlx5_core' <(%s lsmod)", chrootDefinition))
	if mlx5Err != nil && len(stderr) != 0 {
		log.Log.Error(mlx5Err, "tryEnableRdma(): failed to check for kernel module 'mlx5_core'", "stderr", stderr)
		return false, fmt.Errorf(stderr)
	}

	if mlx4Err != nil && mlx5Err != nil {
		log.Log.Error(nil, "tryEnableRdma(): no RDMA capable devices")
		return false, nil
	}

	isRhelSystem, err := h.IsRHELSystem()
	if err != nil {
		log.Log.Error(err, "tryEnableRdma(): failed to check if the machine is base on RHEL")
		return false, err
	}

	// RHEL check
	if isRhelSystem {
		return h.EnableRDMAOnRHELMachine()
	}

	isUbuntuSystem, err := h.IsUbuntuSystem()
	if err != nil {
		log.Log.Error(err, "tryEnableRdma(): failed to check if the machine is base on Ubuntu")
		return false, err
	}

	if isUbuntuSystem {
		return h.EnableRDMAOnUbuntuMachine()
	}

	osName, err := h.GetOSPrettyName()
	if err != nil {
		log.Log.Error(err, "tryEnableRdma(): failed to check OS name")
		return false, err
	}

	log.Log.Error(nil, "tryEnableRdma(): Unsupported OS", "name", osName)
	return false, fmt.Errorf("unable to load RDMA unsupported OS: %s", osName)
}

func (h *hostManager) EnableRDMAOnRHELMachine() (bool, error) {
	log.Log.Info("EnableRDMAOnRHELMachine()")
	isCoreOsSystem, err := h.IsCoreOS()
	if err != nil {
		log.Log.Error(err, "EnableRDMAOnRHELMachine(): failed to check if the machine runs CoreOS")
		return false, err
	}

	// CoreOS check
	if isCoreOsSystem {
		isRDMALoaded, err := h.RdmaIsLoaded()
		if err != nil {
			log.Log.Error(err, "EnableRDMAOnRHELMachine(): failed to check if RDMA kernel modules are loaded")
			return false, err
		}

		return isRDMALoaded, nil
	}

	// RHEL
	log.Log.Info("EnableRDMAOnRHELMachine(): enabling RDMA on RHEL machine")
	isRDMAEnable, err := h.EnableRDMA(rhelRDMAConditionFile, rhelRDMAServiceName, rhelPackageManager)
	if err != nil {
		log.Log.Error(err, "EnableRDMAOnRHELMachine(): failed to enable RDMA on RHEL machine")
		return false, err
	}

	// check if we need to install rdma-core package
	if isRDMAEnable {
		isRDMALoaded, err := h.RdmaIsLoaded()
		if err != nil {
			log.Log.Error(err, "EnableRDMAOnRHELMachine(): failed to check if RDMA kernel modules are loaded")
			return false, err
		}

		// if ib kernel module is not loaded trigger a loading
		if isRDMALoaded {
			err = h.TriggerUdevEvent()
			if err != nil {
				log.Log.Error(err, "EnableRDMAOnRHELMachine() failed to trigger udev event")
				return false, err
			}
		}
	}

	return true, nil
}

func (h *hostManager) EnableRDMAOnUbuntuMachine() (bool, error) {
	log.Log.Info("EnableRDMAOnUbuntuMachine(): enabling RDMA on RHEL machine")
	isRDMAEnable, err := h.EnableRDMA(ubuntuRDMAConditionFile, ubuntuRDMAServiceName, ubuntuPackageManager)
	if err != nil {
		log.Log.Error(err, "EnableRDMAOnUbuntuMachine(): failed to enable RDMA on Ubuntu machine")
		return false, err
	}

	// check if we need to install rdma-core package
	if isRDMAEnable {
		isRDMALoaded, err := h.RdmaIsLoaded()
		if err != nil {
			log.Log.Error(err, "EnableRDMAOnUbuntuMachine(): failed to check if RDMA kernel modules are loaded")
			return false, err
		}

		// if ib kernel module is not loaded trigger a loading
		if isRDMALoaded {
			err = h.TriggerUdevEvent()
			if err != nil {
				log.Log.Error(err, "EnableRDMAOnUbuntuMachine() failed to trigger udev event")
				return false, err
			}
		}
	}

	return true, nil
}

func (h *hostManager) IsRHELSystem() (bool, error) {
	log.Log.Info("IsRHELSystem(): checking for RHEL machine")
	path := redhatReleaseFile
	if !vars.UsingSystemdMode {
		path = pathlib.Join(hostPathFromDaemon, path)
	}
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			log.Log.V(2).Info("IsRHELSystem() not a RHEL machine")
			return false, nil
		}

		log.Log.Error(err, "IsRHELSystem() failed to check for os release file", "path", path)
		return false, err
	}

	return true, nil
}

func (h *hostManager) IsCoreOS() (bool, error) {
	log.Log.Info("IsCoreOS(): checking for CoreOS machine")
	path := redhatReleaseFile
	if !vars.UsingSystemdMode {
		path = pathlib.Join(hostPathFromDaemon, path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		log.Log.Error(err, "IsCoreOS(): failed to read RHEL release file on path", "path", path)
		return false, err
	}

	if strings.Contains(string(data), "CoreOS") {
		return true, nil
	}

	return false, nil
}

func (h *hostManager) IsUbuntuSystem() (bool, error) {
	log.Log.Info("IsUbuntuSystem(): checking for Ubuntu machine")
	path := genericOSReleaseFile
	if !vars.UsingSystemdMode {
		path = pathlib.Join(hostPathFromDaemon, path)
	}

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			log.Log.Error(nil, "IsUbuntuSystem() os-release on path doesn't exist", "path", path)
			return false, err
		}

		log.Log.Error(err, "IsUbuntuSystem() failed to check for os release file", "path", path)
		return false, err
	}

	stdout, stderr, err := h.utilsHelper.RunCommand("/bin/sh", "-c", fmt.Sprintf("grep -i --quiet 'ubuntu' %s", path))
	if err != nil && len(stderr) != 0 {
		log.Log.Error(err, "IsUbuntuSystem(): failed to check for ubuntu operating system name in os-releasae file", "stderr", stderr)
		return false, fmt.Errorf(stderr)
	}

	if len(stdout) > 0 {
		return true, nil
	}

	return false, nil
}

func (h *hostManager) RdmaIsLoaded() (bool, error) {
	log.Log.V(2).Info("RdmaIsLoaded()")
	chrootDefinition := utils.GetChrootExtension()

	// check if the driver is already loaded in to the system
	_, stderr, err := h.utilsHelper.RunCommand("/bin/sh", "-c", fmt.Sprintf("grep --quiet '\\(^ib\\|^rdma\\)' <(%s lsmod)", chrootDefinition))
	if err != nil && len(stderr) != 0 {
		log.Log.Error(err, "RdmaIsLoaded(): fail to check if ib and rdma kernel modules are loaded", "stderr", stderr)
		return false, fmt.Errorf(stderr)
	}

	if err != nil {
		return false, nil
	}

	return true, nil
}

func (h *hostManager) EnableRDMA(conditionFilePath, serviceName, packageManager string) (bool, error) {
	path := conditionFilePath
	if !vars.UsingSystemdMode {
		path = pathlib.Join(hostPathFromDaemon, path)
	}
	log.Log.Info("EnableRDMA(): checking for service file", "path", path)

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			log.Log.V(2).Info("EnableRDMA(): RDMA server doesn't exist")
			err = h.InstallRDMA(packageManager)
			if err != nil {
				log.Log.Error(err, "EnableRDMA() failed to install RDMA package")
				return false, err
			}

			err = h.TriggerUdevEvent()
			if err != nil {
				log.Log.Error(err, "EnableRDMA() failed to trigger udev event")
				return false, err
			}

			return false, nil
		}

		log.Log.Error(err, "EnableRDMA() failed to check for os release file", "path", path)
		return false, err
	}

	log.Log.Info("EnableRDMA(): service installed", "name", serviceName)
	return true, nil
}

func (h *hostManager) InstallRDMA(packageManager string) error {
	log.Log.Info("InstallRDMA(): installing RDMA")
	chrootDefinition := utils.GetChrootExtension()

	stdout, stderr, err := h.utilsHelper.RunCommand("/bin/sh", "-c", fmt.Sprintf("%s %s install -y rdma-core", chrootDefinition, packageManager))
	if err != nil && len(stderr) != 0 {
		log.Log.Error(err, "InstallRDMA(): failed to install RDMA package", "stdout", stdout, "stderr", stderr)
		return err
	}

	return nil
}

func (h *hostManager) TriggerUdevEvent() error {
	log.Log.Info("TriggerUdevEvent(): installing RDMA")

	err := h.ReloadDriver("mlx4_en")
	if err != nil {
		return err
	}

	err = h.ReloadDriver("mlx5_core")
	if err != nil {
		return err
	}

	return nil
}

func (h *hostManager) ReloadDriver(driverName string) error {
	log.Log.Info("ReloadDriver(): reload driver", "name", driverName)
	chrootDefinition := utils.GetChrootExtension()

	_, stderr, err := h.utilsHelper.RunCommand("/bin/sh", "-c", fmt.Sprintf("%s modprobe -r %s && %s modprobe %s", chrootDefinition, driverName, chrootDefinition, driverName))
	if err != nil && len(stderr) != 0 {
		log.Log.Error(err, "InstallRDMA(): failed to reload kernel module",
			"name", driverName, "stderr", stderr)
		return err
	}

	return nil
}

func (h *hostManager) GetOSPrettyName() (string, error) {
	path := genericOSReleaseFile
	if !vars.UsingSystemdMode {
		path = pathlib.Join(hostPathFromDaemon, path)
	}

	log.Log.Info("GetOSPrettyName(): getting os name from os-release file")

	stdout, stderr, err := h.utilsHelper.RunCommand("/bin/sh", "-c", fmt.Sprintf("cat %s | grep PRETTY_NAME | cut -c 13-", path))
	if err != nil && len(stderr) != 0 {
		log.Log.Error(err, "IsUbuntuSystem(): failed to check for ubuntu operating system name in os-releasae file", "stderr", stderr)
		return "", fmt.Errorf(stderr)
	}

	if len(stdout) > 0 {
		return stdout, nil
	}

	return "", fmt.Errorf("failed to find pretty operating system name")
}

// IsKernelLockdownMode returns true when kernel lockdown mode is enabled
// TODO: change this to return error
func (h *hostManager) IsKernelLockdownMode() bool {
	path := utils.GetHostExtension()
	path = filepath.Join(path, "/sys/kernel/security/lockdown")

	stdout, stderr, err := h.utilsHelper.RunCommand("/bin/sh", "-c", "cat", path)
	log.Log.V(2).Info("IsKernelLockdownMode()", "output", stdout, "error", err)
	if err != nil {
		log.Log.Error(err, "IsKernelLockdownMode(): failed to check for lockdown file", "stderr", stderr)
		return false
	}
	return strings.Contains(stdout, "[integrity]") || strings.Contains(stdout, "[confidentiality]")
}
