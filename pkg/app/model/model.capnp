using Go = import "/go.capnp";
@0xdb8274f9144abc7e;
$Go.package("model");
$Go.import("github.com/oysterpack/oysterpack.go/pkg/app/model");

# The goal is to model the oysterpack application platform.
# This is a work in progress ...

struct Domain @0x919720e9ef6739b7 {
    id              @0 :UInt64;
    name            @1 :Text;

    createdOn       @2 :UInt64;
    updatedOn       @3 :UInt64;
    deletedOn       @4 :UInt64;

    parentDomain    @5 :UInt64;
}

struct App @0xbb0c08731dac73a6 {
    domainId            @0 :UInt64;
    id                  @1 :UInt64;
    name                @2 :Text;

    createdOn           @3 :UInt64;
    updatedOn           @4 :UInt64;
    deletedOn           @5 :UInt64;
}

struct AppRelease @0xa2fab062a5262acf {
    appId               @0 :UInt64;
    releaseId           @1 :UInt64;

    releasedOn          @2 :UInt64;

    serviceIds          @3 :List(UInt64);
    domainDependencies  @4 :List(UInt64);
}

struct ServiceConfig @0xbecdd492fe0f4023 {
    serviceId       @0 :UInt64;

    config          @1 :UInt64;

    createdOn       @2 :UInt64;
    updatedOn       @3 :UInt64;
    deletedOn       @4 :UInt64;
}