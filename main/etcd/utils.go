package etcd

import (
	"encoding/json"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/mlycore/endgame/main/kubernetes"
	"k8s.io/api/admission/v1beta1"
)

// CheckUnregisterStatus helps check if the node finished unregister work
func checkUnregisterStatus(namespace, podName, containerName string, client kubernetes.KubeClient) (bool, string) {
	stdout, stderr, _ := client.ExecInPod(namespace, podName, containerName, []string{"etcdctl", "member", "list"})
	log.Tracef("namespace=%s, podName=%s, containerName=%s, stdout=%v, stderr=%v", namespace, podName, containerName, stdout, stderr)

	var memberhash string
	hostname := podName
	for _, s := range stdout {
		if strings.Contains(s, "name="+hostname) {
			log.Tracef("member list s: %v", s)
			memberhash = strings.Split(s, ":")[0]
			return false, memberhash
		}
	}

	return true, ""
}

// NewAdmissionReview will make any AdmissionReview
func NewAdmissionReview(allow bool, message string) *v1beta1.AdmissionReview {
	return &v1beta1.AdmissionReview{
		Response: &v1beta1.AdmissionResponse{
			Allowed: allow,
			Result: &metav1.Status{
				Message: message,
			},
		},
	}
}

// EncodeAdmissionReview is used for json marshal the admission review response
func EncodeAdmissionReview(ar *v1beta1.AdmissionReview) []byte {
	resp, err := json.Marshal(ar)
	if err != nil {
		log.Errorf("response marshal error: %s", err)
		return nil
	}
	return resp
}

// NewAdmissionReviewError used for make error AdmissionReview quickly
func NewAdmissionReviewError(err error) []byte {
	ar := NewAdmissionReview(false, fmt.Sprintf("%s", err))
	return EncodeAdmissionReview(ar)
}
