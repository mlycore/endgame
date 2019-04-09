package main

import (
	"net/http"
	"os"
	"strings"

	"github.com/mlycore/endgame/main/etcd"
	"github.com/mlycore/endgame/main/kubernetes"

	mlog "github.com/maxwell92/gokits/log"
)

var log = mlog.Log

func init() {
	var level string
	if level = os.Getenv("LOGLEVEL"); strings.EqualFold(level, "") {
		level = "INFO"
	}
	log.SetLevelByName(level)
}

func main() {
	c := kubernetes.KubernetesClientset()
	etcd := &etcd.EtcdHandler{
		Client: c.Clientset,
	}
	http.HandleFunc("/etcd", etcd.GracefulStop)
	http.ListenAndServeTLS(":443", "../certs/server-cert.pem", "../certs/server-key.pem", nil)
}
