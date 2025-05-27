package host

import (
	"net"
	"regexp"
	"strings"

	"github.com/jaypipes/ghw/pkg/cpu"
	"github.com/jaypipes/ghw/pkg/pci"
	netlinkPkg "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/internal/lib/netlink"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/store"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/utils"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/vars"
	mlxutils "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/vendors/mellanox"

	"github.com/vishvananda/netlink"
)

type FakeHostHelper struct {
	utils.CmdInterface
	HostManagerInterface
	store.ManagerInterface
	mlxutils.MellanoxInterface
}

func NewFakeHostHelper() *FakeHostHelper {
	vars.FilesystemRoot = "/tmp"
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

	ret.pfLinks = []netlink.Link{
		&netlink.Dummy{
			LinkAttrs: netlink.LinkAttrs{
				Name:         "eno1",
				HardwareAddr: mustParseMAC("aa:aa:aa:00:00:01"),
			},
		},
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
	pfLinks         []netlink.Link
	vfLinksByPfName map[string][]netlink.Link
}

// CPU implements ghw.GHWLib.
func (s *systemMock) CPU() (*cpu.Info, error) {
	panic("unimplemented")
}

// PCI implements ghw.GHWLib.
func (s *systemMock) PCI() (*pci.Info, error) {
	panic("unimplemented")
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
	for _, pf := range s.pfLinks {
		ret = append(ret, pf)

		vfs := s.vfLinksByPfName[pf.Attrs().Name]
		for _, vf := range vfs {
			ret = append(ret, vf)
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
	panic("unimplemented")
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
	panic("unimplemented")
}

// GetNetNames implements dputils.DPUtilsLib.
func (s *systemMock) GetNetNames(pciAddr string) ([]string, error) {
	panic("unimplemented")
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
	panic("unimplemented")
}

// SriovConfigured implements dputils.DPUtilsLib.
func (s *systemMock) SriovConfigured(addr string) bool {
	panic("unimplemented")
}

// Chroot implements utils.CmdInterface.
func (s *systemMock) Chroot(string) (func() error, error) {
	panic("unimplemented")
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
	"/bin/sh -c chroot /tmp/host lsmod | grep --quiet '.*'": okNoOutput,
	"/bin/sh -c chroot /tmp/host modprobe .*":               okNoOutput,
	"/bin/bash /tmp/bindata/scripts/udev-find-sriov-pf.sh":  okNoOutput,
}

type commandCallback func(fullcmd string) (string, string, error)

func okNoOutput(_ string) (string, string, error) {
	return "", "", nil
}
