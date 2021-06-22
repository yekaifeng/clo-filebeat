# clo-filebeat
Re-implement the collector in Openshift cluster logging from fluentd to filebeat

# How to rebuild filebeat docker image

~~~
cd clo-filebeat/
mkdir -p ${GOPATH}/src/github.com/elastic
ln -s beats-7.12.1 ${GOPATH}/src/github.com/elastic/beats
cd beats-7.12.1/filebeat/
export GOPROXY=https://goproxy.cn,direct
make

docker build -t docker.io/kennethye/filebeat-oss:7.12.1-rebuild -f Dockerfile-rebuild .
~~~

# How to deploy filebeat rebuild image

- Install openshift cluster logging stack （EFK）

set a new node selector for fluentd in ClusterLogging CRD
~~~
oc edit clusterlogging instance

apiVersion: "logging.openshift.io/v1"
kind: "ClusterLogging"
metadata:
  name: "instance"
  namespace: "openshift-logging"
spec:
  managementState: "Managed"
  collection:
    logs:
      type: "fluentd"
      fluentd:
        nodeSelector:
          #node-role.kubernetes.io/worker: ""
          fluentd: "true"
        resources:
          limits:
            cpu: 4
            memory: 1Gi
          requests:
            cpu: 100m
            memory: 512Mi
~~~

- With above configuration, no fluentd pod will be deployed on any worker nodes
- deploy filebeat as new log collector

~~~
oc get secret fluentd -n openshift-logging -o yaml > secret-fluentd.yaml
sed -i s/openshift-logging/kube-system/g secret-fluentd.yaml

oc apply -f secret-fluentd.yaml
oc apply -f deploy/filebeat-oss-kubernetes-7.12.1-es-output.yaml
~~~




