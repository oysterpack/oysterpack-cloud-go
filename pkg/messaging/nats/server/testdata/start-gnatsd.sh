#!/usr/bin/env bash

# creates a cluster of 3 nodes

# seed node
gnatsd --tlscert   $PKI_ROOT/oysterpack/certs/nats.dev.oysterpack.com.crt \
       --tlskey    $PKI_ROOT/oysterpack/keys/nats.dev.oysterpack.com.key \
       --tlsverify \
       --tlscacert $PKI_ROOT/oysterpack/certs/oysterpack.crt \
       -p 4222 -m 8222 --cluster tls://localhost:5222 --routes tls://localhost:5222 -D


gnatsd --tlscert   $PKI_ROOT/oysterpack/certs/nats.dev.oysterpack.com.crt \
       --tlskey    $PKI_ROOT/oysterpack/keys/nats.dev.oysterpack.com.key \
       --tlsverify \
       --tlscacert $PKI_ROOT/oysterpack/certs/oysterpack.crt \
       -p 4223 -m 8223 --cluster tls://localhost:5223 --routes tls://localhost:5222 -D


gnatsd --tlscert   $PKI_ROOT/oysterpack/certs/nats.dev.oysterpack.com.crt \
       --tlskey    $PKI_ROOT/oysterpack/keys/nats.dev.oysterpack.com.key \
       --tlsverify \
       --tlscacert $PKI_ROOT/oysterpack/certs/oysterpack.crt \
       -p 4224 -m 8224 --cluster tls://localhost:5224 --routes tls://localhost:5222 -D