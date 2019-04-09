package kubernetes

import (
	"flag"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/homedir"

	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	mlog "github.com/maxwell92/gokits/log"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var log = mlog.Log

var _ KubeClient = &Client{}

type KubeClient interface {
	ExecInPod(namespace, podName, containerName string, commands []string) ([]string, []string, error)
	GetPod(namespace, podName string) *corev1.Pod
}

type Client struct {
	Config *rest.Config
	*kubernetes.Clientset
}

/*
var config *rest.Config
var clientset *kubernetes.Clientset
*/

// kubernetesClientset create kubernetes clients
func KubernetesClientset() *Client {
	var kubeconfig *string
	var err error
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatalf("read config from flags error: %s", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("init client with config error: %s", err)
	}
	return &Client{
		Config:    config,
		Clientset: clientset,
	}
}

// ExecInPod implements a remote command execution of Pods.
func (c *Client) ExecInPod(namespace, podName, containerName string, commands []string) ([]string, []string, error) {
	req := c.CoreV1().RESTClient().Post().
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

	exec, err := remotecommand.NewSPDYExecutor(c.Config, "POST", req.URL())
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

// GetPod will retrieve a pod from the cluster
func (c *Client) GetPod(namespace, podName string) *corev1.Pod {
	pod, err := c.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
	if err != nil {
		log.Errorf("get pod error: %s", err)
		return nil
	}
	return pod
}
