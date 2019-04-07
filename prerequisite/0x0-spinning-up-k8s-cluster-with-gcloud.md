## Spinning up Kubernetes cluster with Google Cloud

Since there are a lot of cloud vendors who provide Kubernetes cluster setup, Google Compute Platform would be the first place one to choose.

Spinning up a Kubernetes cluster won't be hard if we just follow the [Kelsey's Kubernetes the hard way](https://github.com/kelseyhightower/kubernetes-the-hard-way). 

Another GCP approach is using Google Kubernetes Engine we cloud also run a Kubernetes Cluster in a few minutes. 

Sometimes we may run into some kind of issues like:
* **GCP doesn't provision PVs dynamically if we just create a pvc**. That's true because we need to setup StorageClass for dynamic provisioning or create pv manually. More details please check [this](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) out.

Let the vendor handle the most boring part and we will focus on tackling the main problem. 


