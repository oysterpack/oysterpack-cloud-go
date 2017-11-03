# Ho to run tests in the package hierarchy

    go test ./... -p 1 -race
    
    where     
        -p n
                the number of programs, such as build commands or test binaries, that can be run in parallel.
                The default is the number of CPUs available. 
                
                The reason we set it 1 is avoid port conflicts when testing multiple packages in parallel.
                
        -race
                enable data race detection.
                Supported only on linux/amd64, freebsd/amd64, darwin/amd64 and windows/amd64.
    

# How to run test code coverage

    go test -covermode=count -coverprofile=cover.out     
    go tool cover -html cover.out

# How to run benchmarks

    go test -bench=. -benchmem
    
the -benchmem flag will include memory allocation statistics in the report

# How to profile
- it's usually better to profile specific benchmarks that have been constructed to be representative of workloads one cares about
    - benchmarking test cases is almost never representative, which is why they are disabled by using the filter **-run=NONE**
    
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
        
# GO IDE Setup
1. Plugins
    - Docker
    - Markdown
    - YAML
2. [Configuring Inotify Watches Limit](https://confluence.jetbrains.com/display/IDEADEV/Inotify+Watches+Limit)

Execute the following commands from this directory:

    sudo cp 60-idea.conf /etc/sysctl.d/
    sudo service procps start

