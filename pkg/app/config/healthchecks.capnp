using Go = import "/go.capnp";
using Metrics = import "metrics.capnp";
@0xe42204a141ec4e6f;
$Go.package("config");
$Go.import("github.com/oysterpack/oysterpack.go/pkg/app/config");


struct HealthcheckSpec @0xe65df7cace0dd1c2 {
    healthCheckID           @0 :UInt64;
    metricSpec              @1 :Metrics.GaugeMetricSpec;
    runTimeIntervalSeconds  @2 :UInt16;
}

