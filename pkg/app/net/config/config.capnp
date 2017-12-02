using Go = import "/go.capnp";
@0xcee75c59b9f2a30b;
$Go.package("config");
$Go.import("github.com/oysterpack/oysterpack.go/pkg/app/net/config");

struct ServerSpec @0xe57b76fedcda1734 {
    serviceSpec     @0 :ServiceSpec;
    serverCert      @1 :X509KeyPair;
    caCert          @2 :Data;
    maxConns        @3 :UInt32;
}

struct ClientSpec @0x853a22bea61af6f5 {
    serviceSpec     @0 :ServiceSpec;
    clientCert      @1 :X509KeyPair;
    caCert          @2 :Data $Go.doc("PEM file format");
}

struct ServiceSpec @0x8e98877ce02ee396 {
    domainID        @0 :UInt64;
    appId           @1 :UInt64;
    serviceId       @2 :UInt64;

    port            @3 :UInt16;
}

struct X509KeyPair @0xf82cc68ebab66792 {
    key     @0 :Data $Go.doc("PEM file format");
    cert    @1 :Data $Go.doc("PEM file format");
}