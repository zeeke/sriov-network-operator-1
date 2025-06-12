package host

import (
	"fmt"
	"net"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/jaypipes/ghw/pkg/cpu"
	"github.com/jaypipes/ghw/pkg/pci"
	"github.com/jaypipes/pcidb"
	netlinkPkg "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/internal/lib/netlink"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/store"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/utils"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/vars"
	mlxutils "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/vendors/mellanox"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/test/util/fakefilesystem"

	sriovv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/vishvananda/netlink"
)

type FakeHostHelper struct {
	utils.CmdInterface
	HostManagerInterface
	store.ManagerInterface
	mlxutils.MellanoxInterface
}

func NewFakeHostHelper() *FakeHostHelper {
	//vars.FilesystemRoot = "/tmp/sriov-test"
	//os.RemoveAll(vars.FilesystemRoot)
	//err := os.MkdirAll(vars.FilesystemRoot, 0755)
	//if err != nil {
	//	panic(err)
	//}

	var err error

	f := &fakefilesystem.FS{
		Dirs:  []string{"/host/proc"},
		Files: map[string][]byte{"/host/proc/cmdline": hostProcCmdLine},
	}

	// TODO - clean filesystem
	vars.FilesystemRoot, _, err = f.Use()
	if err != nil {
		panic(err)
	}

	vars.Destdir = vars.FilesystemRoot

	// TODO - find a better way to refer to bindata
	err = os.CopyFS(vars.FilesystemRoot+"/bindata", os.DirFS("./bindata"))
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll(path.Join(
		vars.FilesystemRoot,
		"/host/etc/udev",
	), 0755)
	if err != nil {
		panic(err)
	}

	sriovv1.InitNicIDMapFromList([]string{
		"8086 158b 154c",
		"15b3 1015 1016",
		"8086 159b 1889",
	})

	sysMock := newSystemMock()

	storeManager, err := store.NewManager()
	if err != nil {
		panic(err)
	}

	hostManager, err := NewHostManager2(sysMock, sysMock, sysMock, sysMock, sysMock, sysMock)
	if err != nil {
		panic(err)
	}

	return &FakeHostHelper{
		HostManagerInterface: hostManager,
		CmdInterface:         sysMock,
		ManagerInterface:     storeManager,
		MellanoxInterface:    mlxutils.New(sysMock),
	}
}

func newSystemMock() *systemMock {
	ret := &systemMock{}

	ret.mockedDevices = []*mockedDevice{
		newIntelE810(),
	}

	return ret
}

func mustParseMAC(s string) net.HardwareAddr {
	ret, err := net.ParseMAC(s)
	if err != nil {
		panic(err)
	}
	return ret
}

type systemMock struct {
	mockedDevices []*mockedDevice
	//pfLinks         []netlink.Link
	//vfLinksByPfName map[string][]netlink.Link
}

func (s *systemMock) findMockDeviceByPCIAddress(pfPciAddress string) *mockedDevice {
	for _, device := range s.mockedDevices {
		if device.pfDevice.Address == pfPciAddress {
			return device
		}
	}

	panic(fmt.Errorf("device %s not found", pfPciAddress))
}

func (s *systemMock) findPICDeviceByAddress(anyPciAddress string) *pci.Device {
	for _, mockDevice := range s.mockedDevices {
		if mockDevice.pfDevice.Address == anyPciAddress {
			return mockDevice.pfDevice
		}

		for _, vfDevice := range mockDevice.vfDevices {
			if vfDevice.Address == anyPciAddress {
				return vfDevice
			}
		}
	}

	panic(fmt.Errorf("device %s not found", anyPciAddress))
}

// CPU implements ghw.GHWLib.
func (s *systemMock) CPU() (*cpu.Info, error) {
	panic("unimplemented")
}

// PCI implements ghw.GHWLib.
func (s *systemMock) PCI() (*pci.Info, error) {
	ret := &pci.Info{}
	for _, device := range s.mockedDevices {
		ret.Devices = append(ret.Devices, device.pciDevices()...)
	}
	return ret, nil
}

// GetVfRepresentor implements sriovnet.SriovnetLib.
func (s *systemMock) GetVfRepresentor(uplink string, vfIndex int) (string, error) {
	panic("unimplemented")
}

