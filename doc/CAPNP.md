# [CAPNP](https://capnproto.org)

## Installation : UNIX

    curl -O https://capnproto.org/capnproto-c++-0.6.1.tar.gz
    tar zxf capnproto-c++-0.6.1.tar.gz
    cd capnproto-c++-0.6.1
    ./configure
    make -j6 check
    sudo make install
    
This will install capnp, the Cap’n Proto command-line tool. It will also install libcapnp, libcapnpc, and libkj in 
/usr/local/lib and headers in /usr/local/include/capnp and /usr/local/include/kj.

## [go-capnproto2](https://github.com/capnproto/go-capnproto2)

installation

    $ go get -u -t zombiezen.com/go/capnproto2/...
    $ go test -v zombiezen.com/go/capnproto2/...
    
## Compile the capnp schemas and generate go code

    capnp compile -I$GOPATH/src/zombiezen.com/go/capnproto2/std -ogo *.capnp 
    
## BUGS
- [server/server.go - panic close of closed channel](https://github.com/capnproto/go-capnproto2/issues/101)
