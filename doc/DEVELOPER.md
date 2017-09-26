# How to run test code coverage

    go test -covermode=count -coverprofile=cover.out     
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
    
### How to run Go Report Card
[OysterPack Go Report Card](https://goreportcard.com/report/github.com/oysterpack/oysterpack.go)

# How to format the code

    goimport -w -l ./
    
# Best Practices
1. Define all package errors in a centralized location in a file name errors.go
2. All error types should be implemented as references, e.g.

        // IllegalStateError indicates we are in an illegal state
        type IllegalStateError struct {
            State
            Message string
        }
        
        func (e *IllegalStateError) Error() string {
            if e.Message == "" {
                return e.State.String()
            }
            return fmt.Sprintf("%v : %v", e.State, e.Message)
        }
        
# How to run prometheus locally
1. Install prometheus locally using go (see [https://github.com/prometheus/prometheus](https://github.com/prometheus/prometheus))

        GO15VENDOREXPERIMENT=1 go get github.com/prometheus/prometheus/cmd/prometheus
        
2. prometheus -config.file <prometheus.yml>

   **NOTE**: run prometheus outside of the workspace because by default prometheus will create its data directory in the current working directory.

