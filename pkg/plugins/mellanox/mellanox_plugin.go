package mellanox

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/log"

	sriovnetworkv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/consts"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/helper"
	plugin "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/plugins"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/vars"
	mlx "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/vendors/mellanox"
)

var PluginName = "mellanox"

type MellanoxPlugin struct {
	PluginName string
	helpers    helper.HostHelpersInterface
}

var pciAddressesToReset []string
var attributesToChange map[string]mlx.MlxNic
var mellanoxNicsStatus map[string]map[string]sriovnetworkv1.InterfaceExt
var mellanoxNicsSpec map[string]sriovnetworkv1.Interface

// Initialize our plugin and set up initial values
func NewMellanoxPlugin(helpers helper.HostHelpersInterface) (plugin.VendorPlugin, error) {
	mellanoxNicsStatus = map[string]map[string]sriovnetworkv1.InterfaceExt{}

	return &MellanoxPlugin{
		PluginName: PluginName,
		helpers:    helpers,
	}, nil
}

// Name returns the name of the plugin
func (p *MellanoxPlugin) Name() string {
	return p.PluginName
}

// OnNodeStateChange Invoked when SriovNetworkNodeState CR is created or updated, return if need dain and/or reboot node
func (p *MellanoxPlugin) OnNodeStateChange(new *sriovnetworkv1.SriovNetworkNodeState) (needDrain bool, needReboot bool, err error) {
	log.Log.Info("mellanox plugin OnNodeStateChange()")

	needDrain = false
	needReboot = false
	err = nil
	pciAddressesToReset = []string{}
	attributesToChange = map[string]mlx.MlxNic{}
	mellanoxNicsStatus = map[string]map[string]sriovnetworkv1.InterfaceExt{}
	mellanoxNicsSpec = map[string]sriovnetworkv1.Interface{}
	processedNics := map[string]bool{}

	// fill mellanoxNicsStatus
	for _, iface := range new.Status.Interfaces {
		if iface.Vendor != mlx.MellanoxVendorID {
			continue
		}

		pciPrefix := mlx.GetPciAddressPrefix(iface.PciAddress)
		if ifaces, ok := mellanoxNicsStatus[pciPrefix]; ok {
			ifaces[iface.PciAddress] = iface
		} else {
			mellanoxNicsStatus[pciPrefix] = map[string]sriovnetworkv1.InterfaceExt{iface.PciAddress: iface}
		}
	}

	// Add only mellanox cards that required changes in the map, to help track dual port NICs
	for _, iface := range new.Spec.Interfaces {
		pciPrefix := mlx.GetPciAddressPrefix(iface.PciAddress)
		if _, ok := mellanoxNicsStatus[pciPrefix]; !ok {
			continue
		}
		mellanoxNicsSpec[iface.PciAddress] = iface
	}

	if p.helpers.IsKernelLockdownMode() {
		if len(mellanoxNicsSpec) > 0 {
			log.Log.Info("Lockdown mode detected, failing on interface update for mellanox devices")
			return false, false, fmt.Errorf("mellanox device detected when in lockdown mode")
		}
		log.Log.Info("Lockdown mode detected, skpping mellanox nic processing")
		return
	}

	for _, ifaceSpec := range mellanoxNicsSpec {
		pciPrefix := mlx.GetPciAddressPrefix(ifaceSpec.PciAddress)
		// skip processed nics, help not running the same logic 2 times for dual port NICs
		if _, ok := processedNics[pciPrefix]; ok {
			continue
		}
		processedNics[pciPrefix] = true
		fwCurrent, fwNext, err := p.helpers.GetMlxNicFwData(ifaceSpec.PciAddress)
		if err != nil {
			return false, false, err
		}

		isDualPort := mlx.IsDualPort(ifaceSpec.PciAddress, mellanoxNicsStatus)
		// Attributes to change
		attrs := &mlx.MlxNic{TotalVfs: -1}
		var changeWithoutReboot bool

		totalVfs, totalVfsNeedReboot, totalVfsChangeWithoutReboot := mlx.HandleTotalVfs(fwCurrent, fwNext, attrs, ifaceSpec, isDualPort, mellanoxNicsSpec)
		sriovEnNeedReboot, sriovEnChangeWithoutReboot := mlx.HandleEnableSriov(totalVfs, fwCurrent, fwNext, attrs)
		// CTODO can we keep `needReboot` here?
		nicNeedReboot := totalVfsNeedReboot || sriovEnNeedReboot
		changeWithoutReboot = totalVfsChangeWithoutReboot || sriovEnChangeWithoutReboot

		needLinkChange, err := mlx.HandleLinkType(pciPrefix, fwCurrent, attrs, mellanoxNicsSpec, mellanoxNicsStatus)
		if err != nil {
			return false, false, err
		}
		nicNeedReboot = nicNeedReboot || needLinkChange

		pfNumOfVfValidNeedReboot, pfNumOfVfValidChangeWithoutReboot, err := p.handlePfNumOfVfValid(ifaceSpec.PciAddress, fwCurrent, fwNext, attrs)
		if err != nil {
			return false, false, err
		}
		nicNeedReboot = nicNeedReboot || pfNumOfVfValidNeedReboot
		changeWithoutReboot = changeWithoutReboot || pfNumOfVfValidChangeWithoutReboot

		// no FW changes allowed when NIC is externally managed
		if ifaceSpec.ExternallyManaged {
			if totalVfsNeedReboot || totalVfsChangeWithoutReboot {
				return false, false, fmt.Errorf(
					"interface %s required a change in the TotalVfs but the policy is externally managed failing: firmware TotalVf %d requested TotalVf %d",
					ifaceSpec.PciAddress, fwCurrent.TotalVfs, totalVfs)
			}
			if needLinkChange {
				return false, false, fmt.Errorf("change required for link type but the policy is externally managed, failing")
			}
		}

		if nicNeedReboot || changeWithoutReboot {
			attributesToChange[ifaceSpec.PciAddress] = *attrs
		}

		if nicNeedReboot {
			pciAddressesToReset = append(pciAddressesToReset, ifaceSpec.PciAddress)
		}
		needReboot = needReboot || nicNeedReboot
	}

	// Set total VFs to 0 for mellanox interfaces with no spec
	for pciPrefix, portsMap := range mellanoxNicsStatus {
		if _, ok := processedNics[pciPrefix]; ok {
			continue
		}

		// Add the nic to processed Nics to not repeat the process for dual nic ports
		processedNics[pciPrefix] = true
		pciAddress := pciPrefix + "0"

		// Skip devices not configured by the operator
		isConfigured, err := p.nicConfiguredByOperator(portsMap)
		if err != nil {
			return false, false, err
		}
		if !isConfigured {
			log.Log.V(2).Info("None of the ports are configured by the operator skipping firmware reset",
				"portMap", portsMap)
			continue
		}

		// Skip externally managed NICs
		hasExternally, err := p.nicHasExternallyManagedPFs(portsMap)
		if err != nil {
			return false, false, err
		}
		if hasExternally {
			log.Log.V(2).Info("One of the ports is configured as externally managed skipping firmware reset",
				"portMap", portsMap)
			continue
		}

		// Skip unsupported devices
		if id := sriovnetworkv1.GetVfDeviceID(portsMap[pciAddress].DeviceID); id == "" {
			continue
		}

		_, fwNext, err := p.helpers.GetMlxNicFwData(pciAddress)
		if err != nil {
			return false, false, err
		}

		if fwNext.TotalVfs > 0 || fwNext.EnableSriov {
			attributesToChange[pciAddress] = mlx.MlxNic{TotalVfs: 0}
			log.Log.V(2).Info("Changing TotalVfs to 0, doesn't require rebooting", "fwNext.totalVfs", fwNext.TotalVfs)
		}
	}

	if needReboot {
		needDrain = true
	}
	log.Log.V(2).Info("mellanox plugin", "need-drain", needDrain, "need-reboot", needReboot)
	return
}

