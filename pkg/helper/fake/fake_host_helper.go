package fake

import (
	v1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/helper"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/store"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/types"
	mlxutils "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/vendors/mellanox"
	"github.com/vishvananda/netlink"
)

var x helper.HostHelpersInterface = NewHostHelpers()

type FakeHostHelper struct {
}

// CleanSriovFilesFromHost implements helper.HostHelpersInterface.
func (f *FakeHostHelper) CleanSriovFilesFromHost(isOpenShift bool) error {
	panic("unimplemented")
}

// DiscoverRDMASubsystem implements helper.HostHelpersInterface.
func (f *FakeHostHelper) DiscoverRDMASubsystem() (string, error) {
	panic("unimplemented")
}

// GetCPUVendor implements helper.HostHelpersInterface.
func (f *FakeHostHelper) GetCPUVendor() (types.CPUVendor, error) {
	panic("unimplemented")
}

// ReadConfFile implements helper.HostHelpersInterface.
func (f *FakeHostHelper) ReadConfFile() (spec *types.SriovConfig, err error) {
	panic("unimplemented")
}

// ReadSriovResult implements helper.HostHelpersInterface.
func (f *FakeHostHelper) ReadSriovResult() (*types.SriovResult, error) {
	panic("unimplemented")
}

// ReadSriovSupportedNics implements helper.HostHelpersInterface.
func (f *FakeHostHelper) ReadSriovSupportedNics() ([]string, error) {
	panic("unimplemented")
}

// RemoveSriovResult implements helper.HostHelpersInterface.
func (f *FakeHostHelper) RemoveSriovResult() error {
	panic("unimplemented")
}

// SetRDMASubsystem implements helper.HostHelpersInterface.
func (f *FakeHostHelper) SetRDMASubsystem(mode string) error {
	panic("unimplemented")
}

// WaitUdevEventsProcessed implements helper.HostHelpersInterface.
func (f *FakeHostHelper) WaitUdevEventsProcessed(timeout int) error {
	panic("unimplemented")
}

// WriteConfFile implements helper.HostHelpersInterface.
func (f *FakeHostHelper) WriteConfFile(newState *v1.SriovNetworkNodeState) (bool, error) {
	panic("unimplemented")
}

// WriteSriovResult implements helper.HostHelpersInterface.
func (f *FakeHostHelper) WriteSriovResult(result *types.SriovResult) error {
	panic("unimplemented")
}

// WriteSriovSupportedNics implements helper.HostHelpersInterface.
func (f *FakeHostHelper) WriteSriovSupportedNics() error {
	panic("unimplemented")
}

// AddDisableNMUdevRule implements helper.HostHelpersInterface.
func (f *FakeHostHelper) AddDisableNMUdevRule(pfPciAddress string) error {
	panic("unimplemented")
}

// AddPersistPFNameUdevRule implements helper.HostHelpersInterface.
func (f *FakeHostHelper) AddPersistPFNameUdevRule(pfPciAddress string, pfName string) error {
	panic("unimplemented")
}

// AddVfRepresentorUdevRule implements helper.HostHelpersInterface.
func (f *FakeHostHelper) AddVfRepresentorUdevRule(pfPciAddress string, pfName string, pfSwitchID string, pfSwitchPort string) error {
	panic("unimplemented")
}

// BindDefaultDriver implements helper.HostHelpersInterface.
func (f *FakeHostHelper) BindDefaultDriver(pciAddr string) error {
	panic("unimplemented")
}

// BindDpdkDriver implements helper.HostHelpersInterface.
func (f *FakeHostHelper) BindDpdkDriver(pciAddr string, driver string) error {
	panic("unimplemented")
}

// BindDriverByBusAndDevice implements helper.HostHelpersInterface.
func (f *FakeHostHelper) BindDriverByBusAndDevice(bus string, device string, driver string) error {
	panic("unimplemented")
}

// CheckRDMAEnabled implements helper.HostHelpersInterface.
func (f *FakeHostHelper) CheckRDMAEnabled() (bool, error) {
	panic("unimplemented")
}

// Chroot implements helper.HostHelpersInterface.
func (f *FakeHostHelper) Chroot(string) (func() error, error) {
	panic("unimplemented")
}

// ClearPCIAddressFolder implements helper.HostHelpersInterface.
func (f *FakeHostHelper) ClearPCIAddressFolder() error {
	panic("unimplemented")
}

// CompareServices implements helper.HostHelpersInterface.
func (f *FakeHostHelper) CompareServices(serviceA *types.Service, serviceB *types.Service) (bool, error) {
	panic("unimplemented")
}

