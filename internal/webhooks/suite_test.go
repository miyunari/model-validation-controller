package webhooks

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/miyunari/model-validation-controller/api/v1alpha1"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/test"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	cfg       *rest.Config
	k8sClient client.Client // You'll be using this client in your tests.
	testEnv   *envtest.Environment
	ctx       context.Context
	cancel    context.CancelFunc
)

func TestAPIs(t *testing.T) {
	fs := test.InitKlog(t)
	_ = fs.Set("v", "5")
	klog.SetOutput(GinkgoWriter)
	ctrl.SetLogger(klog.NewKlogr())

	RegisterFailHandler(Fail)
	RunSpecs(t, "Webhook Suite")
}

var _ = BeforeSuite(func() {
	ctx, cancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,

		// The BinaryAssetsDirectory is only required if you want to run the tests directly
		// without call the makefile target test. If not informed it will look for the
		// default path defined in controller-runtime which is /usr/local/kubebuilder/.
		// Note that you must have the required binaries setup under the bin directory to perform
		// the tests directly. When we run make test it will be setup and used automatically.
		BinaryAssetsDirectory: filepath.Join("..", "..", "bin", "k8s",
			fmt.Sprintf("1.31.0-%s-%s", runtime.GOOS, runtime.GOARCH)),
		WebhookInstallOptions: envtest.WebhookInstallOptions{Paths: []string{filepath.Join("..", "..", "config", "webhook")}},
	}

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = v1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	webhookServer := webhook.NewServer(webhook.Options{
		Host:    testEnv.WebhookInstallOptions.LocalServingHost,
		Port:    testEnv.WebhookInstallOptions.LocalServingPort,
		CertDir: testEnv.WebhookInstallOptions.LocalServingCertDir,
	})

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:        scheme.Scheme,
		WebhookServer: webhookServer,
	})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	// Create a decoder for your webhook
	decoder := admission.NewDecoder(scheme.Scheme)
	podWebhookHandler := NewPodInterceptor(mgr.GetClient(), decoder)
	mgr.GetWebhookServer().Register("/mutate-v1-pod", &admission.Webhook{
		Handler: podWebhookHandler,
	})

	go func() {
		defer GinkgoRecover()
		err = mgr.Start(ctx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()
})

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
