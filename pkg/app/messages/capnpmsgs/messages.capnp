using Go = import "/go.capnp";
@0x87fc44aa3255d2ae;
$Go.package("capnpmsgs");
$Go.import("github.com/oysterpack/oysterpack.go/pkg/app/messages/capnpmsgs");

struct Envelope @0xf38cccd618967ecd {
    struct ReplyTo @0xab748f26e2efa4c7 {
        to          @0 :Text $Go.doc("reply address");
        messageType @1 :UInt64 $Go.doc("the expected reply message type");
    }

    id              @0 :Text $Go.doc("unique message id");
    created         @1 :Int64 $Go.doc("message envelope creation timestamp - as a Unix time, the number of seconds elapsed since January 1, 1970 UTC");

    to              @2 :Text $Go.doc("who the envelope is addressed to");

    replyTo         @3 :ReplyTo $Go.doc("indicates a reply is expected at the specified actor address with the specified message type");
    correlationId   @4 :Text $Go.doc("used to correlate this message to another message - use case is pairing request and response messages");

    messageType     @5 :UInt64 $Go.doc("message type for the embedded binary message");
    message         @6 :Data $Go.doc("serialized message");
}

struct MessageProcessingError @0xa70dd5f5d238faaa {
    appId           @0 :UInt64 $Go.doc("app id");
    appInstanceId   @1 :Text;
    functionId      @2 :UInt64 $Go.doc("remote function that failed");

    messageId       @3 :Text $Go.doc("failed message id");
    messageType     @4 :UInt64 = 0 $Go.doc("failed message type");

    errCode         @5 :UInt64;
    errMessage      @6 :Text;
}

# system messages
struct PingRequest @0xa8084ffe960780bb {}

struct PingResponse @0xc8ea6e5ec7f91c46 {
    from            @0 :Text $Go.doc("from address");
}




