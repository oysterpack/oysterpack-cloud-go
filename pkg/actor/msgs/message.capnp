using Go = import "/go.capnp";
@0x87fc44aa3255d2ae;
$Go.package("msgs");
$Go.import("github.com/oysterpack/oysterpack.go/pkg/actor/msgs");

struct Envelope @0xf38cccd618967ecd {
    id          @0 :Text $Go.doc("unique message id");
    created     @1 :Int64 $Go.doc("message envelope creation timestamp - as a Unix time, the number of seconds elapsed since January 1, 1970 UTC");

    replyTo     @2 :ChannelAddress $Go.doc("optional reply address");

    channel     @3 :Text $Go.doc("the channel the message is associated with");
    messageType @4 :UInt8 = 0 $Go.doc("message type is used when multiple types of messages can be sent on a channel");

    message     @5 :Data $Go.doc("serialized message");
}

struct Address @0x9fd358f04cb684bd {
    path    @0 :List(Text) $Go.doc("actor path");
    id      @1 :Text $Go.doc("actor id");
}

struct ChannelAddress @0xd801266d9df371b7 {
    channel @0 :Text $Go.doc("message channel name");
    address @1 :Address $Go.doc("actor address");
}

struct MessageProcessingError @0xa70dd5f5d238faaa {
    struct Message @0xd187ae75f5896d22 {
        id          @0 :Text $Go.doc("message id");
        created     @1 :Int64 $Go.doc("message envelope creation timestamp - as a Unix time, the number of seconds elapsed since January 1, 1970 UTC");

        replyTo     @2 :ChannelAddress $Go.doc("optional reply address");

        channel     @3 :Text $Go.doc("the channel the message is associated with");
        messageType @4 :UInt8 = 0 $Go.doc("message type is used when multiple types of messages can be sent on a channel");
    }

    path    @0 :List(Text) $Go.doc("actor path - where the error occurred");
    message @1 :Message $Go.doc("Error occurred while processing this message.");
    err     @2 :Text $Go.doc("Error message");
}
