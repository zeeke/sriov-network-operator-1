package generic

import (
	"testing"

	sriovnetworkv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/stretchr/testify/assert"
)

func TestNeedDrainNode_NoNeedToDrain(t *testing.T) {
	desired := sriovnetworkv1.Interfaces{{
		PciAddress: "0000:00:00.0",
		NumVfs:     1,
		VfGroups: []sriovnetworkv1.VfGroup{{
			DeviceType:   "netdevice",
			PolicyName:   "policy-1",
			ResourceName: "resource-1",
			VfRange:      "0-0",
		}},
	}}

	current := sriovnetworkv1.InterfaceExts{{
		PciAddress:  "0000:00:00.0",
		NumVfs:      1,
		TotalVfs:    1,
		DeviceID:    "1015",
		Vendor:      "15b3",
		Name:        "sriovif1",
		Mtu:         1500,
		Mac:         "0c:42:a1:55:ee:46",
		Driver:      "mlx5_core",
		EswitchMode: "legacy",
		LinkSpeed:   "25000 Mb/s",
		LinkType:    "ETH",
		VFs: []sriovnetworkv1.VirtualFunction{{
			PciAddress: "0000:00:00.1",
			DeviceID:   "1016",
			Vendor:     "15b3",
			VfID:       0,
			Name:       "sriovif1v0",
			Mtu:        1500,
			Mac:        "8e:d6:2c:62:87:1b",
			Driver:     "mlx5_core",
		}},
	}}

	assert.False(t, needDrainNode(desired, current))
}
