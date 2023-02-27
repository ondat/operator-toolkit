package admission

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestAdmissionWebhook(t *testing.T) {
	RegisterFailHandler(Fail)
	suiteName := "Admission Webhook Suite"
	RunSpecs(t, suiteName)
}

var _ = BeforeSuite(func(ctx SpecContext) {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
}, NodeTimeout(60*time.Second))