// ConfigSriovDeviceVirtual implements helper.HostHelpersInterface.
func (f *FakeHostHelper) ConfigSriovDeviceVirtual(iface *v1.Interface) error {
	panic("unimplemented")
}

// ConfigSriovInterfaces implements helper.HostHelpersInterface.
func (f *FakeHostHelper) ConfigSriovInterfaces(storeManager store.ManagerInterface, interfaces []v1.Interface, ifaceStatuses []v1.InterfaceExt, skipVFConfiguration bool) error {
	panic("unimplemented")
}

// ConfigureBridges implements helper.HostHelpersInterface.
func (f *FakeHostHelper) ConfigureBridges(bridgesSpec v1.Bridges, bridgesStatus v1.Bridges) error {
	panic("unimplemented")
}

// ConfigureVfGUID implements helper.HostHelpersInterface.
func (f *FakeHostHelper) ConfigureVfGUID(vfAddr string, pfAddr string, vfID int, pfLink netlink.Link) error {
	panic("unimplemented")
}

// CreateVDPADevice implements helper.HostHelpersInterface.
func (f *FakeHostHelper) CreateVDPADevice(pciAddr string, vdpaType string) error {
	panic("unimplemented")
}

// DeleteVDPADevice implements helper.HostHelpersInterface.
func (f *FakeHostHelper) DeleteVDPADevice(pciAddr string) error {
	panic("unimplemented")
}

// DetachInterfaceFromManagedBridge implements helper.HostHelpersInterface.
func (f *FakeHostHelper) DetachInterfaceFromManagedBridge(pciAddr string) error {
	panic("unimplemented")
}

// DiscoverBridges implements helper.HostHelpersInterface.
func (f *FakeHostHelper) DiscoverBridges() (v1.Bridges, error) {
	panic("unimplemented")
}

// DiscoverSriovDevices implements helper.HostHelpersInterface.
func (f *FakeHostHelper) DiscoverSriovDevices(storeManager store.ManagerInterface) ([]v1.InterfaceExt, error) {
	panic("unimplemented")
}

// DiscoverVDPAType implements helper.HostHelpersInterface.
func (f *FakeHostHelper) DiscoverVDPAType(pciAddr string) string {
	panic("unimplemented")
}

// EnableHwTcOffload implements helper.HostHelpersInterface.
func (f *FakeHostHelper) EnableHwTcOffload(ifaceName string) error {
	panic("unimplemented")
}

// EnableService implements helper.HostHelpersInterface.
func (f *FakeHostHelper) EnableService(service *types.Service) error {
	panic("unimplemented")
}

// GetCheckPointNodeState implements helper.HostHelpersInterface.
func (f *FakeHostHelper) GetCheckPointNodeState() (*v1.SriovNetworkNodeState, error) {
	panic("unimplemented")
}

// GetCurrentKernelArgs implements helper.HostHelpersInterface.
func (f *FakeHostHelper) GetCurrentKernelArgs() (string, error) {
	panic("unimplemented")
}

// GetDevlinkDeviceParam implements helper.HostHelpersInterface.
func (f *FakeHostHelper) GetDevlinkDeviceParam(pciAddr string, paramName string) (string, error) {
	panic("unimplemented")
}

// GetDriverByBusAndDevice implements helper.HostHelpersInterface.
func (f *FakeHostHelper) GetDriverByBusAndDevice(bus string, device string) (string, error) {
	panic("unimplemented")
}

// GetInterfaceIndex implements helper.HostHelpersInterface.
func (f *FakeHostHelper) GetInterfaceIndex(pciAddr string) (int, error) {
	panic("unimplemented")
}

// GetLinkType implements helper.HostHelpersInterface.
func (f *FakeHostHelper) GetLinkType(name string) string {
	panic("unimplemented")
}

// GetMellanoxBlueFieldMode implements helper.HostHelpersInterface.
func (f *FakeHostHelper) GetMellanoxBlueFieldMode(string) (mlxutils.BlueFieldMode, error) {
	panic("unimplemented")
}

// GetMlxNicFwData implements helper.HostHelpersInterface.
func (f *FakeHostHelper) GetMlxNicFwData(pciAddress string) (current *mlxutils.MlxNic, next *mlxutils.MlxNic, err error) {
	panic("unimplemented")
}

// GetNetDevLinkAdminState implements helper.HostHelpersInterface.
func (f *FakeHostHelper) GetNetDevLinkAdminState(ifaceName string) string {
	panic("unimplemented")
}

// GetNetDevLinkSpeed implements helper.HostHelpersInterface.
func (f *FakeHostHelper) GetNetDevLinkSpeed(name string) string {
	panic("unimplemented")
}

