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
	"strings"
)

// Topic represents the name of a messaging topic
type Topic string

// Validate checks that the topic is not blank
func (a Topic) Validate() error {
	if strings.TrimSpace(string(a)) == "" {
		return ErrTopicMustNotBeBlank
	}
	return nil
}

// TrimSpace returns a new Topic with whitespace trimmed
func (a Topic) TrimSpace() Topic {
	return Topic(strings.TrimSpace(string(a)))
}

func (a Topic) AsReplyTo() ReplyTo {
	return ReplyTo(a)
}

func (a Topic) AsQueue() Queue {
	return Queue(a)
}

// Queue represents the name of a messaging queue
type Queue string

// Message represents the message envelope
type Message struct {
	// Topic is the topic that the message is published to or received from. This is required and must never be blank
	Topic Topic
	// ReplyTo is intended to support a request-response model. This is optional, i.e., a blank value means no reply is specified.
	ReplyTo ReplyTo
	// Data is the message data
	Data []byte
}

// ReplyTo is the topic name to send replies to
type ReplyTo string

// Validate checks that the reply topic is not blank
func (a ReplyTo) Validate() error {
	if strings.TrimSpace(string(a)) == "" {
		return ErrReplyToMustNotBeBlank
	}
	return nil
}

// TrimSpace returns a new ReplyTo with whitespace trimmed
func (a ReplyTo) TrimSpace() ReplyTo {
	return ReplyTo(strings.TrimSpace(string(a)))
}

func (a ReplyTo) AsTopic() Topic {
	return Topic(a)
}

// Response is used for request-response messaging
type Response struct {
	*Message
	Error error
}

// Success returns true if there was no error reported
func (a *Response) Success() bool {
	return a.Error == nil
}
