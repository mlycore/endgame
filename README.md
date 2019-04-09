# endgame
![](https://goreportcard.com/badge/github.com/mlycore/endgame)

## Background

Kubernetes's Pod Lifecycle PreStop hook mechanism is not a "one-fits-all" solution. We may need to write our own CustomResourceDefinitions and Controllers (or Operators) sometimes. However it's always intelligence comsuming and need a lot of work for maintainance. We want a simpler and easy to understand solution. With the help of Kubernetes Apiserver Validating Admission Webhook, we could 'hijack' each Pod DELETE requests and check if it has finished rest work and could be removed safely from the stateful cluster, like etcd.

## Solution

### Design 

First of all, we need a web server which will receive the Admission Request from Kube-Apiserver and make sure it's the exact Pod DELETE request we want. Notice there will be two DELETE request, one turning the Pod into Terminating and set the DeleteTimestamp field, another for removing the Pod object. We will take care of the first request. 

Then we should check if it's a valid request that this node is still registered in the cluster. If it's registered, we start a remote command execution to unregister it, maybe do some cleanup work, and disallow the Admission Request we received.

Finally when the node has already unregistered, we allow the Admission Requets, and the Pod will be purged.

When scaling down, the desired state is going to mismatch the actual state if the requests are blocked by the Validating Admission Webhook, several subsequential Pod DELETE requests will be sent in the reconcile loops of the controller. All of the requests will be handled by the webhook and only when the node is truely unregistered, the last request should be allowed and the Pod will be purged. 

This sequence diagram is placed below
![](https://github.com/mlycore/endgame/blob/master/pics/seq.png)

More prerequsition contents please check *prerequisite* directory.

### Verification

Use `make build` in *main/* directory to build up the experiment envrionment if using GKE (modifications are needed if using other cloud vendors). After the 3 etcd Pods turns into Running, we run the `endgame` server in *main/bin* and start to send simulate continuous data write / read request to the etcd cluster with script *main/scripts* `verify.sh`.

The results turn out good and the etcd cluster always works.

![](https://github.com/mlycore/endgame/blob/master/pics/result.png)

Use `make clean` to cleanup.

## Undetermined Issues

* How to ensue the node is really unregistered from the cluster ?
* Will this solution introduce more uncertainty when the cluster has problems ?
* How to extend this to fit other stateful distributed systems ?
