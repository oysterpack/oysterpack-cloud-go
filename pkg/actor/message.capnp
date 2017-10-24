using Go = import "/go.capnp";
@0x87fc44aa3255d2ae;
$Go.package("actor");
$Go.import("github.com/oysterpack/oysterpack.go/pkg/actor");

struct Message @0xf38cccd618967ecd {

    struct Sender @0xf113452111fb2455 $Go.doc("refers to the actor that sent the message"){
        system  @0 :Text $Go.doc("actor system");
        id      @1 :Text $Go.doc("actor id");
        path    @2 :List(Text) $Go.doc("actor path");
    }

    enum Compression @0x8865b0351c00948c {
        none    @0;
        zlib    @1;
        gzip    @2;
        lz4     @3;
    }

    id          @0 :Text $Go.doc("unique message id");

    created     @1 :Int64 $Go.doc("message creation timestamp - as a Unix time, the number of seconds elapsed since January 1, 1970 UTC");

    sender      @2 :Sender;

    type        @3 :Text $Go.doc("message data type - how to parse the data");

    compression @4 :Compression = zlib $Go.doc("message data compression - default is zlib");

    data        @5 :Data $Go.doc("message data - usually a capnp message");
}