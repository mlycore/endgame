build:
	kubectl apply -f manifests/etcd-namespace.yaml
	kubectl apply -f manifests/etcd-service.yaml
	kubectl apply -f manifests/etcd-statefulset.yaml
	/bin/bash -C scripts/make-certs.sh
	/bin/bash -C scripts/ca-bundle.sh
	kubectl apply -f manifests/validatingwebhookservice.yaml
	kubectl apply -f manifests/validatingwebhookendpoints.yaml
	kubectl apply -f manifests/validatingwebhookconfiguration.yaml
	go build -ldflags="-s -w" -o bin/endgame endgame.go utils.go handler.go
clean:
	kubectl delete -f manifests/etcd-namespace.yaml
	kubectl delete -f manifests/validatingwebhookconfiguration.yaml
	sed -i "s/caBundle: .*/caBundle: {CA_BUNDLE}/" manifests/validatingwebhookconfiguration.yaml > /dev/null
	rm -rf certs
	rm -rf bin
