# TLS

### Use the following command to decrypt an encrypted RSA key:

    openssl rsa -in ssl.key.secure -out ssl.key
    
### How to decode a certificate

    openssl x509 -in certificate.crt -text -noout

There are also online tools : [Certificate Decoder](https://www.sslshopper.com/certificate-decoder.html)


## Articles
[Mutual TLS authentication in Go](http://www.levigross.com/2015/11/21/mutual-tls-authentication-in-go/)
[So you want to expose Go on the Internet](https://blog.gopheracademy.com/advent-2016/exposing-go-on-the-internet/)

## [Let's Encrypt](https://letsencrypt.org/)
- [golang.org/x/crypto/acme/autocert](https://godoc.org/golang.org/x/crypto/acme/autocert)
- [Generate and Use Free TLS Certificates with Lego](https://blog.gopheracademy.com/advent-2015/generate-free-tls-certs-with-lego/)

## [TinyCert](https://www.tinycert.org)