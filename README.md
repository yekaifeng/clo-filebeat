# clo-filebeat
Re-implement the collector in Openshift cluster logging from fluentd to filebeat

# Design consideration
Openshift EFK stack has 3 kinds of log: audit, infra and app. While it comes to filebeat,
I sill follow this rule to set the "log_type" with those 3 categories. And send them to elasticsearch
corresponding index.

### log_type: audit
Source:
- /var/log/audit/*.log
- /var/log/kube-apiserver/*.log
- /var/log/oauth-apiserver/*.log
- /var/log/openshift-apiserver/*.log

Destination:
- ES index: audit-write

### log_type: infra
Source:
- /var/log/containers/*.log (namespace start with 'openshift-' )

Destination:
- ES index: infra-write

### log_type: app
Source:
- /var/log/containers/*.log (namespace NOT start with 'openshift-' )

Destination:
- ES index: app-write

# Support features
- Collector container and audit logs, then send to Openshift Elasticsearch
- Collector container and audit logs, then send to Standalone Kafka/Redhat MQ Stream
- Support parse json string in logs
- Java Multi lines detection


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
~~~

- For sending logs to Elasticsearch

~~~
oc apply -f deploy/filebeat-oss-kubernetes-7.12.1-es-output.yaml
~~~

- For sending logs to Kafka

~~~
oc apply -f deploy/filebeat-oss-kubernetes-7.12.1-kafka-output.yaml
~~~

- For sending logs to Elasticsearch and Kafka

~~~
oc apply -f deploy/filebeat-oss-kubernetes-7.12.1-es-kafka.yaml
~~~


