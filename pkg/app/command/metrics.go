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

	COMMAND_MSG_COUNT               = app.MetricID(0xdccbcbcf9235211d)
	COMMAND_ERR_COUNT               = app.MetricID(0xd8b89ef52341ad4b)
	COMMAND_CONTEXT_EXPIRED_COUNT   = app.MetricID(0x86f01c622f5894c5)
	COMMAND_PROCESSING_TIME_NANOSEC = app.MetricID(0x9684d6de9cf15d73)

	PIPELINE_MSG_COUNT               = app.MetricID(0xdf381a0443622af1)
	PIPELINE_ERR_COUNT               = app.MetricID(0xf02c810ad839d592)
	PIPELINE_CONTEXT_EXPIRED_COUNT   = app.MetricID(0xd1449438ff78383d)
	PIPELINE_PROCESSING_TIME_NANOSEC = app.MetricID(0xca8cbbbb26c8eac7)
)
