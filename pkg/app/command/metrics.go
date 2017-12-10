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

	// ping-pong success counter
	PIPELINE_PING_PONG_COUNT = app.MetricID(0xd6129832e634c841)
	// total accumulative time for ping-pong messaging
	PIPELINE_PING_PONG_TIME_SEC = app.MetricID(0x859177a262f568a6)
	// number of ping requests that expired, i.e., did not make it through the entire pipeline
	PIPELINE_PING_EXPIRED_COUNT = app.MetricID(0x8f3a48d571fd33b1)
	// total accumulative time for ping-pong messaging
	PIPELINE_PING_EXPIRED_TIME_SEC = app.MetricID(0xd431ade2dbd030d2)

	// all times are in Unix time: the number of seconds elapsed since January 1, 1970 UTC.
	PIPELINE_LAST_SUCCESS_TIME      = app.MetricID(0xbad3338b90169ac8)
	PIPELINE_LAST_FAILURE_TIME      = app.MetricID(0x833978794af500be)
	PIPELINE_LAST_EXPIRED_TIME      = app.MetricID(0x9f2c4e6187cf5aa9)
	PIPELINE_LAST_PING_SUCCESS_TIME = app.MetricID(0xb110d5e8dc031a82)
	PIPELINE_LAST_PING_EXPIRED_TIME = app.MetricID(0xc92a144a7fc2ead8)
)

var (
	COUNTER_VECTOR_METRIC_IDS = []app.MetricID{
		COMMAND_RUN_COUNT,
		COMMAND_FAILED_COUNT,
		COMMAND_PROCESSING_TIME_SEC,
		COMMAND_PROCESSING_TIME_SEC_FAILED,
	}

	COUNTER_METRIC_IDS = []app.MetricID{
		PIPELINE_RUN_COUNT,
		PIPELINE_FAILED_COUNT,
		PIPELINE_CONTEXT_EXPIRED_COUNT,
		PIPELINE_PROCESSING_TIME_SEC,
		PIPELINE_PROCESSING_TIME_SEC_FAILED,
		PIPELINE_CHANNEL_DELIVERY_TIME_SEC,

		PIPELINE_PING_PONG_COUNT,
		PIPELINE_PING_PONG_TIME_SEC,
		PIPELINE_PING_EXPIRED_COUNT,
		PIPELINE_PING_EXPIRED_TIME_SEC,
	}

	GAUGE_METRIC_IDS = []app.MetricID{
		PIPELINE_LAST_SUCCESS_TIME,
		PIPELINE_LAST_FAILURE_TIME,
		PIPELINE_LAST_EXPIRED_TIME,
		PIPELINE_LAST_PING_SUCCESS_TIME,
		PIPELINE_LAST_PING_EXPIRED_TIME,
	}
)