// Change implements ethtool.EthtoolLib.
func (s *systemMock) Change(ifaceName string, config map[string]bool) error {
	panic("unimplemented")
}

// FeatureNames implements ethtool.EthtoolLib.
func (s *systemMock) FeatureNames(ifaceName string) (map[string]uint, error) {
	panic("unimplemented")
}

// Features implements ethtool.EthtoolLib.
func (s *systemMock) Features(ifaceName string) (map[string]bool, error) {
	panic("unimplemented")
}

// DevLinkGetDeviceByName implements netlink.NetlinkLib.
func (s *systemMock) DevLinkGetDeviceByName(bus string, device string) (*netlink.DevlinkDevice, error) {
	panic("unimplemented")
}

// DevLinkSetEswitchMode implements netlink.NetlinkLib.
func (s *systemMock) DevLinkSetEswitchMode(dev *netlink.DevlinkDevice, newMode string) error {
	panic("unimplemented")
}

// DevlinkGetDeviceParamByName implements netlink.NetlinkLib.
func (s *systemMock) DevlinkGetDeviceParamByName(bus string, device string, param string) (*netlink.DevlinkParam, error) {
	panic("unimplemented")
}

// DevlinkSetDeviceParam implements netlink.NetlinkLib.
func (s *systemMock) DevlinkSetDeviceParam(bus string, device string, param string, cmode uint8, value interface{}) error {
	panic("unimplemented")
}

// IsLinkAdminStateUp implements netlink.NetlinkLib.
func (s *systemMock) IsLinkAdminStateUp(link netlinkPkg.Link) bool {
	panic("unimplemented")
}

// LinkByIndex implements netlink.NetlinkLib.
func (s *systemMock) LinkByIndex(index int) (netlinkPkg.Link, error) {
	panic("unimplemented")
}

// LinkByName implements netlink.NetlinkLib.
func (s *systemMock) LinkByName(name string) (netlinkPkg.Link, error) {
	panic("unimplemented")
}

// LinkList implements netlink.NetlinkLib.
func (s *systemMock) LinkList() ([]netlinkPkg.Link, error) {
	ret := []netlinkPkg.Link{}
	for _, pf := range s.mockedDevices {
		ret = append(ret, pf.pfLink)

		for _, vfLink := range pf.vfLinks {
			ret = append(ret, vfLink)
		}
	}
	return ret, nil
}

// LinkSetMTU implements netlink.NetlinkLib.
func (s *systemMock) LinkSetMTU(link netlinkPkg.Link, mtu int) error {
	panic("unimplemented")
}

// LinkSetUp implements netlink.NetlinkLib.
func (s *systemMock) LinkSetUp(link netlinkPkg.Link) error {
	panic("unimplemented")
}

// LinkSetVfHardwareAddr implements netlink.NetlinkLib.
func (s *systemMock) LinkSetVfHardwareAddr(link netlinkPkg.Link, vf int, hwaddr net.HardwareAddr) error {
	panic("unimplemented")
}

// LinkSetVfNodeGUID implements netlink.NetlinkLib.
func (s *systemMock) LinkSetVfNodeGUID(link netlinkPkg.Link, vf int, nodeguid net.HardwareAddr) error {
	panic("unimplemented")
}

// LinkSetVfPortGUID implements netlink.NetlinkLib.
func (s *systemMock) LinkSetVfPortGUID(link netlinkPkg.Link, vf int, portguid net.HardwareAddr) error {
	panic("unimplemented")
}

// RdmaLinkByName implements netlink.NetlinkLib.
func (s *systemMock) RdmaLinkByName(name string) (*netlink.RdmaLink, error) {
	panic("unimplemented")
}

// RdmaSystemGetNetnsMode implements netlink.NetlinkLib.
func (s *systemMock) RdmaSystemGetNetnsMode() (string, error) {
	return "shared", nil // TODO
}

// VDPADelDev implements netlink.NetlinkLib.
func (s *systemMock) VDPADelDev(name string) error {
	panic("unimplemented")
}

// VDPAGetDevByName implements netlink.NetlinkLib.
func (s *systemMock) VDPAGetDevByName(name string) (*netlink.VDPADev, error) {
	panic("unimplemented")
}

