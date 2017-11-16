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

    runtime            @8 () -> (runtime :Runtime);
}

interface Service @0xb25b411cec149334 {
    id          @0 () -> (serviceId :UInt64);

    logLevel    @1 () -> (level :LogLevel);

    alive       @2 () -> (alive :Bool);
    kill        @3 () -> ();
}

enum LogLevel @0xce802aa8977a9aee {
    debug   @0;
    info    @1;
    warn    @2;
    error   @3;
}

interface Runtime @0xdda2e02140fe8f08 {
    goVersion       @0 () -> (version :Text);
    numCPU          @1 () -> (count :UInt32);
    numGoroutine    @2 () -> (count :UInt32);
    memStats        @3 () -> (stats :MemStats);
}

struct MemStats @0xcdf011e3e3860026 {
    alloc           @0 :UInt64 $Go.doc("Alloc is bytes of allocated heap objects");
    totalAlloc      @1 :UInt64 $Go.doc("TotalAlloc is cumulative bytes allocated for heap objects.");
    sys             @2 :UInt64 $Go.doc("Sys is the total bytes of memory obtained from the OS.");
    lookups         @3 :UInt64 $Go.doc("Lookups is the number of pointer lookups performed by the runtime");
    mallocs         @4 :UInt64 $Go.doc("Mallocs is the cumulative count of heap objects allocated.");
    frees           @5 :UInt64 $Go.doc("Frees is the cumulative count of heap objects freed.");

    heapAlloc       @6 :UInt64 $Go.doc("HeapAlloc is bytes of allocated heap objects");
    heapSys         @7 :UInt64 $Go.doc("HeapSys is bytes of heap memory obtained from the OS.");
    heapIdle        @8 :UInt64 $Go.doc("HeapIdle is bytes in idle (unused) spans.");
    heapInUse       @9 :UInt64 $Go.doc("HeapInuse is bytes in in-use spans.");
    heapReleased    @10 :UInt64 $Go.doc("HeapReleased is bytes of physical memory returned to the OS.");
    heapObjects     @11 :UInt64 $Go.doc("HeapObjects is the number of allocated heap objects.");

    stackInUse      @12 :UInt64;
    stackSys        @13 :UInt64;

    mSpanInUse      @14 :UInt64;
    mSpanSys        @15 :UInt64;
    mCacheInUse     @16 :UInt64;
    mCacheSys       @17 :UInt64;

    buckHashSys     @18 :UInt64;
    gCSys           @19 :UInt64;
    otherSys        @20 :UInt64;

    nextGC          @21 :UInt64;
    lastGC          @22 :UInt64;

    pauseTotalNs    @23 :UInt64;
    pauseNs         @24 :List(UInt64);
    pauseEnd        @25 :List(UInt64);

    numGC           @26 :UInt32;
    numForcedGC     @27 :UInt32;

    gCCPUFraction   @28 :Float64;

    struct BySize @0xc3e472677f9be8ad {
        size    @0 :UInt32;
        mallocs @1 :UInt64;
        frees   @2 :UInt64;
    }

    bySize          @29 :List(BySize);
}