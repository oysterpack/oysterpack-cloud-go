#!/usr/bin/env bash

. PKI_ENV.sh

# create the PKI directory
mkdir -p $PKI_ROOT

# create the root CA
easypki create --filename oysterpack --ca dev.oysterpack.com

# create a server wildcard cert that can be deployed to any server
# for client-server connections
easypki create --ca-name oysterpack server.dev.oysterpack.com

#  create a client certificate
easypki create --ca-name oysterpack --client client.dev.oysterpack.com

# create an intermediate CA
easypki create --ca-name oysterpack --intermediate app.dev.oysterpack.com

# create a server wildcard cert that can be deployed to any server
# for client-server connections
easypki create --ca-name app.dev.oysterpack.com server.dev.oysterpack.com
# service = fef711bb74ee4e13
# app = d113a2e016e12f0f
# domain = ed5cf026e8734361
easypki create --ca-name app.dev.oysterpack.com fef711bb74ee4e13.d113a2e016e12f0f.ed5cf026e8734361

#  create a client certificate
easypki create --ca-name app.dev.oysterpack.com --client client.dev.oysterpack.com
