# Technology Stack

## Package Management
- [dep](https://github.com/golang/dep)

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
- [gRPC](https://grpc.io/)
  - [gRPC in Go](https://grpc.io/docs/quickstart/go.html)

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