// TODO: implement - https://github.com/k8snetworkplumbingwg/sriov-network-operator/issues/631
// OnNodeStatusChange verify whether SriovNetworkNodeState CR status present changes on configured VFs.
func (p *MellanoxPlugin) CheckStatusChanges(*sriovnetworkv1.SriovNetworkNodeState) (bool, error) {
	return false, nil
}

// Apply config change
func (p *MellanoxPlugin) Apply() error {
	if p.helpers.IsKernelLockdownMode() {
		log.Log.Info("mellanox plugin Apply() - skipping due to lockdown mode")
		return nil
	}
	log.Log.Info("mellanox plugin Apply()")
	if err := p.helpers.MlxConfigFW(attributesToChange); err != nil {
		return err
	}
	if vars.FeatureGate.IsEnabled(consts.MellanoxFirmwareResetFeatureGate) {
		return p.helpers.MlxResetFW(pciAddressesToReset, mellanoxNicsStatus)
	}
	return nil
}

// handlePfNumOfVfValid checks the PF_NUM_OF_VF_VALID override on every port of the NIC requested by the policies.
// The override is a per-PF setting, so it is evaluated and fixed for each port using the port PCI address.
// Attributes for the port with PCI address processedPciAddress are stored in attrs, for the other port they are
// directly added to attributesToChange.
func (p *MellanoxPlugin) handlePfNumOfVfValid(processedPciAddress string, fwCurrent, fwNext, attrs *mlx.MlxNic) (
	needReboot, changeWithoutReboot bool, err error) {
	pciPrefix := mlx.GetPciAddressPrefix(processedPciAddress)
	for _, pciAddress := range []string{pciPrefix + "0", pciPrefix + "1"} {
		portSpec, ok := mellanoxNicsSpec[pciAddress]
		if !ok {
			continue
		}

		portFwCurrent, portFwNext, portAttrs := fwCurrent, fwNext, attrs
		if pciAddress != processedPciAddress {
			portFwCurrent, portFwNext, err = p.helpers.GetMlxNicFwData(pciAddress)
			if err != nil {
				return false, false, err
			}
			portAttrs = &mlx.MlxNic{TotalVfs: -1}
		}

		// no FW changes allowed when NIC is externally managed, fail only if the override prevents creating the VFs
		if portSpec.ExternallyManaged {
			if portFwCurrent.IsPfNumOfVfOverrideEnabled() && portFwCurrent.PfNumOfVf < portSpec.NumVfs {
				return false, false, fmt.Errorf(
					"interface %s has %s enabled in the firmware with %s %d lower than the requested %d VFs "+
						"but the policy is externally managed failing: set %s=False or %s>=%d in the firmware",
					pciAddress, mlx.PfNumOfVfValid, mlx.PfNumOfVf, portFwCurrent.PfNumOfVf, portSpec.NumVfs,
					mlx.PfNumOfVfValid, mlx.PfNumOfVf, portSpec.NumVfs)
			}
			continue
		}

		portNeedReboot, portChangeWithoutReboot := mlx.HandlePfNumOfVfValid(portFwCurrent, portFwNext, portAttrs)
		if portNeedReboot || portChangeWithoutReboot {
			log.Log.Info("mellanox-plugin: PF_NUM_OF_VF_VALID is enabled and overrides NUM_OF_VFS for the PF, it must be disabled",
				"pciAddress", pciAddress, "need-reboot", portNeedReboot)
		}
		if pciAddress != processedPciAddress && portAttrs.PfNumOfVfValid != nil {
			attributesToChange[pciAddress] = *portAttrs
		}
		needReboot = needReboot || portNeedReboot
		changeWithoutReboot = changeWithoutReboot || portChangeWithoutReboot
	}
	return needReboot, changeWithoutReboot, nil
}

