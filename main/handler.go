package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Etcd scale down Pod validating admission
func etcdHandler(w http.ResponseWriter, r *http.Request) {
	req := &v1beta1.AdmissionReview{}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("read request error: %s", err)
		respAr := requestError(err)
		resp := admissionReviewEncoding(respAr)
		w.Write(resp)
		return
	}

	if err := json.Unmarshal(data, &req); err != nil {
		log.Errorf("read request error: %s", err)
		respAr := requestError(err)
		resp := admissionReviewEncoding(respAr)
		w.Write(resp)
		return
	}

	// Verify if it is a AdmissionReviewRequest about Pods
	podResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	if req.Request.Resource != podResource {
		log.Errorf("not a Pod admission request")
		respAr := requestError(errors.New("not a Pod admission request"))
		resp := admissionReviewEncoding(respAr)
		w.Write(resp)
		return
	}

	pod, err := clientset.CoreV1().Pods(req.Request.Namespace).Get(req.Request.Name, metav1.GetOptions{})
	if err != nil {
		log.Errorf("get requested pod error: %s", err)
		respAr := requestError(errors.New("not a Pod admission request"))
		resp := admissionReviewEncoding(respAr)
		w.Write(resp)
		return
	}

	reviewResp := makeAdmissionReview(true, "")
	for _, container := range pod.Spec.Containers {
		if "etcd" == container.Name {
			// Check if this node finished unregister work.
			// If done will allow this AdmissionReview request,
			// otherwise will not.
			// if checkUnregisterStatus() {
			//     break
			// }

			log.Tracef("admission review request: pod=%s, namespace=%s, operation=%s, uid=%s", time.Now().String(), req.Request.Name, req.Request.Namespace, req.Request.Operation, req.Request.UID)
			// ValidatingAdmissionWebhook will receive two requests,
			// one for object turns into Terminating (set the DeletionTimestamp),
			// another for the object purged,
			if pod.DeletionTimestamp == nil {
				hostname := pod.Name
				ok, memberhash := checkUnregisterStatus(pod.Namespace, pod.Name, container.Name)
				if ok {
					break
				}
				log.Tracef("hostname=%s, memberhash=%s", hostname, memberhash)
				stdout, stderr, _ := execInPod(pod.Namespace, pod.Name, container.Name, []string{"etcdctl", "member", "remove", memberhash})
				log.Tracef("amespace=%s, podName=%s, ncontainerName=%s, stdout=%v, stderr=%v", pod.Name, pod.Namespace, container.Name, stdout, stderr)
			}
			break
		}
	}
	if reviewResp.Response.Allowed {
		log.Infof("Pod %s validate admission succeed", pod.Name)
	} else {
		log.Infof("Pod %s validate admission failed", pod.Name)
	}

	resp := admissionReviewEncoding(reviewResp)
	w.Write(resp)
}
