using Go = import "/go.capnp";
@0xdb8274f9144abc7e;
$Go.package("config");
$Go.import("github.com/oysterpack/oysterpack.go/pkg/app/config");

struct MetricsServiceSpec @0xb9780f65d5146efb {
    httpPort    @0 :UInt16;
    metricSpecs @1 :MetricSpecs;
}

struct MetricSpecs @0x88542dcd70c6048b {
    counterSpecs          @0 :List(CounterMetricSpec);
    counterVectorSpecs    @1 :List(CounterVectorMetricSpec);

    gaugeSpecs            @2 :List(GaugeMetricSpec);
    gaugeVectorSpecs      @3 :List(GaugeVectorMetricSpec);

    histogramSpecs        @4 :List(HistogramMetricSpec);
    histogramVectorSpecs  @5 :List(HistogramVectorMetricSpec);
}

struct CounterMetricSpec @0xfe237f35c45ecc97 {
    serviceId   @0 :UInt64;
    metricId    @1 :UInt64;
    help        @2 :Text $Go.doc("Help provides information about this metric. Mandatory!");
}

struct CounterVectorMetricSpec @0xdb34d9fcc1dffa24 {
    metricSpec  @0 :CounterMetricSpec;
    labelNames  @1 :List(Text);
}

struct GaugeMetricSpec @0xeebf043f542943d3 {
    serviceId   @0 :UInt64;
    metricId    @1 :UInt64;
    help        @2 :Text $Go.doc("Help provides information about this metric. Mandatory!");
}

struct GaugeVectorMetricSpec @0xa2274ad761e6a999 {
    metricSpec  @0 :GaugeMetricSpec;
    labelNames  @1 :List(Text);
}

struct HistogramMetricSpec @0x8e79552fdf96a8a7 {
    serviceId   @0 :UInt64;
    metricId    @1 :UInt64;
    help        @2 :Text $Go.doc("Help provides information about this metric. Mandatory!");
    buckets     @3 :List(Float64);
}

struct HistogramVectorMetricSpec @0x8527f1eb82ceeb98 {
    metricSpec  @0 :HistogramMetricSpec;
    labelNames  @1 :List(Text);
}