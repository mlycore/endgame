package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

// execInPod implements a remote command execution of Pods.
func execInPod(namespace, podName, containerName string, commands []string) ([]string, []string, error) {
	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")
	req.VersionedParams(&corev1.PodExecOptions{
		Container: containerName,
		Command:   commands,
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
	}, scheme.ParameterCodec)

	log.Tracef("remotecmd curl=%v", req.URL())

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		log.Debugf("remotecmd SPDY setup error: %s", err)
		return nil, nil, err
	}
	stdIn := newStringReader([]string{})
	stdOut := new(Writer)
	stdErr := new(Writer)

	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdIn,
		Stdout: stdOut,
		Stderr: stdErr,
		Tty:    false,
	})
	if err != nil {
		log.Errorf("remotecmd stream execute error: %s", err)
		log.Tracef("remotecmd strema execute error: stdout=%v, stderr=%s", stdOut.Str, stdErr.Str)
		return nil, nil, err
	}
	return stdOut.Str, stdErr.Str, nil

}

// makeAdmissionReview will make any AdmissionReview
func makeAdmissionReview(allow bool, message string) *v1beta1.AdmissionReview {
	return &v1beta1.AdmissionReview{
		Response: &v1beta1.AdmissionResponse{
			Allowed: allow,
			Result: &metav1.Status{
				Message: message,
			},
		},
	}
}

// requestError is used for simplify request error handling
func requestError(err error) *v1beta1.AdmissionReview {
	return makeAdmissionReview(false, fmt.Sprintf("%s", err))
}

// admissionReviewEncoding is used for json marshal the admission review response
func admissionReviewEncoding(ar *v1beta1.AdmissionReview) []byte {
	resp, err := json.Marshal(ar)
	if err != nil {
		log.Errorf("response marshal error: %s", err)
		return nil
	}
	return resp
}

func newStringReader(ss []string) io.Reader {
	formattedString := strings.Join(ss, "\n")
	reader := strings.NewReader(formattedString)
	return reader
}

// Writer used for retrieve remotecmd execute results
type Writer struct {
	Str []string
}

func (w *Writer) Write(p []byte) (n int, err error) {
	str := string(p)
	if len(str) > 0 {
		w.Str = append(w.Str, str)
	}
	return len(str), nil
}

// CheckUnregisterStatus helps check if the node finished unregister work
func checkUnregisterStatus(namespace, podName, containerName string) (bool, string) {
	stdout, stderr, _ := execInPod(namespace, podName, containerName, []string{"etcdctl", "member", "list"})
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
