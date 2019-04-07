## Building Etcd cluster with StatefulSet

It's a common sense to most of Kubernetes developers (operators too) that we should use StatefulSet to run stateful applications. Briefly speaking StatefulSet solved two key problems who are always need to be considered in a distributed architecture:
* Network Topology
* Storage Topology

Finding a [example](https://github.com/kubernetes/kubernetes/blob/master/test/e2e/testing-manifests/statefulset/etcd/statefulset.yaml) of creating a StatefulSet is easy and remember that there are 3 manifests we need to complete:

1. Headless Service
2. Persistent Volume
3. StatefulSet

------

### 1. Headless Service
A Headless Service is such a Service whose `type` is  ClusterIP at the meanwhile `clusterIP` is None. Here is a example:
```yaml
apiVersion: v1
kind: Service
metadata:
  labels:
    app: etcd
  name: etcd
  namespace: default
spec:
  selector:
    app: etcd
  type: ClusterIP
  clusterIP: None
  ports:
  - name: peer
    port: 2380
    protocol: TCP
    targetPort: 2380
  - name: client
    port: 2379
    protocol: TCP
    targetPort: 2379
  sessionAffinity: None
```
With Headless Service we could identify the pod using a short domain name in the form of `PodName.ServiceName`, such as `etcd-0.etcd`.  With the help of Headless Service we have solved the network topology problem. 

[More details about Headless Service](https://kubernetes.io/docs/concepts/services-networking/service/#headless-services).

### 2. Persistent Volume
As its name saying, Persistent Volume provides an durable data storage. We could consider it as an abstraction of storage and placed with Kubernetes first-class resources. Operators who are administrators knowing the volumes of cluster and actual storage system should create Persistent Volumes with specific size. Developers who run applications want persistent storage only need to claim a persistent volume(Persistent Volume Claim). Then the Persistent Volume Claim and the Persistent Volume get bounded, everything goes well. 

What we talked about in the last paragraph is called "Static Provisioning", there also have "Dynamic Provisioning". Using StorageClass specific storage provider who will create a corresponding PV once a PVC is created. That truly save operators life.

Here's a PVC manifests:
```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  annotations:
    pv.kubernetes.io/bind-completed: "yes"
    pv.kubernetes.io/bound-by-controller: "yes"
    volume.beta.kubernetes.io/storage-provisioner: kubernetes.io/gce-pd
  finalizers:
  - kubernetes.io/pvc-protection
  labels:
    app: etcd
  name: datadir-etcd-0
  namespace: default
spec:
  accessModes:
  - ReadWriteOnce
  dataSource: null
  resources:
    requests:
      storage: 1Gi
  storageClassName: standard
  volumeName: pvc-5df900b1-58d2-11e9-aba2-42010a920241
status:
  accessModes:
  - ReadWriteOnce
  capacity:
    storage: 1Gi
  phase: Bound
```

PV manifests:
```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  annotations:
    kubernetes.io/createdby: gce-pd-dynamic-provisioner
    pv.kubernetes.io/bound-by-controller: "yes"
    pv.kubernetes.io/provisioned-by: kubernetes.io/gce-pd
  finalizers:
  - kubernetes.io/pv-protection
  labels:
    failure-domain.beta.kubernetes.io/region: asia-northeast1
    failure-domain.beta.kubernetes.io/zone: asia-northeast1-a
  name: pvc-5df900b1-58d2-11e9-aba2-42010a920241
spec:
  accessModes:
  - ReadWriteOnce
  capacity:
    storage: 1Gi
  claimRef:
    apiVersion: v1
    kind: PersistentVolumeClaim
    name: datadir-etcd-0
    namespace: default
    fsType: ext4
    pdName: gke-endgame-bbcdaa73-d-pvc-5df900b1-58d2-11e9-aba2-42010a920241
  nodeAffinity:
    required:
      nodeSelectorTerms:
      - matchExpressions:
        - key: failure-domain.beta.kubernetes.io/zone
          operator: In
          values:
          - asia-northeast1-a
        - key: failure-domain.beta.kubernetes.io/region
          operator: In
          values:
          - asia-northeast1
  persistentVolumeReclaimPolicy: Delete
  storageClassName: standard
status:
  phase: Bound
```
Storage Class manifests:
```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  annotations:
    storageclass.beta.kubernetes.io/is-default-class: "true"
  creationTimestamp: 2019-04-04T02:02:09Z
  labels:
    addonmanager.kubernetes.io/mode: EnsureExists
    kubernetes.io/cluster-service: "true"
  name: standard
parameters:
  type: pd-standard
provisioner: kubernetes.io/gce-pd
reclaimPolicy: Delete
volumeBindingMode: Immediate
```

[More details about Persistent Volumes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/)

### 3. StatefulSet

We should treat our applications as Cattles but not Pets, however, we should use StatefulSet if it's really a pet.

StatefulSet is mostly like Deployment with some strategy changes. 
* StatefulSet doesn't use ReplicaSet to control Pods
* Pods are different in the eyes of StatefulSet
* StatefulSet create Pods in a specific order

#### PodManagementPolicy
>PodManagementPolicy controls how pods are created during initial scale up, when replacing pods on nodes, or when scaling down. The default policy is`OrderedReady`, where pods are created in increasing order (pod-0, then pod-1, etc) and the controller will wait until each pod is ready before continuing. When scaling down, the pods are removed in the opposite order. The alternative policy is `Parallel` which will create pods in parallel to match the desired scale without waiting, and on scale down will delete all pods at once.

#### UpdateStrategy
> UpdateStrategy indicates the StatefulSetUpdateStrategy that will be employed to update Pods in the StatefulSet when a revision is made to Template.

> * `RollingUpdateStatefulSetStrategyType` indicates that update will be applied to all Pods in the StatefulSet with respect to the StatefulSet ordering constraints. When a scale operation is performed with this strategy, new Pods will be created from the specification version indicated by the StatefulSet's updateRevision. `Partition` indicates the ordinal at which the StatefulSet should be partitioned.
> * `OnDeleteStatefulSetStrategyType` triggers the legacy behavior. Version tracking and ordered rolling restarts are disabled. Pods are recreated from the StatefulSetSpec when they are manually deleted. When a scale operation is performed with this strategy,specification version indicated by the StatefulSet's currentRevision.

StatefulSets manifests:
```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  labels:
    app: etcd
  name: etcd
  namespace: default
spec:
  podManagementPolicy: OrderedReady
  replicas: 1
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      app: etcd
  serviceName: etcd
  template:
    metadata:
      labels:
        app: etcd
      name: etcd
    spec:
      containers:
      - command:
        - /bin/sh
        - -ec
        - |
          HOSTNAME=$(hostname)
          collect_member() {
              while ! etcdctl member list &>/dev/null; do sleep 1; done
              etcdctl member list | grep http://${HOSTNAME}.${SET_NAME}:2380 | cut -d':' -f1 | cut -d'[' -f1 > /var/run/etcd/member_id
              exit 0
          }
          eps() {
              EPS=""
              for i in $(seq 0 $((${INITIAL_CLUSTER_SIZE} - 1))); do
                  EPS="${EPS}${EPS:+,}http://${SET_NAME}-${i}.${SET_NAME}:2379"
              done
              echo ${EPS}
          }
          member_hash() {
              etcdctl member list | grep http://${HOSTNAME}.${SET_NAME}:2380 | cut -d':' -f1 | cut -d'[' -f1
          }
          if [ -e /var/run/etcd/default.etcd ]; then
              echo "Re-joining etcd member"
              member_id=$(cat /var/run/etcd/member_id)
              ETCDCTL_ENDPOINT=$(eps) etcdctl member update ${member_id} http://${HOSTNAME}.${SET_NAME}:2380
              exec etcd --name ${HOSTNAME} \
                  --listen-peer-urls http://0.0.0.0:2380 \
                  --listen-client-urls http://0.0.0.0:2379 \
                  --advertise-client-urls http://${HOSTNAME}.${SET_NAME}:2379 \
                  --data-dir /var/run/etcd/default.etcd
          fi
          SET_ID=${HOSTNAME:5:${#HOSTNAME}}
          if [ "${SET_ID}" -ge ${INITIAL_CLUSTER_SIZE} ]; then
              export ETCDCTL_ENDPOINT=$(eps)
              MEMBER_HASH=$(member_hash)
              if [ -n "${MEMBER_HASH}" ]; then
                  etcdctl member remove ${MEMBER_HASH}
              fi
              echo "Adding new member"
              etcdctl member add ${HOSTNAME} http://${HOSTNAME}.${SET_NAME}:2380 | grep "^ETCD_" > /var/run/etcd/new_member_envs
              if [ $? -ne 0 ]; then
                  echo "Exiting"
                  rm -f /var/run/etcd/new_member_envs
                  exit 1
              fi
              cat /var/run/etcd/new_member_envs
              source /var/run/etcd/new_member_envs
              collect_member &
              exec etcd --name ${HOSTNAME} \
                  --listen-peer-urls http://0.0.0.0:2380 \
                  --listen-client-urls http://0.0.0.0:2379 \
                  --advertise-client-urls http://${HOSTNAME}.${SET_NAME}:2379 \
                  --data-dir /var/run/etcd/default.etcd \
                  --initial-advertise-peer-urls http://${HOSTNAME}.${SET_NAME}:2380 \
                  --initial-cluster ${ETCD_INITIAL_CLUSTER} \
                  --initial-cluster-state ${ETCD_INITIAL_CLUSTER_STATE}
          fi
          for i in $(seq 0 $((${INITIAL_CLUSTER_SIZE} - 1))); do
              while true; do
                  echo "Waiting for ${SET_NAME}-${i}.${SET_NAME} to come up"
                  ping -W 1 -c 1 ${SET_NAME}-${i}.${SET_NAME} > /dev/null && break
                  sleep 1s
              done
          done
          PEERS=""
          for i in $(seq 0 $((${INITIAL_CLUSTER_SIZE} - 1))); do
              PEERS="${PEERS}${PEERS:+,}${SET_NAME}-${i}=http://${SET_NAME}-${i}.${SET_NAME}:2380"
          done
          collect_member &
          exec etcd --name ${HOSTNAME} \
              --initial-advertise-peer-urls http://${HOSTNAME}.${SET_NAME}:2380 \
              --listen-peer-urls http://0.0.0.0:2380 \
              --listen-client-urls http://0.0.0.0:2379 \
              --advertise-client-urls http://${HOSTNAME}.${SET_NAME}:2379 \
              --initial-cluster-token etcd-cluster-1 \
              --initial-cluster ${PEERS} \
              --initial-cluster-state new \
              --data-dir /var/run/etcd/default.etcd
        env:
        - name: INITIAL_CLUSTER_SIZE
          value: "3"
        - name: SET_NAME
          value: etcd
        image: k8s.gcr.io/etcd:3.2.24
        imagePullPolicy: Always
        name: etcd
        ports:
        - containerPort: 2380
          name: peer
          protocol: TCP
        - containerPort: 2379
          name: client
          protocol: TCP
        resources:
          requests:
            cpu: 100m
            memory: 512Mi
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /var/run/etcd
          name: datadir
      dnsPolicy: ClusterFirst
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      terminationGracePeriodSeconds: 30
  updateStrategy:
    rollingUpdate:
      partition: 0
    type: RollingUpdate
  volumeClaimTemplates:
  - metadata:
      creationTimestamp: null
      name: datadir
    spec:
      accessModes:
      - ReadWriteOnce
      dataSource: null
      resources:
        requests:
          storage: 1Gi
    status:
      phase: Pending
```
[More details about StatefulSets](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/)

