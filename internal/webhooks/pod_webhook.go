package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sigstore/model-validation-operator/internal/constants"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/go-logr/logr"
	"github.com/sigstore/model-validation-operator/api/v1alpha1"
)

// NewPodInterceptor creates a new pod mutating webhook to be registered
func NewPodInterceptor(c client.Client, decoder admission.Decoder) webhook.AdmissionHandler {
	return &podInterceptor{
		client:  c,
		decoder: decoder,
	}
}

// +kubebuilder:webhook:path=/mutate-v1-pod,mutating=true,failurePolicy=fail,groups="",resources=pods,sideEffects=None,verbs=create;update,versions=v1,name=pods.model-validation.rhtas.redhat.com,admissionReviewVersions=v1

// +kubebuilder:rbac:groups=rhtasv1alpha1,resources=ModelValidation,verbs=get;list;watch
// +kubebuilder:rbac:groups=rhtasv1alpha1,resources=ModelValidation/status,verbs=get;update;patch

// podInterceptor extends pods with Model Validation Init-Container if annotation is specified.
type podInterceptor struct {
	client  client.Client
	decoder admission.Decoder
}

// Handle extends pods with Model Validation Init-Container if annotation is specified.
func (p *podInterceptor) Handle(ctx context.Context, req admission.Request) admission.Response {
	logger := log.FromContext(ctx)
	logger.Info("Execute webhook")
	pod := &corev1.Pod{}

	if err := p.decoder.Decode(req, pod); err != nil {
		logger.Error(err, "failed to decode pod")
		return admission.Errored(http.StatusBadRequest, err)
	}
	// TODO: check in webhook config
	if v := pod.Labels["validation.rhtas.redhat.com/ml"]; v != "true" {
		return admission.Allowed("no annoation found, no action needed")
	}

	logger.Info("Search associated Model Validation CR", "pod", pod.Name, "namespace", pod.Namespace)
	rhmvList := &v1alpha1.ModelValidationList{}
	if err := p.client.List(ctx, rhmvList); err != nil {
		msg := "failed to get the ModelValidation Spec, skipping injection"
		logger.Error(err, msg)
		return admission.Errored(http.StatusNotFound, err)
	}

	got := len(rhmvList.Items)
	if got != 1 {
		err := fmt.Errorf("got no or to many specs, expect: 1, got: %d", got)
		logger.Error(err, "skip injection")
		return admission.Errored(http.StatusBadRequest, err)
	}
	rhmv := rhmvList.Items[0]
	// NOTE: check if validation sidecar is already injected. Then no action needed.
	for _, c := range pod.Spec.InitContainers {
		if c.Name == modelValidationInitContainerName {
			return admission.Allowed("validation exists, no action needed")
		}
	}

	args := []string{"verify"}
	args = append(args, validationConfigToArgs(logger, rhmv.Spec.Config, rhmv.Spec.Model.SignaturePath)...)
	args = append(args, rhmv.Spec.Model.Path)

	pp := pod.DeepCopy()
	vm := []corev1.VolumeMount{}
	for _, c := range pod.Spec.Containers {
		vm = append(vm, c.VolumeMounts...)
	}
	pp.Spec.InitContainers = append(pp.Spec.InitContainers, corev1.Container{
		Name:            modelValidationInitContainerName,
		ImagePullPolicy: corev1.PullAlways,
		Image:           constants.ModelTransparencyCliImage,
		Command:         []string{"/usr/local/bin/model_signing"},
		Args:            args,
		VolumeMounts:    vm,
	})
	marshaledPod, err := json.Marshal(pp)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

func validationConfigToArgs(logger logr.Logger, cfg v1alpha1.ValidationConfig, signaturePath string) []string {
	logger.Info("construct args")
	res := []string{}
	if cfg.SigstoreConfig != nil {
		logger.Info("found sigstore config")
		res = append(res,
			"sigstore",
			fmt.Sprintf("--signature=%s", signaturePath),
			"--identity", cfg.SigstoreConfig.CertificateIdentity,
			"--identity_provider", cfg.SigstoreConfig.CertificateOidcIssuer,
		)
		return res
	}

	if cfg.PrivateKeyConfig != nil {
		logger.Info("found private-key config")
		res = append(res,
			"key",
			fmt.Sprintf("--signature=%s", signaturePath),
			"--public_key", cfg.PrivateKeyConfig.KeyPath,
		)
		return res
	}

	if cfg.PkiConfig != nil {
		logger.Info("found pki config")
		res = append(res,
			"certificate",
			fmt.Sprintf("--signature=%s", signaturePath),
			"--certificate_chain", cfg.PkiConfig.CertificateAuthority,
		)
		return res
	}
	logger.Info("missing validation config")
	return []string{}
}

const modelValidationInitContainerName = "model-validation"
