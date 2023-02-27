package function

import (
	"crypto/tls"
	"fmt"
	"net"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/ondat/operator-toolkit/singleton"
	tdv1alpha1 "github.com/ondat/operator-toolkit/testdata/api/v1alpha1"
	tkadmission "github.com/ondat/operator-toolkit/webhook/admission"
	"github.com/ondat/operator-toolkit/webhook/builder"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	cfg       *rest.Config
	k8sClient client.Client
	testEnv   *envtest.Environment
)

func TestWebhookFunctions(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Webhook Function Suite")
}

var _ = BeforeSuite(func(ctx SpecContext) {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "example", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: false,
		WebhookInstallOptions: envtest.WebhookInstallOptions{
			Paths: []string{filepath.Join("..", "..", "example", "config", "webhook")},
		},
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	scheme := runtime.NewScheme()
	err = tdv1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	// start webhook server using Manager
	// TODO: Create and run webhook server and manager per test scenarios.
	webhookInstallOptions := &testEnv.WebhookInstallOptions
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme,
		Host:               webhookInstallOptions.LocalServingHost,
		Port:               webhookInstallOptions.LocalServingPort,
		CertDir:            webhookInstallOptions.LocalServingCertDir,
		LeaderElection:     false,
		MetricsBindAddress: "0",
	})
	Expect(err).NotTo(HaveOccurred())

	stn, err := singleton.GetInstance(&tdv1alpha1.GameList{}, scheme)
	Expect(err).NotTo(HaveOccurred())

	gc := GameController{
		CtrlName: "test-game-controller",
		ValidateCreateFuncs: []tkadmission.ValidateCreateFunc{
			ValidateSingletonCreate(stn, k8sClient),
		},
	}
	Expect(gc.SetupWithManager(mgr)).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:webhook

	go func() {
		err = mgr.Start(ctx)
		if err != nil {
			Expect(err).NotTo(HaveOccurred())
		}
	}()

	// wait for the webhook server to get ready
	dialer := &net.Dialer{Timeout: time.Second}
	addrPort := fmt.Sprintf("%s:%d", webhookInstallOptions.LocalServingHost, webhookInstallOptions.LocalServingPort)
	Eventually(func() error {
		conn, err := tls.DialWithDialer(dialer, "tcp", addrPort, &tls.Config{InsecureSkipVerify: true})
		if err != nil {
			return err
		}
		conn.Close()
		return nil
	}).Should(Succeed())
}, NodeTimeout(60*time.Second))

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

// Implement an admission controller.

type GameController struct {
	CtrlName string

	DefaultFuncs        []tkadmission.DefaultFunc
	ValidateCreateFuncs []tkadmission.ValidateCreateFunc
	ValidateUpdateFuncs []tkadmission.ValidateUpdateFunc
	ValidateDeleteFuncs []tkadmission.ValidateDeleteFunc
}

var _ tkadmission.Controller = &GameController{}

func (gc *GameController) Name() string {
	return gc.CtrlName
}

func (gc *GameController) GetNewObject() client.Object {
	return &tdv1alpha1.Game{}
}

func (gc *GameController) RequireDefaulting(obj client.Object) bool {
	return true
}

func (gc *GameController) RequireValidating(obj client.Object) bool {
	return true
}

func (gc *GameController) Default() []tkadmission.DefaultFunc {
	return gc.DefaultFuncs
}

func (gc *GameController) ValidateCreate() []tkadmission.ValidateCreateFunc {
	return gc.ValidateCreateFuncs
}

func (gc *GameController) ValidateUpdate() []tkadmission.ValidateUpdateFunc {
	return gc.ValidateUpdateFuncs
}

func (gc *GameController) ValidateDelete() []tkadmission.ValidateDeleteFunc {
	return gc.ValidateDeleteFuncs
}

// NOTE: The endpoints are based on the webhook configuration registered from
// the manifest files in the example project in the envtest
// Environment.WebhookInstallOptions.
func (gc *GameController) SetupWithManager(mgr manager.Manager) error {
	return builder.WebhookManagedBy(mgr).
		MutatePath("/mutate-app-example-com-v1alpha1-game").
		ValidatePath("/validate-app-example-com-v1alpha1-game").
		Complete(gc)
}
