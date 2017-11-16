using Go = import "/go.capnp";
@0xdb8274f9144abc7e;
$Go.package("capnprpc");
$Go.import("github.com/oysterpack/oysterpack.go/pkg/app/capnprpc");

interface App @0xf052e7e084b31199 {
	id                 @0 () -> (appId :UInt64);
	releaseId          @1 () -> (releaseId :UInt64);
	instance           @2 () -> (instanceId :Text);

	# app start timestamp - as a Unix time, the number of seconds elapsed since January 1, 1970 UTC
	startedOn          @3 () -> (startedOn :Int64);

    logLevel           @4 () -> (level :LogLevel);

    registeredServices @5 () -> (serviceIds :List(UInt64));
    service            @6 (id :UInt64) -> (service :Service);

    kill               @7 () -> ();
}

enum LogLevel @0xce802aa8977a9aee {
    debug   @0;
    info    @1;
    warn    @2;
    error   @3;
}

interface Service @0xb25b411cec149334 {
    id          @0 () -> (serviceId :UInt64);

    logLevel    @1 () -> (level :LogLevel);

    alive       @2 () -> (alive :Bool);
    kill        @3 () -> ();
}