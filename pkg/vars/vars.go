package vars

import (
	"os"
	"regexp"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"

	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/consts"
)

var (
	// ClusterType used by the operator to specify the platform it's running on
	// supported values [kubernetes,openshift]
	ClusterType string

	// DevMode controls the developer mode in the operator
	// developer mode allows the operator to use un-supported network devices
	DevMode bool

	// EnableAdmissionController allows the user to disable the operator webhooks
	EnableAdmissionController bool

	// NodeName initialize and used by the config-daemon to identify the node it's running on
	NodeName = ""

	// Destdir destination directory for the checkPoint file on the host
	Destdir string

	// PlatformType specify the current platform the operator is running on
	PlatformType = consts.Baremetal
	// PlatformsMap contains supported platforms for virtual VF
	PlatformsMap = map[string]consts.PlatformTypes{
		"openstack": consts.VirtualOpenStack,
	}

	// SupportedVfIds list of supported virtual functions IDs
	// loaded on daemon initialization by reading the supported-nics configmap
	SupportedVfIds []string

	// DpdkDrivers supported DPDK drivers for virtual functions
	DpdkDrivers = []string{"igb_uio", "vfio-pci", "uio_pci_generic"}

	// InChroot global variable to mark that the config-daemon code is inside chroot on the host file system
	InChroot = false

	// UsingSystemdMode global variable to mark the config-daemon is running on systemd mode
	UsingSystemdMode = false

	// FilesystemRoot used by test to mock interactions with filesystem
	FilesystemRoot = ""

	//Cluster variables
	Config *rest.Config    = nil
	Scheme *runtime.Scheme = nil

	// PfPhysPortNameRe regex to find switchdev devices on the host
	PfPhysPortNameRe = regexp.MustCompile(`p\d+`)
)

func init() {
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

	EnableAdmissionController = false
	enableAdmissionController := os.Getenv("ADMISSION_CONTROLLERS_ENABLED")
	if enableAdmissionController == "True" {
		EnableAdmissionController = true
	}
}
