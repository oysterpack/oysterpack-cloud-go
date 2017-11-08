using Go = import "/go.capnp";
@0x87fc44aa3255d2ae;
$Go.package("msgs");
$Go.import("github.com/oysterpack/oysterpack.go/pkg/actor/msgs");

struct Envelope @0xf38cccd618967ecd {
    struct ReplyTo @0xab748f26e2efa4c7 {
        address     @0 :Address;
        messageType @1 :UInt64 $Go.doc("the expected reply message type");
    }

    id              @0 :Text $Go.doc("unique message id");
    created         @1 :Int64 $Go.doc("message envelope creation timestamp - as a Unix time, the number of seconds elapsed since January 1, 1970 UTC");

    address         @2 :Address $Go.doc("who the envelope is addressed to, i.e., the actor address");

    replyTo         @3 :ReplyTo $Go.doc("indicates a reply is expected at the specified actor address with the specified message type");
    correlationId   @4 :Text $Go.doc("used to correlate this message to another message - use case is pairing request and response messages");

    messageType     @5 :UInt64 $Go.doc("message type for the embedded binary message");
    message         @6 :Data $Go.doc("serialized message");
}

struct Address @0x9fd358f04cb684bd {
    path    @0 :Text $Go.doc("actor path");
    id      @1 :Text $Go.doc("actor id");
}

struct MessageProcessingError @0xa70dd5f5d238faaa {
    actorAddress    @0 :Address $Go.doc("actor path - where the error occurred");

    messageId       @1 :Text $Go.doc("failed message id");
    messageType     @2 :UInt64 = 0 $Go.doc("failed message type");

    errCode         @3 :Int64;
    errMessage      @4 :Text;
}

# Actor lifecycle messages
struct Started @0xf87ced6eabceba45 {}

struct Stopping @0xd064305fa8416cb7 {}

struct Stopped @0x874f20540dfe792f {}

struct PingRequest @0xa8084ffe960780bb {}

struct PingResponse @0xc8ea6e5ec7f91c46 {}




