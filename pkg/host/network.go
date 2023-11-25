package host

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/jaypipes/ghw"
	"github.com/vishvananda/netlink"
	"sigs.k8s.io/controller-runtime/pkg/log"

	dputils "github.com/k8snetworkplumbingwg/sriov-network-device-plugin/pkg/utils"

	sriovnetworkv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/global/consts"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/global/vars"
)

// TryToGetVirtualInterfaceName get the interface name of a virtio interface
func (h *HostManager) TryToGetVirtualInterfaceName(pciAddr string) string {
	log.Log.Info("TryToGetVirtualInterfaceName() get interface name for device", "device", pciAddr)

	// To support different driver that is not virtio-pci like mlx
	name := h.TryGetInterfaceName(pciAddr)
	if name != "" {
		return name
	}

	netDir, err := filepath.Glob(filepath.Join(vars.FilesystemRoot, consts.SysBusPciDevices, pciAddr, "virtio*", "net"))
	if err != nil || len(netDir) < 1 {
		return ""
	}

	fInfos, err := os.ReadDir(netDir[0])
	if err != nil {
		log.Log.Error(err, "TryToGetVirtualInterfaceName(): failed to read net directory", "dir", netDir[0])
		return ""
	}

	names := make([]string, 0)
	for _, f := range fInfos {
		names = append(names, f.Name())
	}

	if len(names) < 1 {
		return ""
	}

	return names[0]
}

func (h *HostManager) TryGetInterfaceName(pciAddr string) string {
	names, err := dputils.GetNetNames(pciAddr)
	if err != nil || len(names) < 1 {
		return ""
	}
	netDevName := names[0]

	// Switchdev PF and their VFs representors are existing under the same PCI address since kernel 5.8
	// if device is switchdev then return PF name
	for _, name := range names {
		if !h.IsSwitchdev(name) {
			continue
		}
		// Try to get the phys port name, if not exists then fallback to check without it
		// phys_port_name should be in formant p<port-num> e.g p0,p1,p2 ...etc.
		if physPortName, err := h.GetPhysPortName(name); err == nil {
			if !vars.PfPhysPortNameRe.MatchString(physPortName) {
				continue
			}
		}
		return name
	}

	log.Log.V(2).Info("tryGetInterfaceName()", "name", netDevName)
	return netDevName
}

func (h *HostManager) GetNicSriovMode(pciAddress string) (string, error) {
	log.Log.V(2).Info("GetNicSriovMode()", "device", pciAddress)

	devLink, err := netlink.DevLinkGetDeviceByName("pci", pciAddress)
	if err != nil {
		if errors.Is(err, syscall.ENODEV) {
			// the device doesn't support devlink
			return "", nil
		}
		return "", err
	}

	return devLink.Attrs.Eswitch.Mode, nil
}

func (h *HostManager) GetPhysSwitchID(name string) (string, error) {
	swIDFile := filepath.Join(vars.FilesystemRoot, consts.SysClassNet, name, "phys_switch_id")
	physSwitchID, err := os.ReadFile(swIDFile)
	if err != nil {
		return "", err
	}
	if physSwitchID != nil {
		return strings.TrimSpace(string(physSwitchID)), nil
	}
	return "", nil
}

func (h *HostManager) GetPhysPortName(name string) (string, error) {
	devicePortNameFile := filepath.Join(vars.FilesystemRoot, consts.SysClassNet, name, "phys_port_name")
	physPortName, err := os.ReadFile(devicePortNameFile)
	if err != nil {
		return "", err
	}
	if physPortName != nil {
		return strings.TrimSpace(string(physPortName)), nil
	}
	return "", nil
}

func (h *HostManager) IsSwitchdev(name string) bool {
	switchID, err := h.GetPhysSwitchID(name)
	if err != nil || switchID == "" {
		return false
	}

	return true
}

