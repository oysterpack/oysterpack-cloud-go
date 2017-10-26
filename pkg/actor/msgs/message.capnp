using Go = import "/go.capnp";
@0x87fc44aa3255d2ae;
$Go.package("msgs");
$Go.import("github.com/oysterpack/oysterpack.go/pkg/actor/msgs");

struct Header @0xf38cccd618967ecd {
    id          @0 :Text $Go.doc("unique message id");
    created     @1 :Int64 $Go.doc("message creation timestamp - as a Unix time, the number of seconds elapsed since January 1, 1970 UTC");
    reply       @2 :ChannelAddress;
}

struct Address @0x9fd358f04cb684bd {
    system  @0 :Text $Go.doc("actor system");
    path    @1 :List(Text) $Go.doc("actor path");
    id      @2 :Text $Go.doc("actor id");
}

struct ChannelAddress @0xd801266d9df371b7 {
    channel @0 :Text $Go.doc("message channel name");
    address @1 :Address $Go.doc("actor address");
}
