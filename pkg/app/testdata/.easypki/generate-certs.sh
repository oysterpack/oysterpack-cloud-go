#!/usr/bin/env bash

. PKI_ENV.sh

# create the PKI directory
mkdir -p $PKI_ROOT

# create the CA authority
easypki create --filename oysterpack --ca dev.oysterpack.com

# create a server wildcard cert that can be deployed to any server
# for client-server connections
easypki create --ca-name oysterpack --dns "*" --ip :: server.dev.oysterpack.com

#  create a client certificate
easypki create --ca-name oysterpack --client client.dev.oysterpack.com