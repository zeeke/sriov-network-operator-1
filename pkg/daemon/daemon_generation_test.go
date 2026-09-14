package daemon

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	sriovnetworkv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/consts"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/featuregate"
	mockhelper "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/helper/mock"
	mockplatform "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/platform/mock"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/status"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/vars"
)

var errStaleSpecReachedPlugin = errors.New("stale spec reached plugin")

// TestReconcileSkipsStaleSpecAfterStatusUpdate reproduces a spec update that
// lands between Reconcile's initial Get and its InProgress status update. The
// daemon must requeue before calling a plugin with the now-stale spec.
func TestReconcileSkipsStaleSpecAfterStatusUpdate(t *testing.T) {
	ctx := context.Background()

	previousSystemdMode := vars.UsingSystemdMode
	previousManageBridges := vars.ManageSoftwareBridges
	previousExternalDrainer := vars.UseExternalDrainer
	t.Cleanup(func() {
		vars.UsingSystemdMode = previousSystemdMode
		vars.ManageSoftwareBridges = previousManageBridges
		vars.UseExternalDrainer = previousExternalDrainer
	})
	vars.UsingSystemdMode = false
	vars.ManageSoftwareBridges = false
	vars.UseExternalDrainer = false

	scheme := runtime.NewScheme()
	if err := sriovnetworkv1.AddToScheme(scheme); err != nil {
		t.Fatalf("add SriovNetworkNodeState scheme: %v", err)
	}

	nodeState := &sriovnetworkv1.SriovNetworkNodeState{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "node-1",
			Namespace:  "test",
			Generation: 1,
			Annotations: map[string]string{
				consts.NodeStateDrainAnnotation:        consts.DrainIdle,
				consts.NodeStateDrainAnnotationCurrent: consts.DrainIdle,
			},
		},
		// Keeping the same status avoids needing an EventRecorder: this test is
		// interested only in the generation hand-off before plugin invocation.
		Status: sriovnetworkv1.SriovNetworkNodeStateStatus{
			SyncStatus: consts.SyncStatusInProgress,
		},
	}
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(nodeState).Build()

	mockCtrl := gomock.NewController(t)
	hostHelper := mockhelper.NewMockHostHelpersInterface(mockCtrl)
	hostHelper.EXPECT().DiscoverRDMASubsystem().Return("shared", nil)
	platform := mockplatform.NewMockInterface(mockCtrl)
	platform.EXPECT().DiscoverSriovDevices().Return([]sriovnetworkv1.InterfaceExt{}, nil)

	plugin := &recordingPlugin{}
	statusPatcher := &specUpdatingStatusPatcher{client: k8sClient}
	reconciler := &NodeReconciler{
		client:            k8sClient,
		hostHelpers:       hostHelper,
		platformInterface: platform,
		featureGate:       featuregate.New(),
		mainPlugin:        plugin,
		statusPatcher:     statusPatcher,
	}

	result, err := reconciler.Reconcile(ctx, ctrl.Request{
		NamespacedName: client.ObjectKeyFromObject(nodeState),
	})
	if err != nil {
		t.Fatalf("reconcile called a plugin for a stale spec instead of requeueing: %v", err)
	}
	if result.RequeueAfter != 5*time.Second {
		t.Fatalf("expected a 5s requeue after the spec update, got %#v", result)
	}
	if plugin.onNodeStateChangeCalls != 0 {
		t.Fatalf("plugin received stale generation %d after generation %d was stored", plugin.generation, statusPatcher.updatedGeneration)
	}
}

// specUpdatingStatusPatcher models another writer updating the desired spec
// while the daemon performs its SSA status write.
type specUpdatingStatusPatcher struct {
	client            client.Client
	updatedGeneration int64
}

var _ status.Interface = (*specUpdatingStatusPatcher)(nil)

func (p *specUpdatingStatusPatcher) ApplyCondition(context.Context, client.Object, ...metav1.Condition) error {
	return nil
}

func (p *specUpdatingStatusPatcher) ApplyStatus(ctx context.Context, obj client.Object, _ map[string]interface{}, _ []metav1.Condition) error {
	latest := &sriovnetworkv1.SriovNetworkNodeState{}
	if err := p.client.Get(ctx, client.ObjectKeyFromObject(obj), latest); err != nil {
		return err
	}

	latest.Spec.Interfaces = []sriovnetworkv1.Interface{{Name: "new-spec"}}
	// The fake client does not emulate apiserver generation changes, so model
	// the generation increment that a real spec update produces.
	latest.Generation++
	p.updatedGeneration = latest.Generation
	return p.client.Update(ctx, latest)
}

type recordingPlugin struct {
	onNodeStateChangeCalls int
	generation             int64
}

func (p *recordingPlugin) Name() string { return "recording" }

func (p *recordingPlugin) OnNodeStateChange(nodeState *sriovnetworkv1.SriovNetworkNodeState) (bool, bool, error) {
	p.onNodeStateChangeCalls++
	p.generation = nodeState.Generation
	return false, false, errStaleSpecReachedPlugin
}

func (p *recordingPlugin) Apply() error { return nil }

func (p *recordingPlugin) CheckStatusChanges(*sriovnetworkv1.SriovNetworkNodeState) (bool, error) {
	return false, nil
}
