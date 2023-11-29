package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	mcfgv1 "github.com/openshift/machine-config-operator/pkg/apis/machineconfiguration.openshift.io/v1"
	daemonconsts "github.com/openshift/machine-config-operator/pkg/daemon/constants"

	"sigs.k8s.io/controller-runtime/pkg/client"

	sriovnetworkv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	constants "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/consts"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/utils"
)

var _ = Describe("Drain Controller", func() {
	BeforeEach(func() {
		Expect(k8sClient.DeleteAllOf(context.Background(), &corev1.Node{})).ToNot(HaveOccurred())
		Expect(k8sClient.DeleteAllOf(context.Background(), &sriovnetworkv1.SriovNetworkNodeState{}, client.InNamespace(namespace))).ToNot(HaveOccurred())
	})

	Context("when there is only one node", func() {

		It("should drain", func(ctx context.Context) {
			node, nodeState := createIdleNode("node1")

			simulateDaemonSetAnnotation(node, constants.DrainRequired)

			expectNodeStateAnnotation(nodeState, constants.DrainComplete)
			expectNodeIsNotSchedulable(node)

			simulateDaemonSetAnnotation(node, constants.DrainIdle)

			expectNodeStateAnnotation(nodeState, constants.DrainIdle)
			expectNodeIsSchedulable(node)
		})
	})

	Context("when there are multiple nodes", func() {

		It("should drain nodes serially", func(ctx context.Context) {

			node1, nodeState1 := createIdleNode("node1")
			node2, nodeState2 := createIdleNode("node2")
			node3, nodeState3 := createIdleNode("node3")

			// Two nodes require to drain at the same time
			simulateDaemonSetAnnotation(node1, constants.DrainRequired)
			simulateDaemonSetAnnotation(node2, constants.DrainRequired)

			// Only the first node drains
			expectNodeStateAnnotation(nodeState1, constants.DrainComplete)
			expectNodeStateAnnotation(nodeState2, constants.DrainIdle)
			expectNodeStateAnnotation(nodeState3, constants.DrainIdle)
			expectNodeIsNotSchedulable(node1)
			expectNodeIsSchedulable(node2)
			expectNodeIsSchedulable(node3)

			simulateDaemonSetAnnotation(node1, constants.DrainIdle)

			expectNodeStateAnnotation(nodeState1, constants.DrainIdle)
			expectNodeIsSchedulable(node1)

			// Second node starts draining
			expectNodeStateAnnotation(nodeState1, constants.DrainIdle)
			expectNodeStateAnnotation(nodeState2, constants.DrainComplete)
			expectNodeStateAnnotation(nodeState3, constants.DrainIdle)
			expectNodeIsSchedulable(node1)
			expectNodeIsNotSchedulable(node2)
			expectNodeIsSchedulable(node3)

			simulateDaemonSetAnnotation(node2, constants.DrainIdle)

			expectNodeStateAnnotation(nodeState1, constants.DrainIdle)
			expectNodeStateAnnotation(nodeState2, constants.DrainIdle)
			expectNodeStateAnnotation(nodeState3, constants.DrainIdle)
			expectNodeIsSchedulable(node1)
			expectNodeIsSchedulable(node2)
			expectNodeIsSchedulable(node3)
		})
	})

	Context("on OpenShift", func() {

		BeforeEach(func() {
			testOpenshiftContext.IsOpenShiftCluster = true
			DeferCleanup(func() {
				testOpenshiftContext.IsOpenShiftCluster = false
			})
		})

		It("should pause MCP when draining", func(ctx context.Context) {
			node, nodeState := createOpenshiftIdleWorkerNode("node1")
			mcp := createTestMachineConfigPool()

			simulateDaemonSetAnnotation(node, constants.DrainRequired)

			expectNodeStateAnnotation(nodeState, constants.DrainComplete)
			expectNodeIsNotSchedulable(node)
			expectMCPIsPaused(mcp)

			simulateDaemonSetAnnotation(node, constants.DrainIdle)

			expectNodeStateAnnotation(nodeState, constants.DrainIdle)
			expectNodeIsSchedulable(node)
			expectMCPIsNotPaused(mcp)
		})
	})
})

func expectNodeStateAnnotation(nodeState *sriovnetworkv1.SriovNetworkNodeState, expectedAnnotationValue string) {
	EventuallyWithOffset(1, func(g Gomega) {
		g.Expect(k8sClient.Get(context.Background(), types.NamespacedName{Namespace: nodeState.Namespace, Name: nodeState.Name}, nodeState)).
			ToNot(HaveOccurred())

		g.Expect(utils.ObjectHasAnnotation(nodeState, constants.NodeStateDrainAnnotationCurrent, expectedAnnotationValue)).
			To(BeTrue(),
				"Node[%s] annotation[%s] == '%s'. Expected '%s'", nodeState.Name, constants.NodeDrainAnnotation, nodeState.GetAnnotations()[constants.NodeStateDrainAnnotationCurrent], expectedAnnotationValue)

	}, "10s", "1s").Should(Succeed())
}

func expectNodeIsNotSchedulable(node *corev1.Node) {
	EventuallyWithOffset(1, func(g Gomega) {
		g.Expect(k8sClient.Get(context.Background(), types.NamespacedName{Name: node.Name}, node)).
			ToNot(HaveOccurred())

		g.Expect(node.Spec.Unschedulable).To(BeTrue())
	}, "10s", "1s").Should(Succeed())
}

