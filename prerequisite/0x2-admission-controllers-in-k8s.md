## Admission controllers in K8s

An admission controller is a piece of code which is intercepted into Kube-Apiserver executed before the resource object updated and after the request is authenticated and  authorized. It's something like the Netfilter's iptables rules PRE-ROUTING which will evaluated the data before it is sent. There are two kind of admission controllers:
* MutuatingAdmissionController
* ValidatingAdmissionController

Semantically speaking MutuatingAdmissionController may modify the resource object and ValidatingAdmissionController will not. And MutuatingAdmissionControll will be executed in the prior of ValidatingAdmissionController.

Generally we could setup Kube-Apiservers's options `--enable-admission-plugins` to specify which admission controller will be enabled.  Standard and Plugin-style admission controllers enabled by default are listed below:
* NamespaceLifecycle
* LimitRanger
* ServiceAccount
* PersistentVolumeClaimResize
* DefaultStorageClass
* DefaultTolerationSeconds
* MutatingAdmissionWebhook
* ValidatingAdmissionWebhook
* ResourceQuota
* Priority

It's a little inflexible using these admission controllers since:
* They need to be compiled into Kube-Apiserver
* They are only configurable when Kube-Apiserver starts up

So there are ways of building admission controllers out of tree and configured at runtime. Admission webhooks are HTTP callbacks that receive admission requests and do something with them, like `MutatingAdmissionWebhook` and `ValidatingAdmissionWebhook`.

Let's take a close look at `ValidatingAdmissionWebhook`. There are 4 steps if we want to setup a ValidatingAdmissionWebhook:
* Write a validating server
* Deploy a validating service
* Configure ValidatingAdmissionWebhook at the runtime
* Authenticate Apiservers

-----
### 1. Write a validating server
A validating server which will handle Kube-Apiserver requests and give responses back,  both in the struct called "AdmissionReview"。Refering to the Kubernetes admission webhook [example](https://github.com/kubernetes/kubernetes/blob/v1.13.0/test/images/webhook/main.go) we could also write own validating server.

### 2. Deploy a validating service
Kube-Apiserver will recognize the webhook service so create a Service for that is a more Kubernetes-Native way.

### 3. Configure ValidatingAdmissionWebhook at the runtime
Then we would like to register the webhook and enable it. Just write a short snippet of code telling Kube-Apiserver what you want:
```yaml
apiVersion: admissionregistration.k8s.io/v1beta1
kind: ValidatingWebhookConfiguration
metadata:
  name: namespace-admission
webhooks:
- clientConfig:
    caBundle: ${CA_BUNDLE}
    service:
      name: etcd-admission
      namespace: default
      path: /etcd
  failurePolicy: Ignore
  name: test.endgame.com
  namespaceSelector: {}
  rules:
  - apiGroups:
    - ""
    apiVersions:
    - v1
    operations:
    - DELETE
    resources:
    - pods
  sideEffects: Unknown
```

The manifest means Kube-Apiserver will send a AdmissionReview to Service etcd-admission.default in the name of test.endgame.com when a Pod is deleted.

### 4. Authenticate ApiServers
> If your admission webhooks require authentication, you can configure the apiservers to use basic auth, bearer token, or a cert to authenticate itself to the webhooks. There are three steps to complete the configuration.

Find more [details about authenticate Kube-Apiserver](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#authenticate-apiservers).

The AdmissionRequest only support https protocol, we'd better give our webhook service a certifacate and let Kubernetes approve it. Such as:

```shell
openssl genrsa -out ${tmpdir}/server-key.pem 2048
openssl req -new -key ${tmpdir}/server-key.pem -subj "/CN=${title}.${title}.svc" -out ${tmpdir}/server.csr -config ${tmpdir}/csr.conf
```
Finaly create a CertificateSigningRequest and let Kubernetes accept it：
```shell 
kubectl certificate approve ${csrName}
```

-----

Wrap it up, Kelsey gives us a very good demo, please check [this](https://github.com/kelseyhightower/denyenv-validating-admission-webhook) out.

[More details about Admission Controllers](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/)
[More details about Dynamic Admission Controllers](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#admission-webhooks)
