# Technology Stack

## Package Management
- [govendor](https://github.com/kardianos/govendor) - currently using
- [dep](https://github.com/golang/dep) - no longer using
    - I switched over to using [govendor](https://github.com/kardianos/govendor) on 11/11/2017. I tried to update the 
      **zombiezen.com/go/capnproto2**  dependency from v2.16.0 to v2.17.0 using [dep](https://github.com/golang/dep), but 
      [dep](https://github.com/golang/dep) failed to do it. Even worse - I tried to start over by deleting the vendor directory 
      and reinit **dep**, but dep would fail init. That was a showstopper. It should be easy to start over.

## Logging
- [zerolog](https://github.com/rs/zerolog) - Zero Allocation JSON Logger
- [Collecting Logs with Apache NiFi](https://bryanbende.com/development/2015/05/17/collecting-logs-with-apache-nifi)

## JSON
- [json-iterator](https://github.com/json-iterator/go)
  - A high-performance 100% compatible drop-in replacement of "encoding/json"

## Metrics
- [Prometheus](https://prometheus.io/)
  - monitoring and alerting service
- [Grafana](https://grafana.com/)
  - metrics visualizations
   
## Service Layer
- [capnp RPC](https://capnproto.org/index.html)
  - [go-capnproto2](https://github.com/capnproto/go-capnproto2)
- [gRPC](https://grpc.io/)
  - [gRPC in Go](https://grpc.io/docs/quickstart/go.html)  
  
## Service Mesh
- [envoy](https://www.envoyproxy.io/)

## Messaging Layer
- [NATS](http://nats.io/)
- [nsq](https://github.com/nsqio/nsq)

## Data Layer

### Key-Value stores
- [bbolt](https://github.com/coreos/bbolt) 
  - use case : application configuration

### PKI  
- [easypki](https://github.com/google/easypki)

### Serialization
- [capnp](https://capnproto.org/index.html)

### NewSQL
- [CockroachDB](https://www.cockroachlabs.com/)

### Distributed Key Value Store
- [etcd](https://coreos.com/etcd)

### Graph DB
- [dgraph](https://dgraph.io/)

### Search
- [solr](https://lucene.apache.org/solr/)
  - [banana](https://github.com/lucidworks/banana)
- [bleve](http://www.blevesearch.com/)
  
## Security
- [Let's Encrypt](https://letsencrypt.org/)
  - [acmetool](https://github.com/hlandau/acme)
  
## Containers
- [Lean and Mean Docker containers](https://go.libhunt.com/project/docker-slim)

## Tools
- git gui : [git-cola](http://git-cola.github.io/index.html)
- [lint](https://www.timeferret.com/lint)
  - integrated via **lint/lint_test.go**
    - all code underneath the **oysterpack** package will be linted - excluding test files
    
## Web UI
- [Gowut (Go Web UI Toolkit)](https://github.com/icza/gowut)

## Deployment
- [deis](https://deis.com/)
- [k8s](https://kubernetes.io/)
