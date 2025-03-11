package tests

import (
	"fmt"
	"os"
	"path"
	"time"

	sriovv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/test/util/cluster"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/test/util/discovery"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/test/util/namespaces"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/test/util/network"
	"github.com/openshift-kni/k8sreporter"
	"k8s.io/apimachinery/pkg/runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("[sriov] Operator Logs", Ordered, ContinueOnFailure, func() {

	BeforeAll(func() {
		err := namespaces.Create(namespaces.Test, clients)
		Expect(err).ToNot(HaveOccurred())

		err = namespaces.Clean(operatorNamespace, namespaces.Test, clients, discovery.Enabled())
		Expect(err).ToNot(HaveOccurred())

		err = namespaces.CleanPods(namespaces.Test, clients)
		Expect(err).ToNot(HaveOccurred())

		WaitForSRIOVStable()

		sriovInfos, err = cluster.DiscoverSriov(clients, operatorNamespace)
		Expect(err).ToNot(HaveOccurred())

		//DeferCleanup(namespaces.Clean, operatorNamespace, namespaces.Test, clients, discovery.Enabled())
	})

	FIt("setup reference scenario", func() {
		var node string = sriovInfos.Nodes[0]

		startTime := time.Now()

		reporter, err := k8sreporter.New(
			"",
			func(s *runtime.Scheme) error { return nil },
			func(ns string) bool { return ns == operatorNamespace },
			"./logs-report",
		)
		Expect(err).ToNot(HaveOccurred())
		err = os.MkdirAll(path.Join("./logs-report"), 0755)
		Expect(err).ToNot(HaveOccurred())

		//DeferCleanup(namespaces.CleanPods, namespaces.Test, clients)
		spawnPods := []func(){}

		nic, _ := sriovInfos.FindOneMellanoxSriovDevice(node)
		if nic != nil {
			By(fmt.Sprintf("configuring mellanox nic %s", nic.Name))
			_, err := network.CreateSriovPolicy(clients, "test-mlx-netdevice-", operatorNamespace, nic.Name+"#10-20", node, 32, "mlxnetdevice", "netdevice")
			Expect(err).ToNot(HaveOccurred())

			err = network.CreateSriovNetwork(clients, nic, "test-mlx-netdevice", namespaces.Test, operatorNamespace, "mlxnetdevice", ipamIpv4)
			Expect(err).ToNot(HaveOccurred())

			_, err = network.CreateSriovPolicy(clients, "test-mlx-rdma-", operatorNamespace, nic.Name+"#21-31", node, 32, "mlxnetdevicerdma", "netdevice", WithRdma())
			Expect(err).ToNot(HaveOccurred())

			err = network.CreateSriovNetwork(clients, nic, "test-mlx-rdma", namespaces.Test, operatorNamespace, "mlxnetdevicerdma", ipamIpv4)
			Expect(err).ToNot(HaveOccurred())

			waitForNetAttachDef("test-mlx-netdevice", namespaces.Test)
			waitForNetAttachDef("test-mlx-rdma", namespaces.Test)

			spawnPods = append(spawnPods, func() {
				createTestPod(node, []string{"test-mlx-netdevice"})
				createTestPod(node, []string{"test-mlx-netdevice"})
				createTestPod(node, []string{"test-mlx-netdevice"})
				createTestPod(node, []string{"test-mlx-netdevice"})
				createTestPod(node, []string{"test-mlx-rdma"})
				createTestPod(node, []string{"test-mlx-rdma"})
				createTestPod(node, []string{"test-mlx-rdma"})
				createTestPod(node, []string{"test-mlx-rdma"})
			})
		}

		nic, _ = sriovInfos.FindOneIntelSriovDevice(node)
		if nic != nil {
			By(fmt.Sprintf("configuring intel nic %s", nic.Name))

			_, err := network.CreateSriovPolicy(clients, "test-intel-netdevice-", operatorNamespace, nic.Name+"#10-20", node, 32, "intelnetdevice", "netdevice")
			Expect(err).ToNot(HaveOccurred())

			err = network.CreateSriovNetwork(clients, nic, "test-intel-netdevice", namespaces.Test, operatorNamespace, "intelnetdevice", ipamIpv4)
			Expect(err).ToNot(HaveOccurred())

			_, err = network.CreateSriovPolicy(clients, "test-intel-vfio-", operatorNamespace, nic.Name+"#21-31", node, 32, "intelvfio", "vfio-pci")
			Expect(err).ToNot(HaveOccurred())

			err = network.CreateSriovNetwork(clients, nic, "test-intel-vfio", namespaces.Test, operatorNamespace, "intelvfio", ipamIpv4)
			Expect(err).ToNot(HaveOccurred())

			waitForNetAttachDef("test-intel-netdevice", namespaces.Test)
			waitForNetAttachDef("test-intel-vfio", namespaces.Test)

			spawnPods = append(spawnPods, func() {
				createTestPod(node, []string{"test-intel-netdevice"})
				createTestPod(node, []string{"test-intel-netdevice"})
				createTestPod(node, []string{"test-intel-netdevice"})
				createTestPod(node, []string{"test-intel-netdevice"})
				createTestPod(node, []string{"test-intel-vfio"})
				createTestPod(node, []string{"test-intel-vfio"})
				createTestPod(node, []string{"test-intel-vfio"})
				createTestPod(node, []string{"test-intel-vfio"})
			})
		}

		By("wait for configuration to be stable")
		WaitForSRIOVStable()
		time.Sleep(time.Minute)

		By("create sriov pods")
		for _, f := range spawnPods {
			f()
		}

		By("wait")
		waitTime := 60 * time.Minute
		time.Sleep(waitTime)

		By("report")

		reportDir := "logs"
		//err = os.MkdirAll(path.Join("./logs-report", reportDir), 0755)
		//Expect(err).ToNot(HaveOccurred())

		reporter.Dump(time.Since(startTime), reportDir)
	})

})

func WithRdma() func(*sriovv1.SriovNetworkNodePolicy) {
	return func(p *sriovv1.SriovNetworkNodePolicy) {
		p.Spec.IsRdma = true
	}
}
