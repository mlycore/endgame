package etcd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/mlycore/endgame/main/kubernetes"
)

// CheckUnregisterStatus helps check if the node finished unregister work
// When the command executed failed if the node should be considered unregistered successfully?
// If the container has already been stopped or restarting, the command is going to be failed, means nothing
// about the register status of this node.
func checkUnregisterStatus(namespace, podName string, container corev1.Container, client kubernetes.KubeClient) (bool, string) {
	var initialClusterSize, setName string
	for _, e := range container.Env {
		if "INITIAL_CLUSTER_SIZE" == e.Name {
			initialClusterSize = e.Value
		}
		if "SET_NAME" == e.Name {
			setName = e.Value
		}
	}

	hostname := podName
	var eps string
	size, _ := strconv.Atoi(initialClusterSize)
	for i := 0; i < size; i++ {
		eps += fmt.Sprintf("http://etcd-%d.%s:2379,", i, setName)
	}
	eps = strings.TrimSuffix(eps, ",")
	log.Tracef("eps=%s\n", eps)

	stdout, stderr, _ := client.ExecInPod(namespace, podName, container.Name, []string{"etcdctl", "--endpoints", eps, "member", "list"})
	log.Tracef("namespace=%s, podName=%s, containerName=%s, stdout=%v, stderr=%v", namespace, podName, container.Name, stdout, stderr)

	var memberhash string
	for _, s := range stdout {
		if strings.Contains(s, hostname) {
			log.Tracef("member list s: %v", s)
			memberhash = strings.Split(s, ":")[0]
			if strings.Contains(memberhash, " ") {
				return true, ""
			}
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
