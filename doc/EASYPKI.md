# How to generate TLS Certs using [easypki](https://github.com/google/easypki)

The example below shows how to generate certs for TLS. The basic steps are :

1. Create a CA (Certificate Authority)
2. Distribute the CA cert to all servers and clients
3. Generate server certificates signed by the CA
4. Generate client certificates signed by the CA
5. Servers need to be configured to trust the CA to authenticate clients
6. Clients need to be configured to trust the CA to authenticate servers


    # export env variables for your PKI 
    export PKI_ROOT=$GOPATH/.oysterpack/easypki/pki
    export PKI_ORGANIZATION="OysterPack Inc."
    export PKI_ORGANIZATIONAL_UNIT=IT
    export PKI_COUNTRY=US
    export PKI_LOCALITY="Rochester"
    export PKI_PROVINCE="New York"

    # create the PKI directory
    mkdir -p $PKI_ROOT

    # create the CA authority
    easypki create --filename oysterpack --ca ca.oysterpack.com
    
    # create a server wildcard cert that can be deployed to any server 
    easypki create --ca-name oysterpack -dns "*" nats.oysterpack.com   
    
    #  create a client certificate
    easypki create --ca-name oysterpack --client client1.oysterpack.com
 
    # example usage
    gnatsd  --tlscert   $PKI_ROOT/oysterpack/certs/nats.oysterpack.com.crt \
            --tlskey    $PKI_ROOT/oysterpack/keys/nats.oysterpack.com.key \
            --tlsverify \
            --tlscacert $PKI_ROOT/oysterpack/certs/oysterpack.crt \
            -p 4443

The CA directory structure will look like :

    pki
      |-oysterpack
        |-certs                             # public certs
          |-client1.oysterpack.com.crt        # client
          |-nats.oysterpack.com.crt           # server
          |-oysterpack.crt                    # ca
        |-keys                              # private keys
          |-client1.oysterpack.com.key        # client
          |-nats.oysterpack.com.key           # server
          |-oysterpack.key                    # ca
          
# How to decode a certificate

Via the command line:

    openssl x509 -in certificate.crt -text -noout

There are also online tools : [Certificate Decoder](https://www.sslshopper.com/certificate-decoder.html)
