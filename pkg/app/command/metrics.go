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

import "github.com/oysterpack/oysterpack.go/pkg/app"

const (
	LABEL_COMMAND = "cmd"

	// number of times the command has been run
	COMMAND_RUN_COUNT = app.MetricID(0xdccbcbcf9235211d)
	// number of times the command has failed - includes context expirations
	COMMAND_FAILED_COUNT = app.MetricID(0xd8b89ef52341ad4b)
	// total accumulative command processing time
	COMMAND_PROCESSING_TIME_SEC = app.MetricID(0x9684d6de9cf15d73)
	// total accumulative command processing time for commands that failed
	COMMAND_PROCESSING_TIME_SEC_FAILED = app.MetricID(0x86f01c622f5894c5)

	// number of times the pipeline ran, which corresponds to the number of messages were sent into the pipeline, i.e.,
	// the number of messages that were received on the pipeline input channel
	PIPELINE_RUN_COUNT = app.MetricID(0xdf381a0443622af1)
	// number of times that a pipeline workflow failed - including context expirations
	PIPELINE_FAILED_COUNT = app.MetricID(0xf02c810ad839d592)
	// number of times contexts expired in a pipeline
	PIPELINE_CONTEXT_EXPIRED_COUNT = app.MetricID(0xd1449438ff78383d)
	// total accumulative pipeline processing time
	PIPELINE_PROCESSING_TIME_SEC = app.MetricID(0xca8cbbbb26c8eac7)
	// total accumulative pipeline processing time for workflows that ultimately failed
	PIPELINE_PROCESSING_TIME_SEC_FAILED = app.MetricID(0xca8cbbbb26c8eac7)
	// total accumulative time to deliver the message downstream on the pipeline once it has been processed by the command on the pipeline stage
	PIPELINE_CHANNEL_DELIVERY_TIME_SEC = app.MetricID(0xca8cbbbb26c8eac7)
)
