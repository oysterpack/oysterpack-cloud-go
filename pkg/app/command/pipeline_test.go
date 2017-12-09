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

package command_test

import (
	"context"
	"testing"

	"os"

	"github.com/oysterpack/oysterpack.go/pkg/app"
	"github.com/oysterpack/oysterpack.go/pkg/app/command"
	appconfig "github.com/oysterpack/oysterpack.go/pkg/app/config"
	"github.com/oysterpack/oysterpack.go/pkg/app/uid"
	"zombiezen.com/go/capnproto2"
)

func initConfigDir(configDir string) {
	app.Configs.SetConfigDir(configDir)
	os.RemoveAll(configDir)
	os.MkdirAll(configDir, 0755)
}

func initMetricsConfig(serviceID app.ServiceID) error {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return err
	}
	metricsServiceSpec, err := appconfig.NewRootMetricsServiceSpec(seg)
	if err != nil {
		return err
	}
	metricsSpecs, err := appconfig.NewMetricSpecs(seg)
	if err != nil {
		return err
	}
	metricsServiceSpec.SetMetricSpecs(metricsSpecs)

	createCounters := func() error {
		counters, err := metricsSpecs.NewCounterSpecs(6)
		if err != nil {
			return err
		}

		////////////
		pipelineRunCount, err := appconfig.NewCounterMetricSpec(seg)
		if err != nil {
			return err
		}
		pipelineRunCount.SetServiceId(serviceID.UInt64())
		pipelineRunCount.SetMetricId(command.PIPELINE_RUN_COUNT.UInt64())
		if err := pipelineRunCount.SetHelp("Total number of times the pipeline was run"); err != nil {
			return err
		}
		counters.Set(0, pipelineRunCount)

		////////////
		pipelineFailureCount, err := appconfig.NewCounterMetricSpec(seg)
		if err != nil {
			return err
		}
		pipelineFailureCount.SetServiceId(serviceID.UInt64())
		pipelineFailureCount.SetMetricId(command.PIPELINE_FAILED_COUNT.UInt64())
		if err := pipelineFailureCount.SetHelp("Total number of failed pipeline workflows"); err != nil {
			return err
		}
		counters.Set(1, pipelineFailureCount)

		////////////
		pipelineContextExpiredCount, err := appconfig.NewCounterMetricSpec(seg)
		if err != nil {
			return err
		}
		pipelineContextExpiredCount.SetServiceId(serviceID.UInt64())
		pipelineContextExpiredCount.SetMetricId(command.PIPELINE_CONTEXT_EXPIRED_COUNT.UInt64())
		if err := pipelineContextExpiredCount.SetHelp("Total number of Context expirations, i.e., workflows that were either cancelled or timeout."); err != nil {
			return err
		}
		counters.Set(2, pipelineContextExpiredCount)

		////////////
		pipelineProcessingTime, err := appconfig.NewCounterMetricSpec(seg)
		if err != nil {
			return err
		}
		pipelineProcessingTime.SetServiceId(serviceID.UInt64())
		pipelineProcessingTime.SetMetricId(command.PIPELINE_PROCESSING_TIME_SEC.UInt64())
		if err := pipelineProcessingTime.SetHelp("Total pipeline processing time in seconds"); err != nil {
			return err
		}
		counters.Set(3, pipelineProcessingTime)

		////////////
		pipelineFailedProcessingTime, err := appconfig.NewCounterMetricSpec(seg)
		if err != nil {
			return err
		}
		pipelineFailedProcessingTime.SetServiceId(serviceID.UInt64())
		pipelineFailedProcessingTime.SetMetricId(command.PIPELINE_PROCESSING_TIME_SEC_FAILED.UInt64())
		if err := pipelineFailedProcessingTime.SetHelp("Total pipeline processing time in seconds for failed workflows"); err != nil {
			return err
		}
		counters.Set(4, pipelineFailedProcessingTime)

		pipelineDeliveryTime, err := appconfig.NewCounterMetricSpec(seg)
		if err != nil {
			return err
		}
		pipelineDeliveryTime.SetServiceId(serviceID.UInt64())
		pipelineDeliveryTime.SetMetricId(command.PIPELINE_CHANNEL_DELIVERY_TIME_SEC.UInt64())
		if err := pipelineDeliveryTime.SetHelp("Total pipeline channel delivery time in seconds"); err != nil {
			return err
		}
		counters.Set(5, pipelineDeliveryTime)

		return nil
	}

	createCounterVectors := func() error {
		counters, err := metricsSpecs.NewCounterVectorSpecs(4)
		if err != nil {
			return err
		}

		//////////
		commandRunCountVector, err := appconfig.NewCounterVectorMetricSpec(seg)
		if err != nil {
			return err
		}
		commandRunCount, err := appconfig.NewCounterMetricSpec(seg)
		if err != nil {
			return err
		}
		commandRunCount.SetServiceId(serviceID.UInt64())
		commandRunCount.SetMetricId(command.COMMAND_RUN_COUNT.UInt64())
		if err := commandRunCount.SetHelp("Total number of times the pipeline was run"); err != nil {
			return err
		}
		if err := commandRunCountVector.SetMetricSpec(commandRunCount); err != nil {
			return err
		}
		labels, err := commandRunCountVector.NewLabelNames(1)
		if err != nil {
			return err
		}
		labels.Set(0, command.LABEL_COMMAND)
		counters.Set(0, commandRunCountVector)

		/////////
		commandFailureCountVector, err := appconfig.NewCounterVectorMetricSpec(seg)
		if err != nil {
			return err
		}
		commandFailureCount, err := appconfig.NewCounterMetricSpec(seg)
		if err != nil {
			return err
		}
		commandFailureCount.SetServiceId(serviceID.UInt64())
		commandFailureCount.SetMetricId(command.COMMAND_FAILED_COUNT.UInt64())
		if err := commandFailureCount.SetHelp("Total number of failed pipeline workflows"); err != nil {
			return err
		}
		if err := commandFailureCountVector.SetMetricSpec(commandFailureCount); err != nil {
			return err
		}
		labels, err = commandFailureCountVector.NewLabelNames(1)
		if err != nil {
			return err
		}
		labels.Set(0, command.LABEL_COMMAND)
		counters.Set(1, commandFailureCountVector)

		//////////
		commandProcessingTimeVector, err := appconfig.NewCounterVectorMetricSpec(seg)
		commandProcessingTime, err := appconfig.NewCounterMetricSpec(seg)
		if err != nil {
			return err
		}
		commandProcessingTime.SetServiceId(serviceID.UInt64())
		commandProcessingTime.SetMetricId(command.COMMAND_PROCESSING_TIME_SEC.UInt64())
		if err := commandProcessingTime.SetHelp("Total pipeline processing time in nanoseconds"); err != nil {
			return err
		}
		commandProcessingTimeVector.SetMetricSpec(commandProcessingTime)
		labels, err = commandProcessingTimeVector.NewLabelNames(1)
		if err != nil {
			return err
		}
		labels.Set(0, command.LABEL_COMMAND)
		counters.Set(2, commandProcessingTimeVector)

		/////////////
		commandFailedProcessingTimeVector, err := appconfig.NewCounterVectorMetricSpec(seg)
		commandFailedProcessingTime, err := appconfig.NewCounterMetricSpec(seg)
		if err != nil {
			return err
		}
		commandFailedProcessingTime.SetServiceId(serviceID.UInt64())
		commandFailedProcessingTime.SetMetricId(command.COMMAND_PROCESSING_TIME_SEC_FAILED.UInt64())
		if err := commandFailedProcessingTime.SetHelp("Total pipeline processing time in nanoseconds for failed workflows"); err != nil {
			return err
		}
		commandFailedProcessingTimeVector.SetMetricSpec(commandFailedProcessingTime)
		commandProcessingTimeVector.SetMetricSpec(commandProcessingTime)
		labels, err = commandFailedProcessingTimeVector.NewLabelNames(1)
		if err != nil {
			return err
		}
		labels.Set(0, command.LABEL_COMMAND)
		counters.Set(3, commandFailedProcessingTimeVector)

		return nil
	}

	if err := createCounters(); err != nil {
		return err
	}

	if err := createCounterVectors(); err != nil {
		return err
	}

	// store the config
	serviceConfigPath := app.Configs.ServiceConfigPath(app.METRICS_SERVICE_ID)
	configFile, err := os.Create(serviceConfigPath)
	if err != nil {
		return err
	}
	app.MarshalCapnpMessage(msg, configFile)
	configFile.Close()

	return nil
}

