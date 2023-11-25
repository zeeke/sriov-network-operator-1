package platforms

import (
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/platforms/openshift"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/platforms/openstack"
)

//go:generate ../../bin/mockgen -destination mock/mock_platforms.go -source platforms.go
type PlatformHelperInterface interface {
	openshift.OpenshiftContextInterface
	openstack.OpenstackInterface
}

type PlatformHelper struct {
	openshift.OpenshiftContextInterface
	openstack.OpenstackInterface
}

func NewDefaultPlatformHelper() (*PlatformHelper, error) {
	openshiftContext, err := openshift.NewOpenshiftContext()
	if err != nil {
		return nil, err
	}

	openstackContext := openstack.NewOpenstackContext()

	return &PlatformHelper{
		openshiftContext,
		openstackContext,
	}, nil
}
