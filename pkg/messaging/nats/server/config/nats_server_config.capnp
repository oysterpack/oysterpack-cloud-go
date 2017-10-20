using Go = import "/go.capnp";
@0xf8151d76c79d7b95;
$Go.package("config");
$Go.import("github.com/oysterpack/oysterpack.go/pkg/messaging/nats/server/config");

struct NATSServerConfig @0xe6c2ada3d6363f58 {
    struct HostPort @0xd618c87c1da62651 {
        host @0 :Text;
        port @1 :Int32;
    }

    enum NATSLogLevel @0xcd2ea85dab4af771 {
        nolog @0;
    	debug @1;
    	trace @2;
    }

    struct X509KeyPair {
        certPEMBlock    @0 :Data;
        keyPEMBlock     @1 :Data;
    }

    clusterName         @0 :Text $Go.doc("NATS cluster name");

    server              @1 :HostPort = (host = "0.0.0.0", port = 4222) $Go.doc("The address that clients connect to");
    monitor             @2 :HostPort = (host = "0.0.0.0", port = 8222) $Go.doc("The address used to export NATS monitoring HTTP APIs");

    cluster             @3 :HostPort = (host = "0.0.0.0", port = 5222) $Go.doc("The address used to commincate with the NATS cluster nodes");
    routes              @4 :List(Text) $Go.doc("This should be set to point to seed nodes. If not specified, then it adds itself as a route - assuming that this is a cluster seed node") ;

 	logLevel            @5 :NATSLogLevel = nolog;

 	maxPayload          @6 :Int32 = 100000;
 	maxConn             @7 :Int32 = 1024;

 	metricsExporterPort @8 :Int32 = 4444 $Go.doc("The address used to export prometheus metrics via HTTP");

 	serverX509KeyPair   @9 :X509KeyPair $Go.doc("used for client-server TLS");
 	clusterX509KeyPair  @10 :X509KeyPair $Go.doc("used for server-server cluster TLS");
}