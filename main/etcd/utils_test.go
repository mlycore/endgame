package etcd

import (
	"fmt"
	"testing"

	"github.com/mlycore/endgame/main/kubernetes"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func Test_CheckUnregisterStatus(t *testing.T) {
	client := &kubernetes.TestClient{}

	dataSet := []struct {
		namespace  string
		podName    string
		container  corev1.Container
		result     bool
		memberhash string
	}{
		{
			namespace: "",
			podName:   "",
			container: corev1.Container{
				Name: "etcd",
			},
			result: false,
		},
		{
			namespace: "",
			podName:   "etcd-2",
			container: corev1.Container{
				Name: "etcd",
				Env: []corev1.EnvVar{
					{
						Name:  "INITIAL_CLUSTER_SIZE",
						Value: "3",
					},
					{
						Name:  "SET_NAME",
						Value: "etcd",
					},
				},
			},
			result: true,
		},
	}

	for idx, _ := range dataSet {
		result, _ := checkUnregisterStatus(dataSet[idx].namespace, dataSet[idx].podName, dataSet[idx].container, client)
		fmt.Printf("idx=%d, result=%v\n", idx, result)
		assert.Equal(t, result, dataSet[idx].result, "result should be true")
	}
}
