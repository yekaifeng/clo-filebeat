# clo-filebeat
Re-implement the collector in Openshift cluster logging from fluentd to filebeat

# How to rebuild filebeat

~~~
cd clo-filebeat/
mkdir -p ${GOPATH}/src/github.com/elastic
ln -s beats-7.12.1 ${GOPATH}/src/github.com/elastic/beats
cd beats-7.12.1/filebeat/
export GOPROXY=https://goproxy.cn,direct
make
~~~