func (h *HostManager) GetNetdevMTU(pciAddr string) int {
	log.Log.V(2).Info("GetNetdevMTU(): get MTU", "device", pciAddr)
	ifaceName := h.TryGetInterfaceName(pciAddr)
	if ifaceName == "" {
		return 0
	}
	mtuFile := "net/" + ifaceName + "/mtu"
	mtuFilePath := filepath.Join(vars.FilesystemRoot, consts.SysBusPciDevices, pciAddr, mtuFile)
	data, err := os.ReadFile(mtuFilePath)
	if err != nil {
		log.Log.Error(err, "GetNetdevMTU(): fail to read mtu file", "path", mtuFilePath)
		return 0
	}
	mtu, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		log.Log.Error(err, "GetNetdevMTU(): fail to convert mtu to int", "raw-mtu", strings.TrimSpace(string(data)))
		return 0
	}
	return mtu
}

func (h *HostManager) SetNetdevMTU(pciAddr string, mtu int) error {
	log.Log.V(2).Info("SetNetdevMTU(): set MTU", "device", pciAddr, "mtu", mtu)
	if mtu <= 0 {
		log.Log.V(2).Info("SetNetdevMTU(): refusing to set MTU", "mtu", mtu)
		return nil
	}
	b := backoff.NewConstantBackOff(1 * time.Second)
	err := backoff.Retry(func() error {
		ifaceName, err := dputils.GetNetNames(pciAddr)
		if err != nil {
			log.Log.Error(err, "SetNetdevMTU(): fail to get interface name", "device", pciAddr)
			return err
		}
		if len(ifaceName) < 1 {
			return fmt.Errorf("SetNetdevMTU(): interface name is empty")
		}
		mtuFile := "net/" + ifaceName[0] + "/mtu"
		mtuFilePath := filepath.Join(vars.FilesystemRoot, consts.SysBusPciDevices, pciAddr, mtuFile)
		return os.WriteFile(mtuFilePath, []byte(strconv.Itoa(mtu)), os.ModeAppend)
	}, backoff.WithMaxRetries(b, 10))
	if err != nil {
		log.Log.Error(err, "SetNetdevMTU(): fail to write mtu file after retrying")
		return err
	}
	return nil
}

func (h *HostManager) GetNetDevMac(ifaceName string) string {
	log.Log.V(2).Info("GetNetDevMac(): get Mac", "device", ifaceName)
	macFilePath := filepath.Join(vars.FilesystemRoot, consts.SysClassNet, ifaceName, "address")
	data, err := os.ReadFile(macFilePath)
	if err != nil {
		log.Log.Error(err, "GetNetDevMac(): fail to read Mac file", "path", macFilePath)
		return ""
	}

	return strings.TrimSpace(string(data))
}

func (h *HostManager) GetNetDevLinkSpeed(ifaceName string) string {
	log.Log.V(2).Info("GetNetDevLinkSpeed(): get LinkSpeed", "device", ifaceName)
	speedFilePath := filepath.Join(vars.FilesystemRoot, consts.SysClassNet, ifaceName, "speed")
	data, err := os.ReadFile(speedFilePath)
	if err != nil {
		log.Log.Error(err, "GetNetDevLinkSpeed(): fail to read Link Speed file", "path", speedFilePath)
		return ""
	}

	return fmt.Sprintf("%s Mb/s", strings.TrimSpace(string(data)))
}

func (h *HostManager) GetLinkType(ifaceStatus sriovnetworkv1.InterfaceExt) string {
	log.Log.V(2).Info("GetLinkType()", "device", ifaceStatus.PciAddress)
	if ifaceStatus.Name != "" {
		link, err := netlink.LinkByName(ifaceStatus.Name)
		if err != nil {
			log.Log.Error(err, "GetLinkType(): failed to get link", "device", ifaceStatus.Name)
			return ""
		}
		linkType := link.Attrs().EncapType
		if linkType == "ether" {
			return consts.LinkTypeETH
		} else if linkType == "infiniband" {
			return consts.LinkTypeIB
		}
	}

	return ""
}

