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
	"github.com/prometheus/client_golang/prometheus"
)

func StartPipeline(service *app.Service, stages ...Stage) (*Pipeline, error) {
	checkArgs := func() *app.Error {
		if service == nil {
			return app.IllegalArgumentError("A pipeline requires a service to run")
		}
		if !service.Alive() {
			return app.ServiceNotAliveError(service.ID())
		}
		if len(stages) == 0 {
			return app.IllegalArgumentError("A pipeline must have at least 1 stage")
		}
		for _, stage := range stages {
			if stage.Command().run == nil {
				return app.IllegalArgumentError(fmt.Sprintf("Stage Command run function was nil for : ServiceID(0x%x)", service.ID()))
			}
		}
		return nil
	}

	if err := checkArgs(); err != nil {
		return nil, err
	}

	serviceID := service.ID()

	pipeline := &Pipeline{
		Service:   service,
		startedOn: time.Now(),
		in:        make(chan context.Context),
		out:       make(chan context.Context),
		stages:    stages,

		runCounter:            app.MetricRegistry.Counter(serviceID, PIPELINE_RUN_COUNT),
		failedCounter:         app.MetricRegistry.Counter(serviceID, PIPELINE_FAILED_COUNT),
		contextExpiredCounter: app.MetricRegistry.Counter(serviceID, PIPELINE_CONTEXT_EXPIRED_COUNT),
		processingTime:        app.MetricRegistry.Counter(serviceID, PIPELINE_PROCESSING_TIME_SEC),
		processingFailedTime:  app.MetricRegistry.Counter(serviceID, PIPELINE_PROCESSING_TIME_SEC_FAILED),
		channelDeliveryTime:   app.MetricRegistry.Counter(serviceID, PIPELINE_CHANNEL_DELIVERY_TIME_SEC),
	}

	firstStageCommandID := pipeline.stages[0].cmd.id
	var build func(stages []Stage, in, out chan context.Context)
	build = func(stages []Stage, in, out chan context.Context) {
		if len(stages) == 1 {
			stage := stages[0]

			process := func(ctx context.Context) {
				result, err := stage.run(ctx)
				processedTime := time.Now()
				workflowTime := time.Now().Sub(PipelineWorkflowStartTime(ctx)).Seconds()
				pipeline.processingTime.Add(workflowTime)
				if err != nil {
					pipeline.failedCounter.Inc()
					pipeline.processingFailedTime.Add(workflowTime)
					result = SetError(result, stage.Command().id, err)
				}
				select {
				case <-service.Dying():
					return
				case <-result.Done():
					pipelineContextExpired(ctx, pipeline, stage.Command().CommandID()).Log(pipeline.Service.Logger())
				case pipeline.out <- result:
					deliveryTime := time.Now().Sub(processedTime).Seconds()
					pipeline.channelDeliveryTime.Add(deliveryTime)
				}
			}

			poolSize := int(stage.PoolSize())
			for i := 0; i < poolSize; i++ {
				if stage.cmd.id == firstStageCommandID {
					service.Go(func() error {
						for {
							select {
							case <-service.Dying():
								return nil
							case ctx := <-in:
								select {
								case <-ctx.Done():
									pipelineContextExpired(ctx, pipeline, stage.Command().CommandID()).Log(pipeline.Service.Logger())
								default:
									// record the time when the context started the workflow, i.e., entered the first stage of the pipeline
									ctx = PipelineWorkflowStarted(ctx)
									pipeline.runCounter.Inc()
									process(ctx)
								}
							}
						}
					})
				} else {
					service.Go(func() error {
						for {
							select {
							case <-service.Dying():
								return nil
							case ctx := <-in:
								select {
								case <-ctx.Done():
									pipelineContextExpired(ctx, pipeline, stage.Command().CommandID()).Log(pipeline.Service.Logger())
								default:
									process(ctx)
								}
							}
						}
					})
				}

			}
			return
		}

		stage := stages[0]
		poolSize := int(stage.PoolSize())

		process := func(ctx context.Context) {
			result, err := stage.run(ctx)
			processedTime := time.Now()
			if err != nil {
				pipeline.failedCounter.Inc()
				result = SetError(result, stage.Command().id, err)
				select {
				case <-service.Dying():
					return
				case <-result.Done():
					pipelineContextExpired(ctx, pipeline, stage.Command().CommandID()).Log(pipeline.Service.Logger())
				case pipeline.out <- result:
					deliveryTime := time.Now().Sub(processedTime).Seconds()
					pipeline.channelDeliveryTime.Add(deliveryTime)
				}
			} else {
				select {
				case <-service.Dying():
					return
				case <-result.Done():
					pipelineContextExpired(ctx, pipeline, stage.Command().CommandID()).Log(pipeline.Service.Logger())
				case out <- result:
					deliveryTime := time.Now().Sub(processedTime).Seconds()
					pipeline.channelDeliveryTime.Add(deliveryTime)
				}
			}
		}

		for i := 0; i < poolSize; i++ {
			if stage.cmd.id == firstStageCommandID {
				service.Go(func() error {
					for {
						select {
						case <-service.Dying():
							return nil
						case ctx := <-in:
							select {
							case <-ctx.Done():
								pipelineContextExpired(ctx, pipeline, stage.Command().CommandID()).Log(pipeline.Service.Logger())
							default:
								// record the time when the context started the workflow, i.e., entered the first stage of the pipeline
								ctx = PipelineWorkflowStarted(ctx)
								pipeline.runCounter.Inc()
								process(ctx)
							}
						}
					}
				})
			} else {
				service.Go(func() error {
					for {
						select {
						case <-service.Dying():
							return nil
						case ctx := <-in:
							select {
							case <-ctx.Done():
								pipelineContextExpired(ctx, pipeline, stage.Command().CommandID()).Log(pipeline.Service.Logger())
							default:
								process(ctx)
							}
						}
					}
				})
			}

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

	startedOn time.Time

	in, out chan context.Context

	stages []Stage

	runCounter            prometheus.Counter
	failedCounter         prometheus.Counter
	contextExpiredCounter prometheus.Counter
	processingTime        prometheus.Counter
	processingFailedTime  prometheus.Counter
	channelDeliveryTime   prometheus.Counter
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

func NewStage(serviceID app.ServiceID, cmd Command, poolSize uint8) Stage {
	return Stage{cmd: cmd,
		poolSize:             poolSize,
		runCounter:           app.MetricRegistry.CounterVector(serviceID, COMMAND_RUN_COUNT).CounterVec.With(prometheus.Labels{LABEL_COMMAND: cmd.CommandID().Hex()}),
		failedCounter:        app.MetricRegistry.CounterVector(serviceID, COMMAND_FAILED_COUNT).CounterVec.With(prometheus.Labels{LABEL_COMMAND: cmd.CommandID().Hex()}),
		processingTime:       app.MetricRegistry.CounterVector(serviceID, COMMAND_PROCESSING_TIME_SEC).CounterVec.With(prometheus.Labels{LABEL_COMMAND: cmd.CommandID().Hex()}),
		processingFailedTime: app.MetricRegistry.CounterVector(serviceID, COMMAND_PROCESSING_TIME_SEC_FAILED).CounterVec.With(prometheus.Labels{LABEL_COMMAND: cmd.CommandID().Hex()}),
	}
}

// Stage represents a pipeline stage
type Stage struct {
	cmd      Command
	poolSize uint8

	runCounter           prometheus.Counter
	failedCounter        prometheus.Counter
	processingTime       prometheus.Counter
	processingFailedTime prometheus.Counter
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

func (a *Stage) run(in context.Context) (out context.Context, err *app.Error) {
	a.runCounter.Inc()
	start := time.Now()
	out, err = a.cmd.Run(in)
	runTime := time.Now().Sub(start).Seconds()
	a.processingTime.Add(runTime)
	if err != nil {
		a.failedCounter.Inc()
		a.processingFailedTime.Add(runTime)
	}
	return
}
