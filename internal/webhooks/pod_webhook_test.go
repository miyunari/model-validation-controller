package webhooks

import (
	"context"

	"github.com/miyunari/model-validation-controller/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Pod webhook", func() {
	Context("Pod webhook test", func() {

		const (
			Name      = "test"
			Namespace = "default"
		)

		ctx := context.Background()

		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      Name,
				Namespace: Namespace,
			},
		}

		typeNamespaceName := types.NamespacedName{Name: Name, Namespace: Namespace}

		BeforeEach(func() {
			By("Creating the Namespace to perform the tests")
			err := k8sClient.Create(ctx, namespace)
			Expect(err).To(Not(HaveOccurred()))

			By("Create ModelValidation resource")
			err = k8sClient.Create(ctx, &v1alpha1.ModelValidation{
				ObjectMeta: metav1.ObjectMeta{
					Name:      Name,
					Namespace: Namespace,
				},
				Spec: v1alpha1.ModelValidationSpec{
					Model: v1alpha1.Model{
						Path:          "test",
						SignaturePath: "test",
					},
					Config: v1alpha1.ValidationConfig{
						SigstoreConfig:   nil,
						PkiConfig:        nil,
						PrivateKeyConfig: nil,
					},
				},
			})
			Expect(err).To(Not(HaveOccurred()))
		})

		AfterEach(func() {
			// TODO(user): Attention if you improve this code by adding other context test you MUST
			// be aware of the current delete namespace limitations.
			// More info: https://book.kubebuilder.io/reference/envtest.html#testing-considerations
			By("Deleting the Namespace to perform the tests")
			_ = k8sClient.Delete(ctx, namespace)
		})

		It("Should create sidecar container", func() {
			By("create labeled pod")
			instance := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      Name,
					Namespace: Namespace,
					Labels:    map[string]string{"validation.rhtas.redhat.com/ml": "true"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test",
							Image: "test",
						},
					},
				},
			}
			err := k8sClient.Create(ctx, instance)
			Expect(err).To(Not(HaveOccurred()))

			By("Checking that validation sidecar was created")
			found := &corev1.Pod{}
			Eventually(func() error {
				return k8sClient.Get(ctx, typeNamespaceName, found)
			}).Should(Succeed())

			Eventually(
				func(g Gomega) []corev1.Container {
					Expect(k8sClient.Get(ctx, typeNamespaceName, found)).To(Succeed())
					return found.Spec.InitContainers
				},
			).Should(And(
				WithTransform(func(containers []corev1.Container) int { return len(containers) }, Equal(1)),
				WithTransform(func(containers []corev1.Container) string { return containers[0].Image }, Equal("ghcr.io/sigstore/model-transparency-cli:v1.0.1")),
			))
		})
	})
})