// nicHasExternallyManagedPFs returns true if one of the ports(interface) of the NIC is marked as externally managed
// in StoreManagerInterface.
func (p *MellanoxPlugin) nicHasExternallyManagedPFs(nicPortsMap map[string]sriovnetworkv1.InterfaceExt) (bool, error) {
	for _, iface := range nicPortsMap {
		pfStatus, exist, err := p.helpers.LoadPfsStatus(iface.PciAddress)
		if err != nil {
			// nolint:goconst
			log.Log.Error(err, "failed to load PF status from disk. "+
				"This should not happen, to overcome config daemon stuck, "+
				"please remove the PCI file on the host under the operator configuration path",
				"path", consts.PfAppliedConfig, "pciAddress", iface.PciAddress)
			return false, err
		}
		if !exist {
			continue
		}
		if pfStatus.ExternallyManaged {
			log.Log.V(2).Info("PF is extenally managed, skip FW TotalVfs reset")
			return true, nil
		}
	}
	return false, nil
}

// nicConfiguredByOperator returns true if one of the ports(interface) of the NIC is configured by operator
func (p *MellanoxPlugin) nicConfiguredByOperator(nicPortsMap map[string]sriovnetworkv1.InterfaceExt) (bool, error) {
	for _, iface := range nicPortsMap {
		_, exist, err := p.helpers.LoadPfsStatus(iface.PciAddress)
		if err != nil {
			// nolint:goconst
			log.Log.Error(err, "failed to load PF status from disk. "+
				"This should not happen, to overcome config daemon stuck, "+
				"please remove the PCI file on the host under the operator configuration path",
				"path", consts.PfAppliedConfig, "pciAddress", iface.PciAddress)
			return false, err
		}
		if exist {
			log.Log.V(2).Info("PF configured by the operator", "interface", iface)
			return true, nil
		}
	}

	return false, nil
}
