package mellanox

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/mock/gomock"
	corev1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	sriovnetworkv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/consts"
	mock_helper "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/helper/mock"
	plugin "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/plugins"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/vars"
	mlx "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/vendors/mellanox"
)

var _ = Describe("SRIOV", Ordered, func() {
	var (
		m        plugin.VendorPlugin
		h        *mock_helper.MockHostHelpersInterface
		err      error
		testCtrl *gomock.Controller

		sriovNetworkNodeState *sriovnetworkv1.SriovNetworkNodeState
	)

	BeforeAll(func() {
		sriovnetworkv1.NicIDMap = []string{"15b3 1013 1014"}
	})

	BeforeEach(func() {
		testCtrl = gomock.NewController(GinkgoT())
		h = mock_helper.NewMockHostHelpersInterface(testCtrl)
		m, err = NewMellanoxPlugin(h)
		Expect(err).ToNot(HaveOccurred())

		sriovNetworkNodeState = &sriovnetworkv1.SriovNetworkNodeState{ObjectMeta: corev1.ObjectMeta{Name: "worker-0", Namespace: "test"},
			Spec:   sriovnetworkv1.SriovNetworkNodeStateSpec{},
			Status: sriovnetworkv1.SriovNetworkNodeStateStatus{},
		}
	})

	AfterEach(func() {
		testCtrl.Finish()
	})

	Context("CheckStatusChanges", func() {
		It("should always return false and nil", func() {
			drift, err := m.CheckStatusChanges(nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(drift).To(BeFalse())
		})
	})

	Context("OnNodeStateChange", func() {
		It("should return error if we are trying to configure mlx devices and the system is in lock down mode", func() {
			h.EXPECT().IsKernelLockdownMode().Return(true)
			sriovNetworkNodeState.Spec.Interfaces = sriovnetworkv1.Interfaces{
				{Name: "eno1",
					NumVfs:     10,
					PciAddress: "0000:d8:00.0", VfGroups: []sriovnetworkv1.VfGroup{
						{ResourceName: "test",
							PolicyName: "test",
							VfRange:    "0-0"},
					},
				},
			}
			sriovNetworkNodeState.Status.Interfaces = sriovnetworkv1.InterfaceExts{
				{
					Name:       "eno1",
					NumVfs:     0,
					PciAddress: "0000:d8:00.0",
					Vendor:     "15b3",
				},
			}

			_, _, err := m.OnNodeStateChange(sriovNetworkNodeState)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("mellanox device detected when in lockdown mode"))
		})

		It("should not return error the system is in lock down mode but we don't configure any mlx device", func() {
			h.EXPECT().IsKernelLockdownMode().Return(true)
			sriovNetworkNodeState.Spec.Interfaces = sriovnetworkv1.Interfaces{}
			sriovNetworkNodeState.Status.Interfaces = sriovnetworkv1.InterfaceExts{
				{
					Name:       "eno1",
					NumVfs:     0,
					PciAddress: "0000:d8:00.0",
					Vendor:     "15b3",
				},
			}

			needDrain, needReboot, err := m.OnNodeStateChange(sriovNetworkNodeState)
			Expect(err).ToNot(HaveOccurred())
			Expect(needDrain).To(BeFalse())
			Expect(needReboot).To(BeFalse())
		})

		It("should return error if the nic require fw changes but the nic is externally manage", func() {
			h.EXPECT().IsKernelLockdownMode().Return(false)
			h.EXPECT().GetMlxNicFwData("0000:d8:00.0").Return(&mlx.MlxNic{TotalVfs: 0}, &mlx.MlxNic{TotalVfs: 0}, nil)
			sriovNetworkNodeState.Spec.Interfaces = sriovnetworkv1.Interfaces{
				{Name: "eno1",
					NumVfs:            10,
					ExternallyManaged: true,
					PciAddress:        "0000:d8:00.0", VfGroups: []sriovnetworkv1.VfGroup{
						{ResourceName: "test",
							PolicyName: "test",
							VfRange:    "eno1#0-9"},
					},
				},
			}
			sriovNetworkNodeState.Status.Interfaces = sriovnetworkv1.InterfaceExts{
				{
					Name:       "eno1",
					NumVfs:     0,
					PciAddress: "0000:d8:00.0",
					Vendor:     "15b3",
				},
			}

			_, _, err := m.OnNodeStateChange(sriovNetworkNodeState)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("interface 0000:d8:00.0 required a change in the TotalVfs but the policy is externally managed failing: firmware TotalVf 0 requested TotalVf 10"))
		})

		It("should return true on reboot if we need to update the number of vfs in the firmware", func() {
			h.EXPECT().IsKernelLockdownMode().Return(false)
			h.EXPECT().GetMlxNicFwData("0000:d8:00.0").Return(&mlx.MlxNic{TotalVfs: 0}, &mlx.MlxNic{TotalVfs: 0}, nil)
			sriovNetworkNodeState.Spec.Interfaces = sriovnetworkv1.Interfaces{
				{Name: "eno1",
					NumVfs:     10,
					PciAddress: "0000:d8:00.0", VfGroups: []sriovnetworkv1.VfGroup{
						{ResourceName: "test",
							PolicyName: "test",
							VfRange:    "eno1#0-9"},
					},
				},
			}
			sriovNetworkNodeState.Status.Interfaces = sriovnetworkv1.InterfaceExts{
				{
					Name:       "eno1",
					NumVfs:     0,
					PciAddress: "0000:d8:00.0",
					Vendor:     "15b3",
				},
			}

			needDrain, needReboot, err := m.OnNodeStateChange(sriovNetworkNodeState)
			Expect(err).ToNot(HaveOccurred())
			Expect(needDrain).To(BeTrue())
			Expect(needReboot).To(BeTrue())
		})

		It("should return true on reboot adding vfs for one PF and removing for the other", func() {
			h.EXPECT().IsKernelLockdownMode().Return(false)
			h.EXPECT().GetMlxNicFwData("0000:d8:00.0").Return(&mlx.MlxNic{TotalVfs: 0}, &mlx.MlxNic{TotalVfs: 0}, nil)
			h.EXPECT().GetMlxNicFwData("0000:d9:00.0").Return(&mlx.MlxNic{TotalVfs: 10}, &mlx.MlxNic{TotalVfs: 10}, nil)
			h.EXPECT().LoadPfsStatus("0000:d9:00.0").Return(&sriovnetworkv1.Interface{ExternallyManaged: false}, true, nil).AnyTimes()
			sriovNetworkNodeState.Spec.Interfaces = sriovnetworkv1.Interfaces{
				{Name: "eno1",
					NumVfs:     10,
					PciAddress: "0000:d8:00.0", VfGroups: []sriovnetworkv1.VfGroup{
						{ResourceName: "test",
							PolicyName: "test",
							VfRange:    "eno1#0-9"},
					},
				},
			}
			sriovNetworkNodeState.Status.Interfaces = sriovnetworkv1.InterfaceExts{
				{
					Name:       "eno1",
					NumVfs:     0,
					PciAddress: "0000:d8:00.0",
					Vendor:     "15b3",
					DeviceID:   "1013",
				},
				{
					Name:       "eno2",
					NumVfs:     10,
					PciAddress: "0000:d9:00.0",
					Vendor:     "15b3",
					DeviceID:   "1013",
				},
			}

			needDrain, needReboot, err := m.OnNodeStateChange(sriovNetworkNodeState)
			Expect(err).ToNot(HaveOccurred())
			Expect(needDrain).To(BeTrue())
			Expect(needReboot).To(BeTrue())
			value, exist := attributesToChange["0000:d8:00.0"]
			Expect(exist).To(BeTrue())
			Expect(value.TotalVfs).To(Equal(10))
			Expect(value.EnableSriov).To(BeTrue())
			value, exist = attributesToChange["0000:d9:00.0"]
			Expect(exist).To(BeTrue())
			Expect(value.TotalVfs).To(Equal(0))
			Expect(value.EnableSriov).To(BeFalse())
		})

		It("should return false if we just need to reset the vfs", func() {
			h.EXPECT().IsKernelLockdownMode().Return(false)
			h.EXPECT().GetMlxNicFwData("0000:d8:00.0").Return(&mlx.MlxNic{TotalVfs: 10}, &mlx.MlxNic{TotalVfs: 10}, nil)
			h.EXPECT().LoadPfsStatus("0000:d8:00.0").Return(&sriovnetworkv1.Interface{ExternallyManaged: false}, true, nil).AnyTimes()
			sriovNetworkNodeState.Spec.Interfaces = sriovnetworkv1.Interfaces{}
			sriovNetworkNodeState.Status.Interfaces = sriovnetworkv1.InterfaceExts{
				{
					Name:       "eno1",
					NumVfs:     10,
					PciAddress: "0000:d8:00.0",
					Vendor:     "15b3",
					DeviceID:   "1013",
				},
			}

			needDrain, needReboot, err := m.OnNodeStateChange(sriovNetworkNodeState)
			Expect(err).ToNot(HaveOccurred())
			Expect(needDrain).To(BeFalse())
			Expect(needReboot).To(BeFalse())
			value, exist := attributesToChange["0000:d8:00.0"]
			Expect(exist).To(BeTrue())
			Expect(value.TotalVfs).To(Equal(0))
			Expect(value.EnableSriov).To(BeFalse())

		})

		It("should return error if we are not able to read the PF status file form host", func() {
			h.EXPECT().IsKernelLockdownMode().Return(false)
			h.EXPECT().LoadPfsStatus("0000:d8:00.0").Return(nil, false, fmt.Errorf("failed to read file"))
			sriovNetworkNodeState.Spec.Interfaces = sriovnetworkv1.Interfaces{}
			sriovNetworkNodeState.Status.Interfaces = sriovnetworkv1.InterfaceExts{
				{
					Name:       "eno1",
					NumVfs:     10,
					PciAddress: "0000:d8:00.0",
					Vendor:     "15b3",
				},
			}

			needDrain, needReboot, err := m.OnNodeStateChange(sriovNetworkNodeState)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("failed to read file"))
			Expect(needDrain).To(BeFalse())
			Expect(needReboot).To(BeFalse())
		})

		It("should failed if policy is externally manage and we need to change nic type", func() {
			h.EXPECT().IsKernelLockdownMode().Return(false)
			h.EXPECT().GetMlxNicFwData("0000:d8:00.0").Return(&mlx.MlxNic{TotalVfs: 10, LinkTypeP1: "ETH"}, &mlx.MlxNic{TotalVfs: 10, LinkTypeP1: "ETH"}, nil)
			sriovNetworkNodeState.Spec.Interfaces = sriovnetworkv1.Interfaces{
				{Name: "eno1",
					NumVfs:            10,
					LinkType:          "IB",
					ExternallyManaged: true,
					PciAddress:        "0000:d8:00.0", VfGroups: []sriovnetworkv1.VfGroup{
						{ResourceName: "test",
							PolicyName: "test",
							VfRange:    "eno1#0-9"},
					},
				},
			}
			sriovNetworkNodeState.Status.Interfaces = sriovnetworkv1.InterfaceExts{
				{
					Name:       "eno1",
					NumVfs:     10,
					PciAddress: "0000:d8:00.0",
					Vendor:     "15b3",
					LinkType:   "ETH",
				},
			}

			needDrain, needReboot, err := m.OnNodeStateChange(sriovNetworkNodeState)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("change required for link type but the policy is externally managed, failing"))
			Expect(needDrain).To(BeFalse())
			Expect(needReboot).To(BeFalse())
		})

		It("should failed if not able to check if the nic is externally managed PF", func() {
			h.EXPECT().IsKernelLockdownMode().Return(false)
			h.EXPECT().LoadPfsStatus("0000:d8:00.0").Return(&sriovnetworkv1.Interface{ExternallyManaged: true}, true, nil).AnyTimes()
			sriovNetworkNodeState.Spec.Interfaces = sriovnetworkv1.Interfaces{}
			sriovNetworkNodeState.Status.Interfaces = sriovnetworkv1.InterfaceExts{
				{
					Name:       "eno1",
					NumVfs:     10,
					PciAddress: "0000:d8:00.0",
					Vendor:     "15b3",
					LinkType:   "ETH",
				},
			}

			needDrain, needReboot, err := m.OnNodeStateChange(sriovNetworkNodeState)
			Expect(err).ToNot(HaveOccurred())
			Expect(needDrain).To(BeFalse())
			Expect(needReboot).To(BeFalse())
			_, exist := attributesToChange["0000:d8:00.0"]
			Expect(exist).To(BeFalse())
		})

		It("should not reset the firmware if the device was not configured by us", func() {
			h.EXPECT().IsKernelLockdownMode().Return(false)
			h.EXPECT().LoadPfsStatus("0000:d8:00.0").Return(nil, false, nil)
			sriovNetworkNodeState.Spec.Interfaces = sriovnetworkv1.Interfaces{}
			sriovNetworkNodeState.Status.Interfaces = sriovnetworkv1.InterfaceExts{
				{
					Name:       "eno1",
					NumVfs:     10,
					PciAddress: "0000:d8:00.0",
					Vendor:     "15b3",
					LinkType:   "ETH",
				},
			}

			needDrain, needReboot, err := m.OnNodeStateChange(sriovNetworkNodeState)
			Expect(err).ToNot(HaveOccurred())
			Expect(needDrain).To(BeFalse())
			Expect(needReboot).To(BeFalse())
			_, exist := attributesToChange["0000:d8:00.0"]
			Expect(exist).To(BeFalse())
		})
	})

	Context("OnNodeStateChange PF_NUM_OF_VF_VALID", func() {
		const (
			port0 = "0000:d8:00.0"
			port1 = "0000:d8:00.1"
		)

		var fwData = func(valid *bool, pfNumOfVf int) *mlx.MlxNic {
			return &mlx.MlxNic{EnableSriov: true, TotalVfs: 5, PfNumOfVfValid: valid, PfNumOfVf: pfNumOfVf}
		}

		var setSpec = func(externallyManaged bool, pciAddresses ...string) {
			sriovNetworkNodeState.Spec.Interfaces = sriovnetworkv1.Interfaces{}
			sriovNetworkNodeState.Status.Interfaces = sriovnetworkv1.InterfaceExts{}
			for _, pciAddress := range pciAddresses {
				sriovNetworkNodeState.Spec.Interfaces = append(sriovNetworkNodeState.Spec.Interfaces, sriovnetworkv1.Interface{
					Name: "ens1f0np0", NumVfs: 5, PciAddress: pciAddress, ExternallyManaged: externallyManaged,
					VfGroups: []sriovnetworkv1.VfGroup{{ResourceName: "test", PolicyName: "test", VfRange: "0-4"}},
				})
				sriovNetworkNodeState.Status.Interfaces = append(sriovNetworkNodeState.Status.Interfaces, sriovnetworkv1.InterfaceExt{
					Name: "ens1f0np0", PciAddress: pciAddress, Vendor: "15b3", DeviceID: "1013", TotalVfs: 0,
				})
			}
		}

		BeforeEach(func() {
			vars.FeatureGate.Init(map[string]bool{consts.MellanoxFirmwareResetFeatureGate: false})
		})

		It("should disable the override and reboot when SRIOV_EN and NUM_OF_VFS already match (reproducer)", func() {
			h.EXPECT().IsKernelLockdownMode().Return(false).Times(2)
			h.EXPECT().GetMlxNicFwData(port0).Return(fwData(ptr.To(true), 0), fwData(ptr.To(true), 0), nil)
			setSpec(false, port0)

			needDrain, needReboot, err := m.OnNodeStateChange(sriovNetworkNodeState)
			Expect(err).ToNot(HaveOccurred())
			Expect(needDrain).To(BeTrue())
			Expect(needReboot).To(BeTrue())
			Expect(pciAddressesToReset).To(Equal([]string{port0}))

			h.EXPECT().MlxConfigFW(map[string]mlx.MlxNic{
				port0: {TotalVfs: -1, PfNumOfVfValid: ptr.To(false)},
			}).Return(nil)
			Expect(m.Apply()).To(Succeed())
		})

		It("should reboot without writing again when the disabling is staged and converge after reboot", func() {
			setSpec(false, port0)

			// after Apply, before reboot: next boot already has the override disabled
			h.EXPECT().IsKernelLockdownMode().Return(false)
			h.EXPECT().GetMlxNicFwData(port0).Return(fwData(ptr.To(true), 0), fwData(ptr.To(false), 0), nil)
			needDrain, needReboot, err := m.OnNodeStateChange(sriovNetworkNodeState)
			Expect(err).ToNot(HaveOccurred())
			Expect(needDrain).To(BeTrue())
			Expect(needReboot).To(BeTrue())
			Expect(attributesToChange).To(Equal(map[string]mlx.MlxNic{port0: {TotalVfs: -1}}))

			// after reboot: second reconciliation is a no-op
			h.EXPECT().IsKernelLockdownMode().Return(false)
			h.EXPECT().GetMlxNicFwData(port0).Return(fwData(ptr.To(false), 0), fwData(ptr.To(false), 0), nil)
			needDrain, needReboot, err = m.OnNodeStateChange(sriovNetworkNodeState)
			Expect(err).ToNot(HaveOccurred())
			Expect(needDrain).To(BeFalse())
			Expect(needReboot).To(BeFalse())
			Expect(attributesToChange).To(BeEmpty())
			Expect(pciAddressesToReset).To(BeEmpty())
		})

		It("should disable the override without reboot when the enabling is only pending for next boot", func() {
			h.EXPECT().IsKernelLockdownMode().Return(false)
			h.EXPECT().GetMlxNicFwData(port0).Return(fwData(ptr.To(false), 0), fwData(ptr.To(true), 0), nil)
			setSpec(false, port0)

			needDrain, needReboot, err := m.OnNodeStateChange(sriovNetworkNodeState)
			Expect(err).ToNot(HaveOccurred())
			Expect(needDrain).To(BeFalse())
			Expect(needReboot).To(BeFalse())
			Expect(attributesToChange).To(Equal(map[string]mlx.MlxNic{port0: {TotalVfs: -1, PfNumOfVfValid: ptr.To(false)}}))
			Expect(pciAddressesToReset).To(BeEmpty())
		})

		It("should not change anything if the override is already disabled", func() {
			h.EXPECT().IsKernelLockdownMode().Return(false)
			h.EXPECT().GetMlxNicFwData(port0).Return(fwData(ptr.To(false), 0), fwData(ptr.To(false), 0), nil)
			setSpec(false, port0)

			needDrain, needReboot, err := m.OnNodeStateChange(sriovNetworkNodeState)
			Expect(err).ToNot(HaveOccurred())
			Expect(needDrain).To(BeFalse())
			Expect(needReboot).To(BeFalse())
			Expect(attributesToChange).To(BeEmpty())
		})

		It("should not change anything if the firmware doesn't expose the override", func() {
			h.EXPECT().IsKernelLockdownMode().Return(false)
			h.EXPECT().GetMlxNicFwData(port0).Return(fwData(nil, 0), fwData(nil, 0), nil)
			setSpec(false, port0)

			needDrain, needReboot, err := m.OnNodeStateChange(sriovNetworkNodeState)
			Expect(err).ToNot(HaveOccurred())
			Expect(needDrain).To(BeFalse())
			Expect(needReboot).To(BeFalse())
			Expect(attributesToChange).To(BeEmpty())
		})

		It("should fail without changing the firmware if externally managed and the override blocks the requested VFs", func() {
			h.EXPECT().IsKernelLockdownMode().Return(false)
			h.EXPECT().GetMlxNicFwData(port0).Return(fwData(ptr.To(true), 0), fwData(ptr.To(true), 0), nil)
			setSpec(true, port0)

			_, _, err := m.OnNodeStateChange(sriovNetworkNodeState)
			Expect(err).To(MatchError("interface 0000:d8:00.0 has PF_NUM_OF_VF_VALID enabled in the firmware with PF_NUM_OF_VF 0 " +
				"lower than the requested 5 VFs but the policy is externally managed failing: " +
				"set PF_NUM_OF_VF_VALID=False or PF_NUM_OF_VF>=5 in the firmware"))
			Expect(attributesToChange).To(BeEmpty())
		})

		It("should not change the firmware if externally managed and the override allows the requested VFs", func() {
			h.EXPECT().IsKernelLockdownMode().Return(false)
			h.EXPECT().GetMlxNicFwData(port0).Return(fwData(ptr.To(true), 8), fwData(ptr.To(true), 8), nil)
			setSpec(true, port0)

			needDrain, needReboot, err := m.OnNodeStateChange(sriovNetworkNodeState)
			Expect(err).ToNot(HaveOccurred())
			Expect(needDrain).To(BeFalse())
			Expect(needReboot).To(BeFalse())
			Expect(attributesToChange).To(BeEmpty())
		})

		It("should disable the override on the dual-port NIC port that has it enabled", func() {
			h.EXPECT().IsKernelLockdownMode().Return(false)
			h.EXPECT().GetMlxNicFwData(port0).Return(fwData(ptr.To(false), 0), fwData(ptr.To(false), 0), nil)
			h.EXPECT().GetMlxNicFwData(port1).Return(fwData(ptr.To(true), 0), fwData(ptr.To(true), 0), nil)
			setSpec(false, port0, port1)

			needDrain, needReboot, err := m.OnNodeStateChange(sriovNetworkNodeState)
			Expect(err).ToNot(HaveOccurred())
			Expect(needDrain).To(BeTrue())
			Expect(needReboot).To(BeTrue())
			Expect(attributesToChange).To(HaveKeyWithValue(port1, mlx.MlxNic{TotalVfs: -1, PfNumOfVfValid: ptr.To(false)}))
			if port0Attrs, ok := attributesToChange[port0]; ok {
				Expect(port0Attrs).To(Equal(mlx.MlxNic{TotalVfs: -1}))
			}
			// a single firmware reset per NIC
			Expect(pciAddressesToReset).To(HaveLen(1))
		})

		It("should disable the override on both ports of a dual-port NIC", func() {
			h.EXPECT().IsKernelLockdownMode().Return(false)
			h.EXPECT().GetMlxNicFwData(port0).Return(fwData(ptr.To(true), 0), fwData(ptr.To(true), 0), nil)
			h.EXPECT().GetMlxNicFwData(port1).Return(fwData(ptr.To(true), 0), fwData(ptr.To(true), 0), nil)
			setSpec(false, port0, port1)

			needDrain, needReboot, err := m.OnNodeStateChange(sriovNetworkNodeState)
			Expect(err).ToNot(HaveOccurred())
			Expect(needDrain).To(BeTrue())
			Expect(needReboot).To(BeTrue())
			Expect(attributesToChange).To(Equal(map[string]mlx.MlxNic{
				port0: {TotalVfs: -1, PfNumOfVfValid: ptr.To(false)},
				port1: {TotalVfs: -1, PfNumOfVfValid: ptr.To(false)},
			}))
			Expect(pciAddressesToReset).To(HaveLen(1))
		})

		It("should keep the reboot requirement of a NIC when another NIC doesn't need it", func() {
			h.EXPECT().IsKernelLockdownMode().Return(false)
			h.EXPECT().GetMlxNicFwData(port0).Return(fwData(ptr.To(true), 0), fwData(ptr.To(true), 0), nil)
			h.EXPECT().GetMlxNicFwData("0000:d9:00.0").Return(fwData(ptr.To(false), 0), fwData(ptr.To(false), 0), nil)
			setSpec(false, port0, "0000:d9:00.0")

			needDrain, needReboot, err := m.OnNodeStateChange(sriovNetworkNodeState)
			Expect(err).ToNot(HaveOccurred())
			Expect(needDrain).To(BeTrue())
			Expect(needReboot).To(BeTrue())
			Expect(pciAddressesToReset).To(Equal([]string{port0}))
		})
	})

	Context("Apply", func() {
		It("should not call fw configuration for mlx devices in secure boot active environment", func() {
			h.EXPECT().IsKernelLockdownMode().Return(true)
			err := m.Apply()
			Expect(err).ToNot(HaveOccurred())
		})
		It("should return eror if call mlx config fw failed", func() {
			h.EXPECT().IsKernelLockdownMode().Return(false)
			h.EXPECT().MlxConfigFW(gomock.Any()).Return(fmt.Errorf("failed to configure fw"))
			err := m.Apply()
			Expect(err).To(HaveOccurred())
		})

		It("should call mlx config fw without fwreset if feature flag is disabled", func() {
			vars.FeatureGate.Init(map[string]bool{consts.MellanoxFirmwareResetFeatureGate: false})
			h.EXPECT().IsKernelLockdownMode().Return(false)
			h.EXPECT().MlxConfigFW(gomock.Any()).Return(nil)
			err := m.Apply()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should call mlx config fw with fwreset if feature flag is enabled", func() {
			vars.FeatureGate.Init(map[string]bool{consts.MellanoxFirmwareResetFeatureGate: true})
			h.EXPECT().IsKernelLockdownMode().Return(false)
			h.EXPECT().MlxConfigFW(gomock.Any()).Return(nil)
			h.EXPECT().MlxResetFW(gomock.Any(), gomock.Any()).Return(nil)
			err := m.Apply()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should reset VFs on both ports for dual-port NIC when firmware reset is required", func() {
			vars.FeatureGate.Init(map[string]bool{consts.MellanoxFirmwareResetFeatureGate: true})

			// Setup dual-port NIC status
			mellanoxNicsStatus = map[string]map[string]sriovnetworkv1.InterfaceExt{
				"0000:d8:00.": {
					"0000:d8:00.0": {PciAddress: "0000:d8:00.0", Vendor: "15b3"},
					"0000:d8:00.1": {PciAddress: "0000:d8:00.1", Vendor: "15b3"},
				},
			}

			// Setup dual-port NIC spec
			mellanoxNicsSpec = map[string]sriovnetworkv1.Interface{
				"0000:d8:00.0": {PciAddress: "0000:d8:00.0", NumVfs: 10},
				"0000:d8:00.1": {PciAddress: "0000:d8:00.1", NumVfs: 8},
			}

			// Only one port needs firmware reset
			pciAddressesToReset = []string{"0000:d8:00.0"}

			h.EXPECT().IsKernelLockdownMode().Return(false)
			h.EXPECT().MlxConfigFW(gomock.Any()).Return(nil)

			h.EXPECT().MlxResetFW([]string{"0000:d8:00.0"}, gomock.Any()).Return(nil)

			err := m.Apply()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should reset VFs on single port NIC when firmware reset is required", func() {
			vars.FeatureGate.Init(map[string]bool{consts.MellanoxFirmwareResetFeatureGate: true})

			// Setup single-port NIC status
			mellanoxNicsStatus = map[string]map[string]sriovnetworkv1.InterfaceExt{
				"0000:d8:00.": {
					"0000:d8:00.0": {PciAddress: "0000:d8:00.0", Vendor: "15b3"},
				},
			}

			// Setup single-port NIC spec
			mellanoxNicsSpec = map[string]sriovnetworkv1.Interface{
				"0000:d8:00.0": {PciAddress: "0000:d8:00.0", NumVfs: 10},
			}

			pciAddressesToReset = []string{"0000:d8:00.0"}

			h.EXPECT().IsKernelLockdownMode().Return(false)
			h.EXPECT().MlxConfigFW(gomock.Any()).Return(nil)

			h.EXPECT().MlxResetFW([]string{"0000:d8:00.0"}, gomock.Any()).Return(nil)

			err := m.Apply()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should call MlxResetFW with correct parameters for multiple NICs", func() {
			vars.FeatureGate.Init(map[string]bool{consts.MellanoxFirmwareResetFeatureGate: true})

			pciAddressesToReset = []string{"0000:d8:00.0", "0000:d9:00.0"}

			h.EXPECT().IsKernelLockdownMode().Return(false)
			h.EXPECT().MlxConfigFW(gomock.Any()).Return(nil)

			h.EXPECT().MlxResetFW([]string{"0000:d8:00.0", "0000:d9:00.0"}, gomock.Any()).Return(nil)

			err := m.Apply()
			Expect(err).ToNot(HaveOccurred())
		})

		BeforeEach(func() {
			// Reset global state before each test
			pciAddressesToReset = []string{}
			mellanoxNicsStatus = map[string]map[string]sriovnetworkv1.InterfaceExt{}
			mellanoxNicsSpec = map[string]sriovnetworkv1.Interface{}
		})
	})
})
