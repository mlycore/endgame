package etcd

import (
	"fmt"
	"testing"

	"github.com/mlycore/endgame/main/kubernetes"
	"github.com/stretchr/testify/assert"
)

func Test_CheckUnregisterStatus(t *testing.T) {
	client := &kubernetes.TestClient{}

	dataSet := []struct {
		namespace     string
		podName       string
		containerName string
		result        bool
		memberhash    string
	}{
		{
			namespace:     "",
			podName:       "",
			containerName: "",
			result:        false,
		},
		{
			namespace:     "",
			podName:       "etcd-2",
			containerName: "",
			result:        true,
		},
	}

	for idx, _ := range dataSet {
		result, _ := checkUnregisterStatus(dataSet[idx].namespace, dataSet[idx].podName, dataSet[idx].containerName, client)
		fmt.Printf("idx=%d, result=%v\n", idx, result)
		assert.Equal(t, result, dataSet[idx].result, "result should be true")
	}
}