// GetNetDevMac implements helper.HostHelpersInterface.
func (f *FakeHostHelper) GetNetDevMac(name string) string {
	panic("unimplemented")
}

// GetNetDevNodeGUID implements helper.HostHelpersInterface.
func (f *FakeHostHelper) GetNetDevNodeGUID(pciAddr string) string {
	panic("unimplemented")
}

// GetNetdevMTU implements helper.HostHelpersInterface.
func (f *FakeHostHelper) GetNetdevMTU(pciAddr string) int {
	panic("unimplemented")
}

// GetNicSriovMode implements helper.HostHelpersInterface.
func (f *FakeHostHelper) GetNicSriovMode(pciAddr string) string {
	panic("unimplemented")
}

// GetPciAddressFromInterfaceName implements helper.HostHelpersInterface.
func (f *FakeHostHelper) GetPciAddressFromInterfaceName(interfaceName string) (string, error) {
	panic("unimplemented")
}

// GetPhysPortName implements helper.HostHelpersInterface.
func (f *FakeHostHelper) GetPhysPortName(name string) (string, error) {
	panic("unimplemented")
}

// GetPhysSwitchID implements helper.HostHelpersInterface.
func (f *FakeHostHelper) GetPhysSwitchID(name string) (string, error) {
	panic("unimplemented")
}

// HasDriver implements helper.HostHelpersInterface.
func (f *FakeHostHelper) HasDriver(pciAddr string) (bool, string) {
	panic("unimplemented")
}

// IsKernelArgsSet implements helper.HostHelpersInterface.
func (f *FakeHostHelper) IsKernelArgsSet(cmdLine string, karg string) bool {
	panic("unimplemented")
}

// IsKernelLockdownMode implements helper.HostHelpersInterface.
func (f *FakeHostHelper) IsKernelLockdownMode() bool {
	panic("unimplemented")
}

// IsKernelModuleLoaded implements helper.HostHelpersInterface.
func (f *FakeHostHelper) IsKernelModuleLoaded(name string) (bool, error) {
	panic("unimplemented")
}

// IsServiceEnabled implements helper.HostHelpersInterface.
func (f *FakeHostHelper) IsServiceEnabled(servicePath string) (bool, error) {
	panic("unimplemented")
}

// IsServiceExist implements helper.HostHelpersInterface.
func (f *FakeHostHelper) IsServiceExist(servicePath string) (bool, error) {
	panic("unimplemented")
}

// IsSwitchdev implements helper.HostHelpersInterface.
func (f *FakeHostHelper) IsSwitchdev(name string) bool {
	panic("unimplemented")
}

// LoadKernelModule implements helper.HostHelpersInterface.
func (f *FakeHostHelper) LoadKernelModule(name string, args ...string) error {
	panic("unimplemented")
}

// LoadPfsStatus implements helper.HostHelpersInterface.
func (f *FakeHostHelper) LoadPfsStatus(pciAddress string) (*v1.Interface, bool, error) {
	panic("unimplemented")
}

// LoadUdevRules implements helper.HostHelpersInterface.
func (f *FakeHostHelper) LoadUdevRules() error {
	panic("unimplemented")
}

// MlxConfigFW implements helper.HostHelpersInterface.
func (f *FakeHostHelper) MlxConfigFW(attributesToChange map[string]mlxutils.MlxNic) error {
	panic("unimplemented")
}

// MlxResetFW implements helper.HostHelpersInterface.
func (f *FakeHostHelper) MlxResetFW(pciAddresses []string) error {
	panic("unimplemented")
}

// MstConfigReadData implements helper.HostHelpersInterface.
func (f *FakeHostHelper) MstConfigReadData(string) (string, string, error) {
	panic("unimplemented")
}

// PrepareNMUdevRule implements helper.HostHelpersInterface.
func (f *FakeHostHelper) PrepareNMUdevRule(supportedVfIds []string) error {
	panic("unimplemented")
}

// PrepareVFRepUdevRule implements helper.HostHelpersInterface.
func (f *FakeHostHelper) PrepareVFRepUdevRule() error {
	panic("unimplemented")
}

// ReadService implements helper.HostHelpersInterface.
func (f *FakeHostHelper) ReadService(servicePath string) (*types.Service, error) {
	panic("unimplemented")
}

// ReadServiceInjectionManifestFile implements helper.HostHelpersInterface.
func (f *FakeHostHelper) ReadServiceInjectionManifestFile(path string) (*types.Service, error) {
	panic("unimplemented")
}

// ReadServiceManifestFile implements helper.HostHelpersInterface.
func (f *FakeHostHelper) ReadServiceManifestFile(path string) (*types.Service, error) {
	panic("unimplemented")
}

