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
    timestamp       @3 :Int64 $Go.doc("unix nano time");

    # data compression and packing
    # *** NOTE *** turn on compression / packing only after proving that it is needed and provides benefit
    compression     @4 :Compression = none;
    packed          @5 :Bool = false;

    data            @6 :Data;

    # Deadline returns the time when work done on behalf of this message should be canceled.
    deadline :union {
        # if not specified, then a zero timeout means no deadline
        timeoutMSec @7 :UInt16;
        expiresOn   @8 :Int64 $Go.doc("unix nano time");
    }
}

struct Ping @0x9bce611bc724ff89 {}

# used to reply to to a Ping
struct Pong @0xf6486a286fedf2f6 {}

struct SupportedMessageTypes @0xf56d6f421703b1f7 {

    struct Request @0x80b38fdd614a8b73 {}

    struct Response @0x87799f37b0d1886a {
        types @0 :List(UInt64);
    }

}