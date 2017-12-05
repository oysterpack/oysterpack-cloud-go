using Go = import "/go.capnp";
@0xaa44738dedfed9a1;
$Go.package("message");
$Go.import("github.com/oysterpack/oysterpack.go/pkg/app/message");

struct Message @0xc768aaf640842a35 {

    enum Compression @0xf8f433c185247295 {
        none    @0;
        zlib    @1;
    }

    id              @0 :UInt64;
    type            @1 :UInt64;
    correlationID   @2 :UInt64;
    timestamp       @3 :Int64;
    compression     @4 :Compression = zlib;
    data            @5 :Data;
}

struct Ping @0x9bce611bc724ff89 {}

# used to reply to to a Ping
struct Pong @0xf6486a286fedf2f6 {}