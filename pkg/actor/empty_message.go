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

package actor

// EmptyMessage implements the Message interface
// It used to for sending messages that act as signals and require no message body.
type EmptyMessage struct{}

func (a EmptyMessage) UnmarshalBinary(data []byte) error {
	return nil
}

func (a EmptyMessage) MarshalBinary() (data []byte, err error) {
	return []byte{}, nil
}
