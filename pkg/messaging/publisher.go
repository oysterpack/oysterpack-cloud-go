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

package messaging

import (
	"context"
	"time"
)

// Publisher is used to publish data to the specified topic.
// Replies is used to communicate to the subscriber where to send replies to.
// Messages are sent in fire and forget fashion.
type Publisher interface {
	Cluster() ClusterName
	Topic() Topic

	Publish(data []byte) error
	PublishRequest(data []byte, replyTo ReplyTo) error

	Request(data []byte, timeout time.Duration) (response *Message, err error)
	RequestWithContext(ctx context.Context, data []byte) (response *Message, err error)

	AsyncRequest(data []byte, timeout time.Duration, handler func(Response))
	AsyncRequestWithContext(ctx context.Context, data []byte, handler func(Response))

	RequestChannel(data []byte, timeout time.Duration) <-chan Response
	RequestChannelWithContext(ctx context.Context, data []byte) <-chan Response
}
