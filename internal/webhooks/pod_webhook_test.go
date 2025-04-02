package webhooks

import (
	"context"
	"encoding/json"
	"testing"

	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/stretchr/testify/assert"
)

func Test_PodInterceptor_Handle(t *testing.T) {
	const req = `{"kind":"AdmissionReview","apiVersion":"admission.k8s.io/v1","request":{"uid":"c33e1334-1a38-4684-aff8-e8c366ea89b9","kind":{"group":"","version":"v1","kind":"Pod"},"resource":{"group":"","version":"v1","resource":"pods"},"requestKind":{"group":"","version":"v1","kind":"Pod"},"requestResource":{"group":"","version":"v1","resource":"pods"},"name":"ollama","namespace":"model-validation-controller","operation":"CREATE","userInfo":{"username":"admin","groups":["system:masters","system:authenticated"],"extra":{"authentication.kubernetes.io/credential-id":["X509SHA256=acb312b9049f7fbeb8788001d7e61e145f10df1e4a752d1ef2caa3097d233246"]}},"object":{"kind":"Pod","apiVersion":"v1","metadata":{"name":"ollama","namespace":"model-validation-controller","creationTimestamp":null,"annotations":{"kubectl.kubernetes.io/last-applied-configuration":"{\"apiVersion\":\"v1\",\"kind\":\"Pod\",\"metadata\":{\"annotations\":{\"validation.rhtas.redhat.com/ml\":\"true\"},\"name\":\"ollama\",\"namespace\":\"model-validation-controller\"},\"spec\":{\"containers\":[{\"image\":\"ollama/ollama\",\"name\":\"ollama\",\"ports\":[{\"containerPort\":11434}]}]}}\n","validation.rhtas.redhat.com/ml":"true"},"managedFields":[{"manager":"kubectl-client-side-apply","operation":"Update","apiVersion":"v1","time":"2025-01-18T20:43:09Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:kubectl.kubernetes.io/last-applied-configuration":{},"f:validation.rhtas.redhat.com/ml":{}}},"f:spec":{"f:containers":{"k:{\"name\":\"ollama\"}":{".":{},"f:image":{},"f:imagePullPolicy":{},"f:name":{},"f:ports":{".":{},"k:{\"containerPort\":11434,\"protocol\":\"TCP\"}":{".":{},"f:containerPort":{},"f:protocol":{}}},"f:resources":{},"f:terminationMessagePath":{},"f:terminationMessagePolicy":{}}},"f:dnsPolicy":{},"f:enableServiceLinks":{},"f:restartPolicy":{},"f:schedulerName":{},"f:securityContext":{},"f:terminationGracePeriodSeconds":{}}}}]},"spec":{"volumes":[{"name":"kube-api-access-v9hhf","projected":{"sources":[{"serviceAccountToken":{"expirationSeconds":3607,"path":"token"}},{"configMap":{"name":"kube-root-ca.crt","items":[{"key":"ca.crt","path":"ca.crt"}]}},{"downwardAPI":{"items":[{"path":"namespace","fieldRef":{"apiVersion":"v1","fieldPath":"metadata.namespace"}}]}}],"defaultMode":420}}],"containers":[{"name":"ollama","image":"ollama/ollama","ports":[{"containerPort":11434,"protocol":"TCP"}],"resources":{},"volumeMounts":[{"name":"kube-api-access-v9hhf","readOnly":true,"mountPath":"/var/run/secrets/kubernetes.io/serviceaccount"}],"terminationMessagePath":"/dev/termination-log","terminationMessagePolicy":"File","imagePullPolicy":"Always"}],"restartPolicy":"Always","terminationGracePeriodSeconds":30,"dnsPolicy":"ClusterFirst","serviceAccountName":"default","serviceAccount":"default","securityContext":{},"schedulerName":"default-scheduler","tolerations":[{"key":"node.kubernetes.io/not-ready","operator":"Exists","effect":"NoExecute","tolerationSeconds":300},{"key":"node.kubernetes.io/unreachable","operator":"Exists","effect":"NoExecute","tolerationSeconds":300}],"priority":0,"enableServiceLinks":true,"preemptionPolicy":"PreemptLowerPriority"},"status":{}},"oldObject":null,"dryRun":false,"options":{"kind":"CreateOptions","apiVersion":"meta.k8s.io/v1","fieldManager":"kubectl-client-side-apply"}}}`

	var review admissionv1.AdmissionReview
	err := json.Unmarshal([]byte(req), &review)
	assert.NoError(t, err)

	decoder := admission.NewDecoder(scheme.Scheme)
	handler := NewPodInterceptor(fake.NewClientBuilder().Build(), decoder)

	resp := handler.Handle(context.Background(), admission.Request{
		AdmissionRequest: *review.Request,
	})

	assert.NotNil(t, resp)
	assert.True(t, resp.Allowed)
}