// VDPANewDev implements netlink.NetlinkLib.
func (s *systemMock) VDPANewDev(name string, mgmtBus string, mgmtName string, params netlink.VDPANewDevParams) error {
	panic("unimplemented")
}

// GetDriverName implements dputils.DPUtilsLib.
func (s *systemMock) GetDriverName(pciAddr string) (string, error) {
	return s.findPICDeviceByAddress(pciAddr).Driver, nil
}

// GetNetNames implements dputils.DPUtilsLib.
func (s *systemMock) GetNetNames(pciAddr string) ([]string, error) {
	return s.findPICDeviceByAddress(pciAddr)., nil
}

// GetSriovVFcapacity implements dputils.DPUtilsLib.
func (s *systemMock) GetSriovVFcapacity(pf string) int {
	panic("unimplemented")
}

// GetVFID implements dputils.DPUtilsLib.
func (s *systemMock) GetVFID(pciAddr string) (vfID int, err error) {
	panic("unimplemented")
}

// GetVFList implements dputils.DPUtilsLib.
func (s *systemMock) GetVFList(pf string) (vfList []string, err error) {
	panic("unimplemented")
}

// GetVFconfigured implements dputils.DPUtilsLib.
func (s *systemMock) GetVFconfigured(pf string) int {
	panic("unimplemented")
}

// IsSriovPF implements dputils.DPUtilsLib.
func (s *systemMock) IsSriovPF(pciAddr string) bool {
	panic("unimplemented")
}

// IsSriovVF implements dputils.DPUtilsLib.
func (s *systemMock) IsSriovVF(pciAddr string) bool {
	for _, mockDevice := range s.mockedDevices {
		if mockDevice.isSriovVF(pciAddr) {
			return true
		}
	}

	return false
}

// SriovConfigured implements dputils.DPUtilsLib.
func (s *systemMock) SriovConfigured(addr string) bool {
	panic("unimplemented")
}

// Chroot implements utils.CmdInterface.
func (s *systemMock) Chroot(string) (func() error, error) {
	return func() error { return nil }, nil
}

// RunCommand implements utils.CmdInterface.
func (s *systemMock) RunCommand(cmd string, args ...string) (string, string, error) {
	fullcmd := cmd + " " + strings.Join(args, " ")

	for cmdRegexp, stubCmd := range stubCommands {
		reg, err := regexp.Compile(cmdRegexp)
		if err != nil {
			panic(err)
		}

		if reg.MatchString(fullcmd) {
			return stubCmd(fullcmd)
		}
	}

	panic(fullcmd)
}

var stubCommands map[string]commandCallback = map[string]commandCallback{
	"/bin/sh -c chroot /tmp/sriov-operator[0-9]+/host lsmod | grep --quiet '.*'": okNoOutput,
	"/bin/sh -c chroot /tmp/sriov-operator[0-9]+/host lsmod | grep \"^.*\"":      okNoOutput,
	"/bin/sh -c chroot /tmp/sriov-operator[0-9]+/host modprobe .*":               okNoOutput,
	"/bin/bash /tmp/sriov-operator[0-9]+/bindata/scripts/udev-find-sriov-pf.sh":  okNoOutput,
	"/bin/sh bindata/scripts/kargs.sh add .*":                                    okNoOutput,
	"/bin/sh bindata/scripts/kargs.sh remove .*":                                 okNoOutput,
}

type commandCallback func(fullcmd string) (string, string, error)

func okNoOutput(_ string) (string, string, error) {
	return "", "", nil
}

type mockedDevice struct {
	pfLink    netlink.Link
	vfLinks   []netlink.Link
	pfDevice  *pci.Device
	vfDevices []*pci.Device
}

