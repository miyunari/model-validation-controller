# Model Validation Controller

This project is a proof of concept based on the [sigstore/model-transperency-cli](https://github.com/sigstore/model-transparency). It offers a Kubernetes/OpenShift controller designed to validate AI models before they are picked up by actual workload. This project provides a webhook that adds an initcontainer to perform model validation. The controller uses a custom resource to define how the models should be validated, such as utilizing [Sigstore](https://www.sigstore.dev/) or public keys.

### Features

- Model Validation: Ensures AI models are validated before they are used by workloads.
- Webhook Integration: A webhook automatically injects an initcontainer into pods to perform the validation step.
- Custom Resource: Configurable `ModelValidation` custom resource to specify how models should be validated. 
    - Supports methods like [Sigstore](https://www.sigstore.dev/), pki or public key validation.

### Prerequisites

- Kubernetes 1.29+ or OpenShift 4.16+
- Proper configuration for model validation (e.g., Sigstore, public keys)
- A signed model (e.g. check the `testdata` or `examples` folder)

### Installation

The controller can be installed in the `model-validation-controller` namespace via [kustomize](https://kustomize.io/).
```bash
kubectl apply -k https://raw.githubusercontent.com/miyunari/model-validation-controller/main/manifests
# or local
kubectl apply -k manifests
```

Run delete to uninstall the controller.
```bash
kubectl delete -k https://raw.githubusercontent.com/miyunari/model-validation-controller/main/manifests
# or local
kubectl delete -k manifests
```

#### Running the Webhook Server Locally with Self-Signed Certs

The webhook server requires TLS certificates. To run it locally using `make run`, you can generate self-signed certs manually:

```
mkdir -p /tmp/k8s-webhook-server/serving-certs

openssl req -x509 -newkey rsa:2048 -nodes \
  -keyout /tmp/k8s-webhook-server/serving-certs/tls.key \
  -out /tmp/k8s-webhook-server/serving-certs/tls.crt \
  -subj "/CN=localhost" \
  -days 365
```

Set the environment variable:
```
export CERT_DIR=/tmp/k8s-webhook-server/serving-certs
```

Alternatively, you can add it to your shell config.

Run the operator:

```
make run
```

This will start the webhook server on https://localhost:9443 using the generated certs.


### Known limitations

The project is at an early stage and therefore has some limitations.

- There is no validation or defaulting for the custom resource.
- The validation is namespace scoped and cannot be used across multiple namespaces.
- No more than one validation resource can be used per namespace.
- There are no status fields for the custom resource.
- The model and signature path must be specified, there is no auto discovery.
- TLS certificates used by the webhook are self generated.

### Usage

First, a ModelValidation CR must be created as follows:
```yaml
apiVersion: rhtas.redhat.com/v1alpha1
kind: ModelValidation
metadata:
  name: demo
spec:
  config:
    sigstoreConfig:
      certificateIdentity: "https://github.com/miyunari/model-validation-controller/.github/workflows/sign-model.yaml@refs/tags/v0.0.2"
      certificateOidcIssuer: "https://token.actions.githubusercontent.com"
  model:
    path: /data/tensorflow_saved_model
    signaturePath: /data/tensorflow_saved_model/model.sig
```

All pods in the namespace where the custom resource exists that have this label `validation.rhtas.redhat.com/ml: "true"` will be validated.
It should be noted that this does not apply to subsequently labeled pods.

```diff
apiVersion: v1
kind: Pod
metadata:
  name: whatever-workload
+  labels:
+    validation.rhtas.redhat.com/ml: "true"
spec:
  restartPolicy: Never
  containers:
  - name: whatever-workload
    image: nginx
    ports:
    - containerPort: 80
    volumeMounts:
    - name: model-storage
      mountPath: /data
  volumes:
  - name: model-storage
    persistentVolumeClaim:
      claimName: models
```

### Examples

The example folder contains two files `prepare.yaml` and `signed.yaml`.

- prepare: contains a persistent volume claim and a job that downloads a signed test model.
```bash
kubectl apply -f https://raw.githubusercontent.com/miyunari/model-validation-controller/main/examples/prepare.yaml
# or local
kubectl apply -f examples/prepare.yaml
```
- signed: contains a model validation manifest for the validation of this model and a demo pod, which is provided with the appropriate label for validation.
```bash
kubectl apply -f https://raw.githubusercontent.com/miyunari/model-validation-controller/main/examples/verify.yaml
# or local
kubectl apply -f examples/verify.yaml
```

After the example installation, the logs of the generated job should show a successful download:
```bash
$ kubectl logs -n testing job/download-extract-model 
Connecting to github.com (140.82.121.3:443)
Connecting to objects.githubusercontent.com (185.199.108.133:443)
saving to '/data/tensorflow_saved_model.tar.gz'
tensorflow_saved_mod  44% |**************                  | 3983k  0:00:01 ETA
tensorflow_saved_mod 100% |********************************| 8952k  0:00:00 ETA
'/data/tensorflow_saved_model.tar.gz' saved
./
./model.sig
./variables/
./variables/variables.data-00000-of-00001
./variables/variables.index
./saved_model.pb
./fingerprint.pb
```

The controller logs should show that a pod has been modified:
```bash
$ kubectl logs -n model-validation-controller deploy/model-validation-controller
time=2025-01-20T22:13:05.051Z level=INFO msg="Starting webhook server on :8080"
time=2025-01-20T22:13:47.556Z level=INFO msg="new request, path: /webhook"
time=2025-01-20T22:13:47.557Z level=INFO msg="Execute webhook"
time=2025-01-20T22:13:47.560Z level=INFO msg="Search associated Model Validation CR" pod=whatever-workload namespace=model-validation-controller
time=2025-01-20T22:13:47.591Z level=INFO msg="construct args"
time=2025-01-20T22:13:47.591Z level=INFO msg="found sigstore config"
```

Finally, the test pod should be running and the injected initcontainer should have been successfully validated.
```bash
$ kubectl logs -n testing whatever-workload model-validation
INFO:__main__:Creating verifier for sigstore
INFO:tuf.api._payload:No signature for keyid f5312f542c21273d9485a49394386c4575804770667f2ddb59b3bf0669fddd2f
INFO:tuf.api._payload:No signature for keyid ff51e17fcf253119b7033f6f57512631da4a0969442afcf9fc8b141c7f2be99c
INFO:tuf.api._payload:No signature for keyid ff51e17fcf253119b7033f6f57512631da4a0969442afcf9fc8b141c7f2be99c
INFO:tuf.api._payload:No signature for keyid ff51e17fcf253119b7033f6f57512631da4a0969442afcf9fc8b141c7f2be99c
INFO:tuf.api._payload:No signature for keyid ff51e17fcf253119b7033f6f57512631da4a0969442afcf9fc8b141c7f2be99c
INFO:__main__:Verifying model signature from /data/model.sig
INFO:__main__:all checks passed
```
In case the workload is modified, is not executed:
```bash
ERROR:__main__:verification failed: the manifests do not match
```

