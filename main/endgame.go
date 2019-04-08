package main

import (
	"flag"
	"net/http"
	"os"
	"strings"

	"path/filepath"

	mlog "github.com/maxwell92/gokits/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var log = mlog.Log
var clientset *kubernetes.Clientset
var config *rest.Config

func main() {
	var level string
	if level = os.Getenv("LOGLEVEL"); strings.EqualFold(level, "") {
		level = "INFO"
	}
	log.SetLevelByName(level)

	var kubeconfig *string
	var err error
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatalf("read config from flags error: %s", err)
	}
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("init client with config error: %s", err)
	}

	http.HandleFunc("/etcd", etcdHandler)
	http.ListenAndServeTLS(":443", "../certs/server-cert.pem", "../certs/server-key.pem", nil)
}
