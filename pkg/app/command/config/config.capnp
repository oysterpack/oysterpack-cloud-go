using Go = import "/go.capnp";
@0xac5630c48ddf1619;
$Go.package("config");
$Go.import("github.com/oysterpack/oysterpack.go/pkg/app/command/config");

struct Pipeline @0xfb501e5c22fbcd92 {

    struct Stage @0xa3e64eb06ea97afb {
        commandID   @0 :UInt64;
        poolSize    @1 :UInt8;
    }

    serviceID   @0 :UInt64;
    stages      @1 :List(Stage);
}