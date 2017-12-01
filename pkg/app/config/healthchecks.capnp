using Go = import "/go.capnp";
@0xe42204a141ec4e6f;
$Go.package("config");
$Go.import("github.com/oysterpack/oysterpack.go/pkg/app/config");

struct HealthCheckServiceSpec @0xf61f7c33b286f52f {
    healthCheckSpecs        @0 :List(HealthCheckSpec);
    commandServerChanSize   @1 :UInt8;
}

struct HealthCheckSpec @0xe65df7cace0dd1c2 {
    healthCheckID           @0 :UInt64;
    runIntervalSeconds      @1 :UInt16 $Go.doc("How often to run the healthcheck");
    timeoutSeconds          @2 :UInt8 $Go.doc("Max time to alot the healthcheck to run.");
}