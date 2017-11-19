# TLS

### Use the following command to decrypt an encrypted RSA key:

    openssl rsa -in ssl.key.secure -out ssl.key
    
### How to decode a certificate

    openssl x509 -in certificate.crt -text -noout

There are also online tools : [Certificate Decoder](https://www.sslshopper.com/certificate-decoder.html)
