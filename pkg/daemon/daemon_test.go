package daemon

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"

	sriovnetworkv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vishvananda/netlink"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	snclientset "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/client/clientset/versioned"
	fakesnclientset "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/client/clientset/versioned/fake"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/fswrap"
	plugin "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/plugins"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/plugins/generic"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/utils"

	fakemcclientset "github.com/openshift/machine-config-operator/pkg/generated/clientset/versioned/fake"

	fakek8s "k8s.io/client-go/kubernetes/fake"

	// hack: this import is needed to access sysBuPci unexported variable
	_ "unsafe"
)

//go:linkname sysBusPci github.com/k8snetworkplumbingwg/sriov-network-device-plugin/pkg/utils.sysBusPci
var sysBusPci string

var FakeSupportedNicIDsCM corev1.ConfigMap = corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name:      sriovnetworkv1.SupportedNicIDConfigmap,
		Namespace: namespace,
	},
	Data: map[string]string{},
}

var SriovDevicePluginPod corev1.Pod = corev1.Pod{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "sriov-device-plugin-xxxx",
		Namespace: namespace,
		Labels: map[string]string{
			"app": "sriov-device-plugin",
		},
	},
	Spec: corev1.PodSpec{
		NodeName: "test-node",
	},
}

func init() {
	// Increase verbosity to help debugging failures
	flag.Set("logtostderr", "true")
	flag.Set("stderrthreshold", "WARNING")
	flag.Set("v", "2")
}

func TestConfigDaemon(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("Config Daemon tests requires root privileges as they leverage netlink functionalities.")
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Daemon Suite")
}

