using Go = import "/go.capnp";
@0xdb8274f9144abc7e;
$Go.package("capnprpc");
$Go.import("github.com/oysterpack/oysterpack.go/pkg/app/capnprpc");

interface App @0xf052e7e084b31199 {
	id          @0 () -> (appId :UInt64);
	instance    @1 () -> (instanceId :Text);

	# app start timestamp - as a Unix time, the number of seconds elapsed since January 1, 1970 UTC
	startedOn   @2 () -> (startedOn :Int64);
}

