package vars

import (
	"os"
	"regexp"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"

	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/global/consts"
)

var (
	ClusterType string

	// PlatformsMap contains supported platforms for virtual VF
	PlatformsMap map[string]consts.PlatformTypes

	DevMode bool

	InChroot         bool
	UsingSystemdMode bool

	Destdir string

	NodeName string

	PlatformType consts.PlatformTypes

	// FilesystemRoot used by test to mock interactions with filesystem
	FilesystemRoot string

	PfPhysPortNameRe = regexp.MustCompile(`p\d+`)

	SupportedVfIds []string

	DpdkDrivers = []string{"igb_uio", "vfio-pci", "uio_pci_generic"}

	//Cluster variables
	Config *rest.Config
	Scheme *runtime.Scheme
)

func init() {
	NodeName = ""
	InChroot = false
	UsingSystemdMode = false
	Config = nil
	Scheme = nil

	ClusterType = os.Getenv("CLUSTER_TYPE")

	DevMode = false
	mode := os.Getenv("DEV_MODE")
	if mode == "TRUE" {
		DevMode = true
	}

	Destdir = "/host/tmp"
	destdir := os.Getenv("DEST_DIR")
	if destdir != "" {
		Destdir = destdir
	}

	FilesystemRoot = ""

	PlatformsMap = map[string]consts.PlatformTypes{
		"openstack": consts.VirtualOpenStack,
	}

	PlatformType = consts.Baremetal
}
