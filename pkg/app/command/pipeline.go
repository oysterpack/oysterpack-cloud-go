// Copyright (c) 2017 OysterPack, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package command

import (
	"context"
	"time"

	"fmt"

	"github.com/oysterpack/oysterpack.go/pkg/app"
	"github.com/oysterpack/oysterpack.go/pkg/app/uid"
	"github.com/prometheus/client_golang/prometheus"
)

func StartPipeline(service *app.Service, stages ...Stage) (*Pipeline, error) {
	if service == nil {
		return nil, app.IllegalArgumentError("A pipeline requires a service to run")
	}
	if !service.Alive() {
		return nil, app.ServiceNotAliveError(service.ID())
	}
	if len(stages) == 0 {
		return nil, app.IllegalArgumentError("A pipeline must have at least 1 stage")
	}
	for _, stage := range stages {
		if stage.Command().run == nil {
			return nil, app.IllegalArgumentError(fmt.Sprintf("Stage Command run function was nil for : ServiceID(0x%x)", service.ID()))
		}
	}

	pipeline := &Pipeline{
		Service:    service,
		instanceID: uid.NextUIDHash(),
		startedOn:  time.Now(),
		in:         make(chan context.Context),
		out:        make(chan context.Context),
		stages:     stages,
	}

	var build func(stages []Stage, in, out chan context.Context)
	build = func(stages []Stage, in, out chan context.Context) {
		if len(stages) == 1 {
			stage := stages[0]
			poolSize := int(stage.PoolSize())
			for i := 0; i < poolSize; i++ {
				service.Go(func() error {
					for {
						select {
						case <-service.Dying():
							return nil
						case ctx := <-in:
							select {
							case <-ctx.Done():
								PipelineContextExpired(ctx, pipeline, stage.Command().CommandID()).Log(pipeline.Service.Logger())
							default:
								result, err := stage.Command().Run(ctx)
								if err != nil {
									result = SetError(result, stage.Command().id, err)
								}
								select {
								case <-service.Dying():
									return nil
								case <-result.Done():
									PipelineContextExpired(ctx, pipeline, stage.Command().CommandID()).Log(pipeline.Service.Logger())
								case pipeline.out <- result:
								}
							}
						}
					}
				})
			}
			return
		}

		stage := stages[0]
		poolSize := int(stage.PoolSize())
		for i := 0; i < poolSize; i++ {
			service.Go(func() error {
				for {
					select {
					case <-service.Dying():
						return nil
					case ctx := <-in:
						select {
						case <-ctx.Done():
							PipelineContextExpired(ctx, pipeline, stage.Command().CommandID()).Log(pipeline.Service.Logger())
						default:
							result, err := stage.Command().Run(ctx)
							if err != nil {
								result = SetError(result, stage.Command().id, err)
								select {
								case <-service.Dying():
									return nil
								case <-result.Done():
									PipelineContextExpired(ctx, pipeline, stage.Command().CommandID()).Log(pipeline.Service.Logger())
								case pipeline.out <- result:
								}
							} else {
								select {
								case <-service.Dying():
									return nil
								case <-result.Done():
									PipelineContextExpired(ctx, pipeline, stage.Command().CommandID()).Log(pipeline.Service.Logger())
								case out <- result:
								}
							}
						}
					}
				}
			})
		}
		build(stages[1:], out, make(chan context.Context))
	}

	build(stages, pipeline.in, make(chan context.Context))

	return pipeline, nil
}

// Pipeline is a series of stages connected by channels, where each stage is a group of goroutines running the same function.
//
// In each stage, the goroutines
//
// 	- receive values from upstream via inbound channels
//	- perform some function on that data, usually producing new values
//	- send values downstream via outbound channels
//
// Each stage has any number of inbound and outbound channels, except the first and last stages, which have only outbound
// or inbound channels, respectively. The first stage is sometimes called the source or producer; the last stage, the sink or consumer.
//
// For more background information on pipelines see https://blog.golang.org/pipelines
//
// How are context expirations handled on the pipeline ?
//	- The context is dropped, i.e., it no longer continues on the pipeline. The expiration is logged.
//  - The context is checked if it is expired when it is received on each stage and after the command runs.
//
// What happens if an error is returned by a pipeline stage command ?
// 	- The error is added to the Context using CTX_KEY_CMD_ERR as the key. The workflow is aborted, and the context is
// 	  returned immediately on the pipeline output channel.
type Pipeline struct {
	*app.Service

	instanceID uid.UIDHash
	startedOn  time.Time

	in, out chan context.Context

	stages []Stage

	messageCounter        prometheus.CounterVec
	errCounter            prometheus.CounterVec
	contextExpiredCounter prometheus.CounterVec
	processingTime        prometheus.CounterVec
}

func (a *Pipeline) InstanceID() uid.UIDHash {
	return a.instanceID
}

func (a *Pipeline) StartedOn() time.Time {
	return a.startedOn
}

func (a *Pipeline) InputChan() chan<- context.Context {
	return a.in
}

func (a *Pipeline) OutputChan() <-chan context.Context {
	return a.out
}

func (a *Pipeline) Stages() []Stage {
	stages := make([]Stage, len(a.stages))
	for i := 0; i < len(stages); i++ {
		stages[i] = a.stages[i]
	}
	return stages
}

func NewStage(cmd Command, poolSize uint8) Stage {
	return Stage{cmd, poolSize}
}

// Stage represents a pipeline stage
type Stage struct {
	cmd      Command
	poolSize uint8
}

// Command returns the stage's command
func (a *Stage) Command() Command {
	return a.cmd
}

// PoolSize returns the number of concurrent command instances to run in this stage
func (a *Stage) PoolSize() uint8 {
	if a.poolSize == 0 {
		return 1
	}
	return a.poolSize
}