// RebindVfToDefaultDriver implements helper.HostHelpersInterface.
func (f *FakeHostHelper) RebindVfToDefaultDriver(pciAddr string) error {
	panic("unimplemented")
}

// RemoveDisableNMUdevRule implements helper.HostHelpersInterface.
func (f *FakeHostHelper) RemoveDisableNMUdevRule(pfPciAddress string) error {
	panic("unimplemented")
}

// RemovePersistPFNameUdevRule implements helper.HostHelpersInterface.
func (f *FakeHostHelper) RemovePersistPFNameUdevRule(pfPciAddress string) error {
	panic("unimplemented")
}

// RemovePfAppliedStatus implements helper.HostHelpersInterface.
func (f *FakeHostHelper) RemovePfAppliedStatus(pciAddress string) error {
	panic("unimplemented")
}

// RemoveVfRepresentorUdevRule implements helper.HostHelpersInterface.
func (f *FakeHostHelper) RemoveVfRepresentorUdevRule(pfPciAddress string) error {
	panic("unimplemented")
}

// ResetSriovDevice implements helper.HostHelpersInterface.
func (f *FakeHostHelper) ResetSriovDevice(ifaceStatus v1.InterfaceExt) error {
	panic("unimplemented")
}

// RunCommand implements helper.HostHelpersInterface.
func (f *FakeHostHelper) RunCommand(string, ...string) (string, string, error) {
	panic("unimplemented")
}

// SaveLastPfAppliedStatus implements helper.HostHelpersInterface.
func (f *FakeHostHelper) SaveLastPfAppliedStatus(PfInfo *v1.Interface) error {
	panic("unimplemented")
}

// SetDevlinkDeviceParam implements helper.HostHelpersInterface.
func (f *FakeHostHelper) SetDevlinkDeviceParam(pciAddr string, paramName string, value string) error {
	panic("unimplemented")
}

// SetNetdevMTU implements helper.HostHelpersInterface.
func (f *FakeHostHelper) SetNetdevMTU(pciAddr string, mtu int) error {
	panic("unimplemented")
}

// SetNicSriovMode implements helper.HostHelpersInterface.
func (f *FakeHostHelper) SetNicSriovMode(pciAddr string, mode string) error {
	panic("unimplemented")
}

// SetSriovNumVfs implements helper.HostHelpersInterface.
func (f *FakeHostHelper) SetSriovNumVfs(pciAddr string, numVfs int) error {
	panic("unimplemented")
}

// SetVfAdminMac implements helper.HostHelpersInterface.
func (f *FakeHostHelper) SetVfAdminMac(vfAddr string, pfLink netlink.Link, vfLink netlink.Link) error {
	panic("unimplemented")
}

// TryEnableTun implements helper.HostHelpersInterface.
func (f *FakeHostHelper) TryEnableTun() {
	panic("unimplemented")
}

// TryEnableVhostNet implements helper.HostHelpersInterface.
func (f *FakeHostHelper) TryEnableVhostNet() {
	panic("unimplemented")
}

// TryGetInterfaceName implements helper.HostHelpersInterface.
func (f *FakeHostHelper) TryGetInterfaceName(pciAddr string) string {
	panic("unimplemented")
}

// TryToGetVirtualInterfaceName implements helper.HostHelpersInterface.
func (f *FakeHostHelper) TryToGetVirtualInterfaceName(pciAddr string) string {
	panic("unimplemented")
}

// Unbind implements helper.HostHelpersInterface.
func (f *FakeHostHelper) Unbind(pciAddr string) error {
	panic("unimplemented")
}

// UnbindDriverByBusAndDevice implements helper.HostHelpersInterface.
func (f *FakeHostHelper) UnbindDriverByBusAndDevice(bus string, device string) error {
	panic("unimplemented")
}

// UnbindDriverIfNeeded implements helper.HostHelpersInterface.
func (f *FakeHostHelper) UnbindDriverIfNeeded(pciAddr string, isRdma bool) error {
	panic("unimplemented")
}

// UpdateSystemService implements helper.HostHelpersInterface.
func (f *FakeHostHelper) UpdateSystemService(serviceObj *types.Service) error {
	panic("unimplemented")
}

// VFIsReady implements helper.HostHelpersInterface.
func (f *FakeHostHelper) VFIsReady(pciAddr string) (netlink.Link, error) {
	panic("unimplemented")
}

// WriteCheckpointFile implements helper.HostHelpersInterface.
func (f *FakeHostHelper) WriteCheckpointFile(*v1.SriovNetworkNodeState) error {
	panic("unimplemented")
}

func NewHostHelpers() *FakeHostHelper {
	return &FakeHostHelper{}
}
