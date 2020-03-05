package timechaos_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	. "github.com/pingcap/chaos-mesh/controllers/timechaos"
	chaosdaemon "github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/pingcap/chaos-mesh/pkg/mock"
)

func TestTimechaos(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"TimeChaos Suite",
		[]Reporter{envtest.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))
	close(done)
}, 60)

var _ = AfterSuite(func() {
})

// Assert *MockChaosDaemonClient implements chaosdaemon.ChaosDaemonClientInterface.
var _ ChaosDaemonClientInterface = (*MockChaosDaemonClient)(nil)

// todo: move this to somewhere else
type MockChaosDaemonClient struct{}

func (c *MockChaosDaemonClient) SetNetem(ctx context.Context, in *chaosdaemon.NetemRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	panic("implement me")
}

func (c *MockChaosDaemonClient) DeleteNetem(ctx context.Context, in *chaosdaemon.NetemRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	panic("implement me")
}

func (c *MockChaosDaemonClient) FlushIpSet(ctx context.Context, in *chaosdaemon.IpSetRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	panic("implement me")
}

func (c *MockChaosDaemonClient) FlushIptables(ctx context.Context, in *chaosdaemon.IpTablesRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	panic("implement me")
}

func (c *MockChaosDaemonClient) SetTimeOffset(ctx context.Context, in *chaosdaemon.TimeRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	panic("implement me")
}

func (c *MockChaosDaemonClient) RecoverTimeOffset(ctx context.Context, in *chaosdaemon.TimeRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	panic("implement me")
}

func (c *MockChaosDaemonClient) ContainerKill(ctx context.Context, in *chaosdaemon.ContainerRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	panic("implement me")
}

func (c *MockChaosDaemonClient) Close() error {
	if err := mock.On("CloseChaosDaemonClient"); err != nil {
		return err.(error)
	}
	return nil
}

var _ = Describe("TimeChaos", func() {
	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	// Add Tests for OpenAPI validation (or additional CRD features) specified in
	// your API definition.
	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("TimeChaos", func() {
		It("TimeChaos Apply", func() {
			mock.With("MockSelectAndGeneratePods", nil)

			duration := "invalid_duration"

			timechaos := v1alpha1.TimeChaos{
				TypeMeta: metav1.TypeMeta{
					Kind:       "TimeChaos",
					APIVersion: "v1",
				},
				Spec: v1alpha1.TimeChaosSpec{
					Mode:  "FixedPodMode",
					Value: "0",
					Selector: v1alpha1.SelectorSpec{
						Namespaces: []string{"namespace"},
					},
					TimeOffset: v1alpha1.TimeOffset{},
					Duration:   &duration,
					Scheduler:  nil,
				},
			}

			r := Reconciler{
				Client:        fake.NewFakeClient(),
				EventRecorder: &record.FakeRecorder{},
				Log:           ctrl.Log.WithName("controllers").WithName("TimeChaos"),
			}

			err := r.Apply(context.TODO(), ctrl.Request{}, &timechaos)

			Expect(err).To(HaveOccurred())
			// fixme: incorrect Pods mock
			Expect(err.Error()).To(ContainSubstring("no pod is selected"))
		})
	})
})
