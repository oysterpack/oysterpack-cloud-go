#!/usr/bin/env bash

# create the PKI directory
mkdir -p $PKI_ROOT

# create the CA authority
easypki create --filename oysterpack --ca dev.oysterpack.com

# create a server wildcard cert that can be deployed to any server
# for client-server connections
easypki create --ca-name oysterpack --dns "*" nats.dev.oysterpack.com

# for server-server cluster connections
# NOTE: ip is required for servers to validate each other
# TODO: On docker swarm, test if using the docker swarm network name is sufficient
easypki create --ca-name oysterpack --dns "*" --ip 127.0.0.1 cluster.nats.dev.oysterpack.com

#  create a client certificate
easypki create --ca-name oysterpack --client client.nats.dev.oysterpack.com