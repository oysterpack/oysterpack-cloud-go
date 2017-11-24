# How to run prometheus locally
1. Install prometheus locally using go (see [https://github.com/prometheus/prometheus](https://github.com/prometheus/prometheus))

        GO15VENDOREXPERIMENT=1 go get github.com/prometheus/prometheus/cmd/prometheus
        
2. prometheus -config.file <prometheus.yml>

   **NOTE**: run prometheus outside of the workspace because by default prometheus will create its data directory in the current working directory.

# Prometheus - Docker Swarm integration - Service Discovery
https://github.com/ContainerSolutions/prometheus-swarm-discovery