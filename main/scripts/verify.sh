#! /bin/bash
i=0
while true
do
	echo $i
	kubectl exec -ti etcd-0 -n endgame etcdctl set k$i v$i
	i=$((i + 1))	
done