func newIntelE810() *mockedDevice {
	return &mockedDevice{
		pfLink: &netlink.Dummy{
			LinkAttrs: netlink.LinkAttrs{
				Name:         "eno1",
				HardwareAddr: mustParseMAC("aa:aa:aa:00:00:01"),
			},
		},
		vfLinks: []netlink.Link{},
		pfDevice: &pci.Device{
			Driver:  "ice",
			Address: "0000:31:00.0",
			Vendor: &pcidb.Vendor{
				ID:   "8086",
				Name: "Intel",
			},
			Product: &pcidb.Product{
				ID:   "159b",
				Name: "Intel Corporation Ethernet Controller E810-XXV for SFP",
			},
			Revision: "0x00",
			Subsystem: &pcidb.Product{
				ID:   "8086:0001",
				Name: "unknown",
			},
			Class: &pcidb.Class{
				ID:   "02",
				Name: "Network controller",
			},
			Subclass: &pcidb.Subclass{
				ID:   "00",
				Name: "Ethernet controller",
			},
			ProgrammingInterface: &pcidb.ProgrammingInterface{
				ID:   "00",
				Name: "unknonw",
			},
		},
		vfDevices: []*pci.Device{},
	}

}

func (m *mockedDevice) pciDevices() []*pci.Device {
	// ret := []*pci.Device{{
	// 	Driver:  "ice",
	// 	Address: "0000:31:00.0",
	// 	Vendor: &pcidb.Vendor{
	// 		ID:   "8086",
	// 		Name: "Intel",
	// 	},
	// 	Product: &pcidb.Product{
	// 		ID:   "159b",
	// 		Name: "Intel Corporation Ethernet Controller E810-XXV for SFP",
	// 	},
	// 	Revision: "0x00",
	// 	Subsystem: &pcidb.Product{
	// 		ID:   "8086:0001",
	// 		Name: "unknown",
	// 	},
	// 	Class: &pcidb.Class{
	// 		ID:   "02",
	// 		Name: "Network controller",
	// 	},
	// 	Subclass: &pcidb.Subclass{
	// 		ID:   "00",
	// 		Name: "Ethernet controller",
	// 	},
	// 	ProgrammingInterface: &pcidb.ProgrammingInterface{
	// 		ID:   "00",
	// 		Name: "unknonw",
	// 	},
	// }}

	// for i := range m.vfLinks {
	// 	ret = append(ret, &pci.Device{
	// 		Driver:  "ice",
	// 		Address: fmt.Sprintf("0000:31:00.%d", i+1), // TODO - handle double digits
	// 		Vendor: &pcidb.Vendor{
	// 			ID:   "8086",
	// 			Name: "Intel",
	// 		},
	// 		Product: &pcidb.Product{
	// 			ID:   "1889",
	// 			Name: "Intel Corporation Ethernet Adaptive Virtual Function",
	// 		},
	// 		Revision: "0x00",
	// 		Subsystem: &pcidb.Product{
	// 			ID:   "8086:0001",
	// 			Name: "unknown",
	// 		},
	// 		Class: &pcidb.Class{
	// 			ID:   "02",
	// 			Name: "Network controller",
	// 		},
	// 		Subclass: &pcidb.Subclass{
	// 			ID:   "00",
	// 			Name: "Ethernet controller",
	// 		},
	// 		ProgrammingInterface: &pcidb.ProgrammingInterface{
	// 			ID:   "00",
	// 			Name: "unknonw",
	// 		},
	// 	})
	// }

	ret := []*pci.Device{m.pfDevice}
	ret = append(ret, m.vfDevices...)
	return ret

}

func (m *mockedDevice) isSriovVF(pciAddress string) bool {
	for _, vfDevice := range m.vfDevices {
		if vfDevice.Address == pciAddress {
			return true
		}
	}
	return false
}

var hostProcCmdLine []byte = []byte("BOOT_IMAGE=(hd0,gpt3)/boot/ostree/rhcos-3a85e2a0d869da4c0fa4f64c3ab3a1d3826e217cb5852ba645f9667f0547ff17/vmlinuz-5.14.0-570.17.1.el9_6.x86_64 rw ostree=/ostree/boot.0/rhcos/3a85e2a0d869da4c0fa4f64c3ab3a1d3826e217cb5852ba645f9667f0547ff17/0 ignition.platform.id=metal ip=dhcp root=UUID=3df22ef9-b181-4a4a-bc8f-ee4f105cabcf rw rootflags=prjquota boot=UUID=90d44ef1-0af2-4ccb-911e-ee7af86c62da systemd.unified_cgroup_hierarchy=1 cgroup_no_v1=all psi=0 intel_iommu=on iommu=pt")
