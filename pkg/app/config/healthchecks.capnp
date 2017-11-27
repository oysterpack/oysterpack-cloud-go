using Go = import "/go.capnp";
using Metrics = import "metrics.capnp";
@0xe42204a141ec4e6f;
$Go.package("config");
$Go.import("github.com/oysterpack/oysterpack.go/pkg/app/config");

struct HealthCheckServiceSpec @0xf61f7c33b286f52f {
    healthCheckSpecs        @0 :List(HealthCheckSpec);
}

struct HealthCheckSpec @0xe65df7cace0dd1c2 {
    healthCheckID           @0 :UInt64;
    metricSpec              @1 :Metrics.GaugeMetricSpec;
    runTimeIntervalSeconds  @2 :UInt16;
}