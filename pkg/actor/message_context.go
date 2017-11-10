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

type MessageContext struct {
	*Actor
	Message *Envelope
}

func (a *MessageContext) RestartMessageProcessor() error {
	LOG_EVENT_RESTARTING.Log(a.Actor.logger.Info()).Msg("")
	a.Actor.stoppingMessageProcessor()
	a.Actor.stoppedMessageProcessor()

	a.Actor.messageProcessor = a.Actor.messageProcessorFactory()
	if err := a.Actor.startedMessageProcessor(); err != nil {
		return err
	}
	LOG_EVENT_RESTARTED.Log(a.Actor.logger.Info()).Msg("")
	return nil
}

//func (a *MessageContext) Spawn(props *Props) (*Actor, error) {
//	return a.spawn(props)
//}
