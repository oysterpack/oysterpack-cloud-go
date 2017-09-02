# How to run test code coverage

    go test -mode=count -coverprofile=cover.out     
    go tool cover -html cover.out

# How to run benchmarks

    go test -bench=. -benchmem
    
the -benchmem flag will include memory allocation statistics in the report

# How to profile
- it's usually better to profile specific benchmarks that have been constructed to be representative of workloads one cares about
    - benchmarking test cases is almost never represtative, which is why they are disabled by using the filter **-run=NONE**
    
### collect profile data
    go test -run=NONE -bench=. -blockprofile block.out
    go test -run=NONE -bench=. -cpuprofile cpu.out
    go test -run=NONE -bench=. -memprofile mem.out
    go test -run=NONE -bench=. -mutexprofile mutex.out
    
### for all profile flags, see
    go help testflag 
    
### view profile reports
    go tool pprof