package openshift

import (
	mcclientset "github.com/openshift/machine-config-operator/pkg/generated/clientset/versioned"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/consts"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/utils"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/vars"
)

// OpenshiftFlavor holds metadata about the type of Openshift environment the operator is in.
type OpenshiftFlavor string

const (
	// Hypershift flavor of openshift: https://github.com/openshift/hypershift
	OpenshiftFlavorHypershift OpenshiftFlavor = "hypershift"
	// OpenshiftFlavorDefault covers all remaining flavors of openshift not explicitly called out above
	OpenshiftFlavorDefault OpenshiftFlavor = "default"
)

//go:generate ../../../bin/mockgen -destination mock/mock_openshift.go -source openshift.go
type OpenshiftContextInterface interface {
	GetFlavor() OpenshiftFlavor
	GetMcClient() mcclientset.Interface
	IsOpenshiftCluster() bool
	IsHypershift() bool
}

// OpenshiftContext contains metadata and structs utilized to interact with Openshift clusters
type OpenshiftContext struct {
	// McClient is a client for MachineConfigs in an Openshift environment
	McClient mcclientset.Interface

	// IsOpenShiftCluster boolean to point out if the cluster is an OpenShift cluster
	IsOpenShiftCluster bool

	// OpenshiftFlavor holds metadata about the type of Openshift environment the operator is in.
	OpenshiftFlavor OpenshiftFlavor
}

func NewOpenshiftContext() (OpenshiftContextInterface, error) {
	if vars.ClusterType != consts.ClusterTypeOpenshift {
		return &OpenshiftContext{nil, false, ""}, nil
	}

	mcclient, err := mcclientset.NewForConfig(vars.Config)
	if err != nil {
		return nil, err
	}

	openshiftFlavor := OpenshiftFlavorDefault
	infraClient, err := client.New(vars.Config, client.Options{
		Scheme: vars.Scheme,
	})
	if err != nil {
		return nil, err
	}

	isHypershift, err := utils.IsExternalControlPlaneCluster(infraClient)
	if err != nil {
		return nil, err
	}

	if isHypershift {
		openshiftFlavor = OpenshiftFlavorHypershift
	}

	return &OpenshiftContext{mcclient, true, openshiftFlavor}, nil
}

func (c *OpenshiftContext) GetFlavor() OpenshiftFlavor {
	return c.OpenshiftFlavor
}

func (c *OpenshiftContext) GetMcClient() mcclientset.Interface {
	return c.McClient
}

func (c OpenshiftContext) IsOpenshiftCluster() bool {
	return c.IsOpenShiftCluster
}

func (c OpenshiftContext) IsHypershift() bool {
	return c.OpenshiftFlavor == OpenshiftFlavorHypershift
}