func (h *HostManager) DiscoverSriovDevices(storeManager StoreManagerInterface) ([]sriovnetworkv1.InterfaceExt, error) {
	log.Log.V(2).Info("DiscoverSriovDevices")
	pfList := []sriovnetworkv1.InterfaceExt{}

	pci, err := ghw.PCI()
	if err != nil {
		return nil, fmt.Errorf("DiscoverSriovDevices(): error getting PCI info: %v", err)
	}

	devices := pci.ListDevices()
	if len(devices) == 0 {
		return nil, fmt.Errorf("DiscoverSriovDevices(): could not retrieve PCI devices")
	}

	for _, device := range devices {
		devClass, err := strconv.ParseInt(device.Class.ID, 16, 64)
		if err != nil {
			log.Log.Error(err, "DiscoverSriovDevices(): unable to parse device class, skipping",
				"device", device)
			continue
		}
		if devClass != consts.NetClass {
			// Not network device
			continue
		}

		// TODO: exclude devices used by host system

		if dputils.IsSriovVF(device.Address) {
			continue
		}

		driver, err := dputils.GetDriverName(device.Address)
		if err != nil {
			log.Log.Error(err, "DiscoverSriovDevices(): unable to parse device driver for device, skipping", "device", device)
			continue
		}

		deviceNames, err := dputils.GetNetNames(device.Address)
		if err != nil {
			log.Log.Error(err, "DiscoverSriovDevices(): unable to get device names for device, skipping", "device", device)
			continue
		}

		if len(deviceNames) == 0 {
			// no network devices found, skipping device
			continue
		}

		if !vars.DevMode {
			if !sriovnetworkv1.IsSupportedModel(device.Vendor.ID, device.Product.ID) {
				log.Log.Info("DiscoverSriovDevices(): unsupported device", "device", device)
				continue
			}
		}

		iface := sriovnetworkv1.InterfaceExt{
			PciAddress: device.Address,
			Driver:     driver,
			Vendor:     device.Vendor.ID,
			DeviceID:   device.Product.ID,
		}
		if mtu := h.GetNetdevMTU(device.Address); mtu > 0 {
			iface.Mtu = mtu
		}
		if name := h.TryGetInterfaceName(device.Address); name != "" {
			iface.Name = name
			iface.Mac = h.GetNetDevMac(name)
			iface.LinkSpeed = h.GetNetDevLinkSpeed(name)
		}
		iface.LinkType = h.GetLinkType(iface)

		pfStatus, exist, err := storeManager.LoadPfsStatus(iface.PciAddress)
		if err != nil {
			log.Log.Error(err, "DiscoverSriovDevices(): failed to load PF status from disk")
		} else {
			if exist {
				iface.ExternallyManaged = pfStatus.ExternallyManaged
			}
		}

		if dputils.IsSriovPF(device.Address) {
			iface.TotalVfs = dputils.GetSriovVFcapacity(device.Address)
			iface.NumVfs = dputils.GetVFconfigured(device.Address)
			if iface.EswitchMode, err = h.GetNicSriovMode(device.Address); err != nil {
				log.Log.Error(err, "DiscoverSriovDevices(): warning, unable to get device eswitch mode",
					"device", device.Address)
			}
			if dputils.SriovConfigured(device.Address) {
				vfs, err := dputils.GetVFList(device.Address)
				if err != nil {
					log.Log.Error(err, "DiscoverSriovDevices(): unable to parse VFs for device, skipping",
						"device", device)
					continue
				}
				for _, vf := range vfs {
					instance := h.GetVfInfo(vf, devices)
					iface.VFs = append(iface.VFs, instance)
				}
			}
		}
		pfList = append(pfList, iface)
	}

	return pfList, nil
}
