package utils

import (
	"testing"

	"github.com/jaypipes/ghw"
	"github.com/jaypipes/ghw/pkg/net"
	"github.com/jaypipes/ghw/pkg/option"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/pointer"
)

func TestGetOpenstackData_RealPCIA(t *testing.T) {
	ospNetworkDataFile = "./testdata/network_data.json"
	ospMetaDataFile = "./testdata/meta_data.json"
	defer func() {
		ospNetworkDataFile = ospMetaDataDir + "/network_data.json"
		ospMetaDataFile = ospMetaDataDir + "/meta_data.json"
	}()

	ghw.Network = func(opts ...*option.Option) (*net.Info, error) {
		return &net.Info{
			NICs: []*net.NIC{{
				MacAddress: "fa:16:3e:00:00:00",
				PCIAddress: pointer.String("0000:04:00.0"),
			}, {
				MacAddress: "fa:16:3e:11:11:11",
				PCIAddress: pointer.String("0000:99:99.9"),
			}},
		}, nil
	}
	defer func() {
		ghw.Network = net.New
	}()

	metaData, _, err := GetOpenstackData(false)
	assert.NoError(t, err)

	assert.Equal(t, "fa:16:3e:00:00:00", metaData.Devices[0].Mac)
	assert.Equal(t, "0000:04:00.0", metaData.Devices[0].Address)

	assert.Equal(t, "fa:16:3e:11:11:11", metaData.Devices[1].Mac)
	assert.Equal(t, "0000:99:99.9", metaData.Devices[1].Address)
}
