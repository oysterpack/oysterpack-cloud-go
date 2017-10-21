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

package domain

import (
	"time"

	"zombiezen.com/go/capnproto2"
)

// MessageId unique message id
type MessageId string

// MessageType message type - used to communicate the message body type
type MessageType string

type ActionRequest interface {
	// Id unique message id for tracking purposes
	Id() MessageId

	// Type message type
	Type() MessageType

	// Created when the message was created
	Created() time.Time

	// SubjectId who initiated the request
	SubjectId() SubjectId

	// MessageSequence is used when messages are streamed in a request. The message sequence starts at 1.
	MessageSequence() uint32

	// FinalMessage is used to indicate that this is the final message in a stream.
	// For non-streamed messages, this will always be true
	FinalMessage() bool

	// Body the request message body
	// If the message fails parsing, then an error is returned.
	Body() (*capnp.Message, error)
}

type ActionResponse interface {
	// Id unique message id for tracking purposes
	Id() MessageId

	// RequestId correlates the response message to the request message
	RequestId() MessageId

	// Type message type
	Type() MessageType

	// Created when the message was created
	Created() time.Time

	// MessageSequence is used when messages are streamed in a request. The message sequence starts at 1.
	MessageSequence() uint32

	// FinalMessage is used to indicate that this is the final message in a stream.
	// For non-streamed messages, this will always be true
	FinalMessage() bool

	// Body returns the response message payload.
	// If the action status indicates a failure, then there will be no body.
	// If the message fails parsing, then an error is returned.
	Body() (*capnp.Message, error)

	// Status reports if the request was successfully processed.
	Status() ActionStatus
}

// ActionStatusCode status code
type ActionStatusCode int32

// ActionStatus reports the action request status
type ActionStatus interface {
	Success() bool

	Code() ActionStatusCode

	Message() string
}
