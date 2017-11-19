#!/usr/bin/env bash

. PKI_ENV.sh

# create the PKI directory
mkdir -p $PKI_ROOT

# create the root CA
easypki create --filename oysterpack --ca dev.oysterpack.com

# create a server wildcard cert that can be deployed to any server
# for client-server connections
easypki create --ca-name oysterpack --dns "*" --ip :: server.dev.oysterpack.com

#  create a client certificate
easypki create --ca-name oysterpack --client client.dev.oysterpack.com

# create an intermediate CA
easypki create --ca-name oysterpack --intermediate app.dev.oysterpack.com

# create a server wildcard cert that can be deployed to any server
# for client-server connections
easypki create --ca-name app.dev.oysterpack.com --dns "*" --ip :: server.dev.oysterpack.com

#  create a client certificate
easypki create --ca-name app.dev.oysterpack.com --client client.dev.oysterpack.com