var _ = Describe("Config Daemon", func() {
	var stopCh chan struct{}
	var syncCh chan struct{}
	var exitCh chan error
	var refreshCh chan Message

	var cleanFakeFs func()

	var sut *Daemon

	var linkDummy = netlink.Dummy{
		LinkAttrs: netlink.LinkAttrs{
			Name: "dummysriov0",
		},
	}

	BeforeEach(func() {

		stopCh = make(chan struct{})
		refreshCh = make(chan Message, 1024)
		exitCh = make(chan error)
		syncCh = make(chan struct{}, 64)

		// Fill syncCh with values so daemon doesn't wait for a writer
		for i := 0; i < 64; i++ {
			syncCh <- struct{}{}
		}

		// Create virtual filesystem for Daemon
		fakeFs := &FakeFilesystem{
			Dirs: []string{
				"bindata/scripts",
				"host/sys/bus/pci/devices/0000:86:00.0",
			},
			Symlinks: map[string]string{},
			Files: map[string][]byte{
				"/bindata/scripts/enable-rdma.sh":                     []byte(""),
				"/bindata/scripts/load-kmod.sh":                       []byte(""),
				"/host/sys/bus/pci/devices/0000:86:00.0/sriov_numvfs": []byte(""),
			},
		}

		var err error
		cleanFakeFs, err = fakeFs.Use()
		Expect(err).ToNot(HaveOccurred())

		kubeClient := fakek8s.NewSimpleClientset(&FakeSupportedNicIDsCM, &SriovDevicePluginPod)
		kubeClient.Resources = []*metav1.APIResourceList{{GroupVersion: "v1"}}
		client := fakesnclientset.NewSimpleClientset()
		mcClient := fakemcclientset.NewSimpleClientset()

		sut = New("test-node",
			client,
			kubeClient,
			mcClient,
			exitCh,
			stopCh,
			syncCh,
			refreshCh,
			utils.Baremetal,
		)

		p, _ := generic.NewGenericPlugin()
		sut.enabledPlugins = map[string]plugin.VendorPlugin{generic.PluginName: p}

		err = netlink.LinkAdd(&linkDummy)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		close(stopCh)
		close(syncCh)
		close(exitCh)
		close(refreshCh)

		cleanFakeFs()

		err := netlink.LinkDel(&linkDummy)
		Expect(err).ToNot(HaveOccurred())
	})

	Context("Should", func() {

		It("configure a device with 4 VFs", func() {
			go func() {
				Expect(sut.Run(stopCh, exitCh)).To(BeNil())
			}()

			_, err := sut.kubeClient.CoreV1().Nodes().
				Create(context.Background(), &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{Name: "test-node"},
				}, metav1.CreateOptions{})
			Expect(err).To(BeNil())

			nodeState := &sriovnetworkv1.SriovNetworkNodeState{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-node",
					Generation: 123,
				},
				Spec: sriovnetworkv1.SriovNetworkNodeStateSpec{
					Interfaces: []sriovnetworkv1.Interface{
						{
							PciAddress: "0000:86:00.0",
							Name:       "dummysriov0",
							NumVfs:     4,
							VfGroups: []sriovnetworkv1.VfGroup{{
								DeviceType:   "netdevice",
								PolicyName:   "policy1",
								ResourceName: "resource1",
								VfRange:      "0-3",
							}},
						},
					},
				},
				Status: sriovnetworkv1.SriovNetworkNodeStateStatus{
					Interfaces: []sriovnetworkv1.InterfaceExt{
						{
							VFs:        []sriovnetworkv1.VirtualFunction{{}},
							DeviceID:   "158b",
							Driver:     "i40e",
							Mtu:        1500,
							Name:       "dummysriov0",
							PciAddress: "0000:86:00.0",
							Vendor:     "8086",
							NumVfs:     0,
							TotalVfs:   64,
						},
					},
				},
			}
			Expect(
				createSriovNetworkNodeState(sut.client, nodeState)).
				To(BeNil())

			var msg Message
			Eventually(refreshCh, "10s").Should(Receive(&msg))
			Expect(msg.syncStatus).To(Equal("InProgress"), msg.lastSyncError)

			Eventually(refreshCh, "10s").Should(Receive(&msg))
			Expect(msg.syncStatus).To(Equal("Succeeded"), msg.lastSyncError)

			Eventually(func() (string, error) {
				bytes, err := fswrap.ReadFile("/host/sys/bus/pci/devices/0000:86:00.0/sriov_numvfs")
				return string(bytes), err
			}).Should(Equal("4"))
		})

		It("ignore non latest SriovNetworkNodeState generations", func() {
			go func() {
				Expect(sut.Run(stopCh, exitCh)).To(BeNil())
			}()

			_, err := sut.kubeClient.CoreV1().Nodes().Create(context.Background(), &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
				},
			}, metav1.CreateOptions{})
			Expect(err).To(BeNil())

			nodeState1 := &sriovnetworkv1.SriovNetworkNodeState{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-node",
					Generation: 123,
				},
			}
			Expect(
				createSriovNetworkNodeState(sut.client, nodeState1)).
				To(BeNil())

			nodeState2 := &sriovnetworkv1.SriovNetworkNodeState{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-node",
					Generation: 777,
				},
			}
			Expect(
				updateSriovNetworkNodeState(sut.client, nodeState2)).
				To(BeNil())

			var msg Message
			Eventually(refreshCh, "10s").Should(Receive(&msg))
			Expect(msg.syncStatus).To(Equal("InProgress"))

			Eventually(refreshCh, "10s").Should(Receive(&msg))
			Expect(msg.syncStatus).To(Equal("Succeeded"))

			Expect(sut.nodeState.GetGeneration()).To(BeNumerically("==", 777))
		})
	})
})

func createSriovNetworkNodeState(c snclientset.Interface, nodeState *sriovnetworkv1.SriovNetworkNodeState) error {
	_, err := c.SriovnetworkV1().
		SriovNetworkNodeStates(namespace).
		Create(context.Background(), nodeState, metav1.CreateOptions{})
	return err
}

func updateSriovNetworkNodeState(c snclientset.Interface, nodeState *sriovnetworkv1.SriovNetworkNodeState) error {
	_, err := c.SriovnetworkV1().
		SriovNetworkNodeStates(namespace).
		Update(context.Background(), nodeState, metav1.UpdateOptions{})
	return err
}

