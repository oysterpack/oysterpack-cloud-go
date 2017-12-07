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

	"github.com/oysterpack/oysterpack.go/pkg/app/uid"
	"gopkg.in/tomb.v2"
)

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
type Pipeline struct {
	tomb.Tomb

	id         PipelineID
	instanceID uid.UIDHash
	startedOn  time.Time

	in, out chan context.Context

	stages []*Stage
}

func (a *Pipeline) ID() PipelineID {
	return a.id
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