func expectNodeIsSchedulable(node *corev1.Node) {
	EventuallyWithOffset(1, func(g Gomega) {
		g.Expect(k8sClient.Get(context.Background(), types.NamespacedName{Name: node.Name}, node)).
			ToNot(HaveOccurred())

		g.Expect(node.Spec.Unschedulable).To(BeFalse())
	}, "10s", "1s").Should(Succeed())
}

func simulateDaemonSetAnnotation(node *corev1.Node, drainAnnotationValue string) {
	ExpectWithOffset(1,
		utils.AnnotateObject(node, constants.NodeDrainAnnotation, drainAnnotationValue, k8sClient)).
		ToNot(HaveOccurred())
}

func createIdleNode(nodeName string) (*corev1.Node, *sriovnetworkv1.SriovNetworkNodeState) {
	node := corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeName,
			Annotations: map[string]string{
				constants.NodeDrainAnnotation: constants.DrainIdle,
			},
		},
	}

	nodeState := sriovnetworkv1.SriovNetworkNodeState{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nodeName,
			Namespace: namespace,
			Annotations: map[string]string{
				constants.NodeStateDrainAnnotationCurrent: constants.DrainIdle,
			},
		},
	}

	Expect(k8sClient.Create(ctx, &node)).ToNot(HaveOccurred())
	Expect(k8sClient.Create(ctx, &nodeState)).ToNot(HaveOccurred())

	return &node, &nodeState
}

func createOpenshiftIdleWorkerNode(nodeName string) (*corev1.Node, *sriovnetworkv1.SriovNetworkNodeState) {
	node := corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeName,
			Annotations: map[string]string{
				constants.NodeDrainAnnotation:                  constants.DrainIdle,
				daemonconsts.DesiredMachineConfigAnnotationKey: "00-worker",
			},
			Labels: map[string]string{
				"node-role.kubernetes.io/worker": "",
			},
		},
	}

	nodeState := sriovnetworkv1.SriovNetworkNodeState{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nodeName,
			Namespace: namespace,
			Annotations: map[string]string{
				constants.NodeStateDrainAnnotationCurrent: constants.DrainIdle,
			},
		},
	}

	Expect(k8sClient.Create(ctx, &node)).ToNot(HaveOccurred())
	Expect(k8sClient.Create(ctx, &nodeState)).ToNot(HaveOccurred())

	return &node, &nodeState
}

func createTestMachineConfigPool() *mcfgv1.MachineConfigPool {
	mcp := &mcfgv1.MachineConfigPool{
		ObjectMeta: metav1.ObjectMeta{
			Name: "worker",
		},
		Spec: mcfgv1.MachineConfigPoolSpec{
			NodeSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"node-role.kubernetes.io/worker": "",
				},
			},
		},
	}

	err := k8sClient.Create(context.Background(), mcp)
	Expect(err).NotTo(HaveOccurred())

	mcfgv1.SetMachineConfigPoolCondition(&mcp.Status, *mcfgv1.NewMachineConfigPoolCondition(mcfgv1.MachineConfigPoolUpdating, corev1.ConditionFalse, "", ""))
	mcfgv1.SetMachineConfigPoolCondition(&mcp.Status, *mcfgv1.NewMachineConfigPoolCondition(mcfgv1.MachineConfigPoolDegraded, corev1.ConditionFalse, "", ""))
	mcfgv1.SetMachineConfigPoolCondition(&mcp.Status, *mcfgv1.NewMachineConfigPoolCondition(mcfgv1.MachineConfigPoolUpdated, corev1.ConditionTrue, "", ""))

	_, err = testOpenshiftContext.McClient.MachineconfigurationV1().MachineConfigPools().UpdateStatus(context.Background(), mcp, metav1.UpdateOptions{})
	Expect(err).NotTo(HaveOccurred())

	mc := &mcfgv1.MachineConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "00-worker",
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: mcfgv1.GroupVersion.String(),
				Kind:       "MachineConfigPool",
				Name:       mcp.Name,
				UID:        mcp.UID,
			}},
		},
	}

	err = k8sClient.Create(context.Background(), mc)
	Expect(err).NotTo(HaveOccurred())

	return mcp
}

func expectMCPIsPaused(mcp *mcfgv1.MachineConfigPool) {
	EventuallyWithOffset(1, func(g Gomega) {
		g.Expect(k8sClient.Get(context.Background(), types.NamespacedName{Namespace: mcp.Namespace, Name: mcp.Name}, mcp)).
			ToNot(HaveOccurred())

		g.Expect(mcp.Spec.Paused).To(BeTrue())
	}, "10s", "1s").Should(Succeed())
}

func expectMCPIsNotPaused(mcp *mcfgv1.MachineConfigPool) {
	EventuallyWithOffset(1, func(g Gomega) {
		g.Expect(k8sClient.Get(context.Background(), types.NamespacedName{Namespace: mcp.Namespace, Name: mcp.Name}, mcp)).
			ToNot(HaveOccurred())

		g.Expect(mcp.Spec.Paused).To(BeFalse())
	}, "10s", "1s").Should(Succeed())
}