// FakeFilesystem allows to setup isolated fake files structure used for the tests.
type FakeFilesystem struct {
	Dirs     []string
	Files    map[string][]byte
	Symlinks map[string]string
	Commands map[string]string

	// contains the list of Chroot operations
	chrootStack []string
}

// Use function creates entire files structure and returns a function to tear it down. Example usage: defer fs.Use()()
func (f *FakeFilesystem) Use() (func(), error) {
	// create the new fake fs root dir in /tmp/sriov...
	rootDir, err := ioutil.TempDir("", "sriov")
	if err != nil {
		return nil, fmt.Errorf("error creating fake root dir: %w", err)
	}

	sysBusPci = rootDir + "/sys/bus/pci/devices"
	f.chrootStack = []string{rootDir}

	for _, dir := range f.Dirs {
		//nolint: gomnd
		err := os.MkdirAll(path.Join(rootDir, dir), 0755)
		if err != nil {
			return nil, fmt.Errorf("error creating fake directory: %w", err)
		}
	}

	for filename, body := range f.Files {
		//nolint: gomnd
		err := os.WriteFile(path.Join(rootDir, filename), body, 0600)
		if err != nil {
			return nil, fmt.Errorf("error creating fake file: %w", err)
		}
	}

	//nolint: gomnd
	err = os.MkdirAll(path.Join(rootDir, "usr/share/hwdata"), 0755)
	if err != nil {
		return nil, fmt.Errorf("error creating fake directory: %w", err)
	}

	for link, target := range f.Symlinks {
		err = os.Symlink(target, path.Join(rootDir, link))
		if err != nil {
			return nil, fmt.Errorf("error creating fake symlink: %w", err)
		}
	}

	// Mock filesystem files
	oldReadFile := fswrap.ReadFile
	fswrap.ReadFile = func(filename string) ([]byte, error) {
		return ioutil.ReadFile(f.chrootStack[0] + "/" + filename)
	}

	oldWriteFile := fswrap.WriteFile
	fswrap.WriteFile = func(filename string, data []byte, perm fs.FileMode) error {
		return ioutil.WriteFile(f.chrootStack[0]+"/"+filename, data, perm)
	}

	oldMkdirAll := fswrap.MkdirAll
	fswrap.MkdirAll = func(path string, perm os.FileMode) error {
		return os.MkdirAll(f.chrootStack[0]+"/"+path, perm)
	}

	oldChroot := fswrap.Chroot
	fswrap.Chroot = func(path string) (func() error, error) {
		// Push the new root to the head
		f.chrootStack = append([]string{f.chrootStack[0] + "/" + path}, f.chrootStack...)
		sysBusPci = f.chrootStack[0] + "/sys/bus/pci/devices"
		return func() error {
			if len(f.chrootStack) <= 1 {
				return fmt.Errorf("can't exit from chroot: %v", f.chrootStack)
			}

			f.chrootStack = f.chrootStack[1:]
			sysBusPci = f.chrootStack[0] + "/sys/bus/pci/devices"
			return nil
		}, nil
	}

	oldCommand := fswrap.Command
	fswrap.Command = func(name string, arg ...string) *exec.Cmd {
		mangledCmd, ok := f.Commands[name+strings.Join(arg, " ")]
		if !ok {
			return exec.Command(name, arg...)
		}

		cmdChunks := strings.Split(mangledCmd, " ")
		return exec.Command(cmdChunks[0], cmdChunks...)
	}

	return func() {
		// Restore wrap functions
		fswrap.ReadFile = oldReadFile
		fswrap.WriteFile = oldWriteFile
		fswrap.MkdirAll = oldMkdirAll
		fswrap.Chroot = oldChroot
		fswrap.Command = oldCommand

		// remove temporary fake fs
		err := os.RemoveAll(rootDir)
		if err != nil {
			panic(fmt.Errorf("error tearing down fake filesystem: %w", err))
		}
	}, nil
}
