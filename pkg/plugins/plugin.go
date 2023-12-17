package plugin

import (
	"sigs.k8s.io/controller-runtime/pkg/log"

	sriovnetworkv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/utils"
	mlx "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/vendors/mellanox"
)

//go:generate ../../bin/mockgen -destination mock/mock_plugin.go -source plugin.go
type VendorPlugin interface {
	// Return the name of plugin
	Name() string
	// Return the SpecVersion followed by plugin
	Spec() string
	// Invoked when SriovNetworkNodeState CR is created or updated, return if need dain and/or reboot node
	OnNodeStateChange(new *sriovnetworkv1.SriovNetworkNodeState) (bool, bool, error)
	// Apply config change
	Apply() error
}

type HostHelpersInterface interface {
	utils.UtilsInterface
	host.HostManagerInterface
	host.StoreManagerInterface
	mlx.MellanoxInterface
}

type HostHelpers struct {
	utils.UtilsInterface
	host.HostManagerInterface
	host.StoreManagerInterface
	mlx.MellanoxInterface
}

// Use for unit tests
func NewVendorPluginHelpers(utilsHelper utils.UtilsInterface,
	hostManager host.HostManagerInterface,
	storeManager host.StoreManagerInterface,
	mlxHelper mlx.MellanoxInterface) *HostHelpers {
	return &HostHelpers{utilsHelper, hostManager, storeManager, mlxHelper}
}

func NewDefaultVendorPluginHelpers() (*HostHelpers, error) {
	utilsHelper := utils.NewUtilsHelper()
	mlxHelper := mlx.New(utilsHelper)
	hostManager := host.NewHostManager(utilsHelper)
	storeManager, err := host.NewStoreManager()
	if err != nil {
		log.Log.Error(err, "failed to create store manager")
		return nil, err
	}

	return &HostHelpers{
		utilsHelper,
		hostManager,
		storeManager,
		mlxHelper}, nil
}
