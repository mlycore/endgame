package etcd

import (
	"net/http"
)

type KubeClient interface {
	ExecInPod(namespace, podName, containerName string, commands []string) ([]string, []string, error)
}

var client KubeClient

// Etcd scale down Pod validating admission
func etcdHandler(w http.ResponseWriter, r *http.Request) {
}
