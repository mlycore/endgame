package kubernetes

import (
	"errors"

	corev1 "k8s.io/api/core/v1"
)

type TestClient struct {
}

func (c *TestClient) GetPod(namespace, podName string) *corev1.Pod {
	return &corev1.Pod{}
}

func (c *TestClient) ExecInPod(namespace, podName, containerName string, commands []string) ([]string, []string, error) {
	if "" == namespace && "" == podName && "" == containerName && commands == nil {
		return nil, nil, errors.New("parameters can't be empty")
	}
	if len(commands) > 3 && "etcdctl" == commands[0] && "member" == commands[1] && "list" == commands[2] {
		return []string{"2e80f96756a54ca9: name=etcd-0 peerURLs=http://etcd-0.etcd:2380 clientURLs=http://etcd-0.etcd:2379 isLeader=false", "7fd61f3f79d97779: name=etcd-1 peerURLs=http://etcd-1.etcd:2380 clientURLs=http://etcd-1.etcd:2379 isLeader=true", "b429c86e3cd4e077: name=etcd-2 peerURLs=http://etcd-2.etcd:2380 clientURLs=http://etcd-2.etcd:2379 isLeader=false"}, []string{""}, nil
	}

	return []string{"stdout"}, []string{"stderr"}, nil
}
