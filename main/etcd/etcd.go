package etcd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	mlog "github.com/maxwell92/log"
	"github.com/mlycore/endgame/main/kubernetes"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var log = mlog.Log

// EtcdHandler is responsible for handling etcd graceful stop
type EtcdHandler struct {
	Client kubernetes.KubeClient
}

// GracefulStop will first check if this member is still registered.
// If registered, then triggger etcdctl member remove a with remote command execution.
// Otherwise will give a allow AdmissionReviewResponse.
func (c *EtcdHandler) GracefulStop(w http.ResponseWriter, r *http.Request) {
	req, err := c.ReadAdmissionReview(r)
	if err != nil {
		c.WriteError(w, fmt.Sprintf("%s", err), err)
		return
	}

	// Verify if it is a AdmissionReviewRequest about Pods
	podResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	if req.Request.Resource != podResource {
		c.WriteError(w, "not a Pod admission request", nil)
		return
	}

	pod := c.Client.GetPod(req.Request.Namespace, req.Request.Name)
	if pod == nil {
		c.WriteError(w, "get requested pod error", nil)
		return
	}

	reviewResp := NewAdmissionReview(true, "")
	for _, container := range pod.Spec.Containers {
		if "etcd" == container.Name {
			log.Tracef("admission review request: pod=%s, namespace=%s, operation=%s, uid=%s", req.Request.Name, req.Request.Namespace, req.Request.Operation, req.Request.UID)
			// ValidatingAdmissionWebhook will receive two requests,
			// one for object turns into Terminating (set the DeletionTimestamp),
			// another for the object purged,
			// the first request will trigger a graceful stop procedure.
			if pod.DeletionTimestamp == nil {
				hostname := pod.Name

				// Check if this node finished unregister work.
				// If done will allow this AdmissionReview request,
				// otherwise will not.
				existed, memberhash := checkUnregisterStatus(pod.Namespace, pod.Name, container, c.Client)
				if existed {
					reviewResp.Response.Allowed = false
					break
				}
				log.Tracef("hostname=%s, memberhash=%s", hostname, memberhash)
				stdout, stderr, _ := c.Client.ExecInPod(pod.Namespace, pod.Name, container.Name, []string{"etcdctl", "member", "remove", memberhash})
				log.Tracef("namespace=%s, podName=%s, containerName=%s, stdout=%v, stderr=%v", pod.Name, pod.Namespace, container.Name, stdout, stderr)
			}
			break
		}
	}
	if reviewResp.Response.Allowed {
		log.Infof("Pod %s validate admission succeed", pod.Name)
	} else {
		log.Infof("Pod %s validate admission failed", pod.Name)
	}

	resp := EncodeAdmissionReview(reviewResp)
	w.Write(resp)
}

// ReadAdmissionReview will read object from request and save it to an AdmissionReview
func (c *EtcdHandler) ReadAdmissionReview(r *http.Request) (*v1beta1.AdmissionReview, error) {
	req := &v1beta1.AdmissionReview{}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Debugf("read request error: %s", err)
		return nil, err
	}

	err = json.Unmarshal(data, &req)
	if err != nil {
		log.Debugf("read request error: %s", err)
		return nil, err
	}
	return req, nil
}

// WriteError is used for write error response
func (c *EtcdHandler) WriteError(w http.ResponseWriter, message string, err error) {
	resp := NewAdmissionReviewError(err)
	w.Write(resp)
}
