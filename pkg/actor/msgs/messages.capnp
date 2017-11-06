using Go = import "/go.capnp";
@0x87fc44aa3255d2ae;
$Go.package("msgs");
$Go.import("github.com/oysterpack/oysterpack.go/pkg/actor/msgs");

struct Envelope @0xf38cccd618967ecd {
    struct ReplyTo @0xab748f26e2efa4c7 {
        address     @0 :Address;
        messageType @1 :UInt32 $Go.doc("the expected reply message type");
    }

    id              @0 :Text $Go.doc("unique message id");
    created         @1 :Int64 $Go.doc("message envelope creation timestamp - as a Unix time, the number of seconds elapsed since January 1, 1970 UTC");

    address         @2 :Address $Go.doc("who the envelope is addressed to, i.e., the actor address");
    messageType     @3 :UInt32 $Go.doc("the expected reply message type");

    replyTo         @4 :ReplyTo $Go.doc("indicates a reply is expected at the specified actor address with the specified message type");

    correlationId   @5 :Text $Go.doc("used to correlate this message to another message - use case is pairing request and response messages");

    message         @6 :Data $Go.doc("serialized message");
}

struct Address @0x9fd358f04cb684bd {
    path    @0 :List(Text) $Go.doc("actor path");
    id      @1 :Text $Go.doc("actor id");
}