func TestStartPipeline(t *testing.T) {
	SERVICE_ID := app.ServiceID(uid.NextUIDHash())

	configDir := "./testdata/pipeline_test/TestStartPipeline"
	initConfigDir(configDir)
	initMetricsConfig(SERVICE_ID)

	t.Run("single stage - no error", func(t *testing.T) {
		app.ResetWithConfigDir(configDir)
		defer app.Reset()

		service := app.NewService(SERVICE_ID)

		type Key int

		const (
			A = Key(iota)
			B
			SUM
		)

		p, err := command.StartPipeline(service,
			command.NewStage(
				SERVICE_ID,
				command.NewCommand(command.CommandID(1), func(ctx context.Context) (context.Context, *app.Error) {
					a := ctx.Value(A).(int)
					b := ctx.Value(B).(int)
					return context.WithValue(ctx, SUM, a+b), nil
				}),
				1,
			),
		)
		if err != nil {
			t.Fatal(err)
		}

		ctx := context.WithValue(command.NewContext(), A, 1)
		ctx = context.WithValue(ctx, B, 2)
		p.InputChan() <- ctx
		ctx = <-p.OutputChan()
		sum := ctx.Value(SUM).(int)
		t.Logf("sum = %d", sum)
		if sum != 3 {
			t.Errorf("The pipeline did not process the workflow correctly : sum = %d", sum)
		}
	})

	t.Run("10 stage pipeline", func(t *testing.T) {
		app.ResetWithConfigDir(configDir)
		defer app.Reset()

		service := app.NewService(SERVICE_ID)
		type Key int

		const (
			N = Key(iota)
		)
		stage := command.NewStage(
			SERVICE_ID,
			command.NewCommand(command.CommandID(1), func(ctx context.Context) (context.Context, *app.Error) {
				n := ctx.Value(N).(int)
				return context.WithValue(ctx, N, n+1), nil
			}),
			1,
		)

		stages := []command.Stage{}
		for i := 0; i < 10; i++ {
			stages = append(stages, stage)
		}
		p, err := command.StartPipeline(service, stages...)
		if err != nil {
			t.Fatal(err)
		}
		ctx := context.WithValue(command.NewContext(), N, 0)
		p.InputChan() <- ctx
		ctx = <-p.OutputChan()
		n := ctx.Value(N).(int)
		t.Logf("n = %d", n)
		if n != 10 {
			t.Errorf("The pipeline did not process the workflow correctly : n = %d", n)
		}
	})

	t.Run("10 stage pipeline - with reply channel", func(t *testing.T) {
		app.ResetWithConfigDir(configDir)
		defer app.Reset()

		service := app.NewService(SERVICE_ID)
		type Key int

		const (
			N = Key(iota)
		)
		stage := command.NewStage(
			SERVICE_ID,
			command.NewCommand(command.CommandID(1), func(ctx context.Context) (context.Context, *app.Error) {
				n := ctx.Value(N).(int)
				return context.WithValue(ctx, N, n+1), nil
			}),
			1,
		)

		stages := []command.Stage{}
		for i := 0; i < 10; i++ {
			stages = append(stages, stage)
		}
		p, err := command.StartPipeline(service, stages...)
		if err != nil {
			t.Fatal(err)
		}
		ctx := context.WithValue(command.NewContext(), N, 0)
		p.InputChan() <- ctx
		ctx = <-p.OutputChan()
		n := ctx.Value(N).(int)
		t.Logf("n = %d", n)
		if n != 10 {
			t.Errorf("The pipeline did not process the workflow correctly : n = %d", n)
		}
		if workflowID, ok := command.WorkflowID(ctx); !ok {
			t.Error("output context should have a workflow id assigned")
		} else {
			t.Logf("workflowID = %x", workflowID)
		}

		timeStarted := command.PipelineWorkflowStartTime(ctx)
		if timeStarted.IsZero() {
			t.Error("the pipeline workflow start time is not in the context")
		} else {
			t.Logf("timeStarted : %v", timeStarted)
		}

	})
}
