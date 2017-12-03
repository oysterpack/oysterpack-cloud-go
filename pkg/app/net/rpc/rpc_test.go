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

package rpc

import (
	"net"
	"sync"
	"testing"

	"time"

	"fmt"

	"github.com/oysterpack/oysterpack.go/pkg/app/message"
	"github.com/oysterpack/oysterpack.go/pkg/app/uid"
	"github.com/rs/zerolog/log"
	"gopkg.in/tomb.v2"
	"zombiezen.com/go/capnproto2"
)

func TestRPCMessaging(t *testing.T) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	listenerRunning := sync.WaitGroup{}
	listenerRunning.Add(1)
	listenerTomb := tomb.Tomb{}

	listenerTomb.Go(func() error {
		listenerRunning.Done()
		defer log.Info().Msg("Listener handler is dead")
		log.Info().Msg("Listener handler is alive")
		for {
			select {
			case <-listenerTomb.Dying():
				return nil
			default:
				conn, err := l.Accept()
				log.Info().Msg("New connection")
				if err != nil {
					t.Error(err)
					return err
				}

				listenerTomb.Go(func() error {
					defer log.Info().Msg("Connection handler is dead")
					log.Info().Msg("Connection handler is alive")
					defer conn.Close()

					encoder := capnp.NewPackedEncoder(conn)
					decoder := capnp.NewPackedDecoder(conn)
					decoder.ReuseBuffer()

					msgCounter := 0

					log.Info().Msg("Connection handler is initialized")
					for {

						select {
						case <-listenerTomb.Dying():
							return nil
						default:
							if msg, err := decoder.Decode(); err == nil {
								msgCounter++
								log.Info().Msgf("Received message #%d", msgCounter)

								m, err := message.ReadRootMessage(msg)
								if err != nil {
									log.Error().Err(err).Msg("message.ReadRootMessage(msg) failed")
									return err
								}

								log.Info().Msgf("MSG#%d : parsed", msgCounter)

								data, err := m.Data()
								if err != nil {
									log.Error().Err(err).Msg("Message.Data() failed")
									return err
								}
								log.Info().Msgf("request message : id = 0x%x : type = 0x%x, timestamp = %v, correlationId = 0x%x, len(data) = %d",
									m.Id(),
									m.Type(),
									time.Unix(0, int64(m.Timestamp())),
									m.CorrelationID(),
									len(data),
								)

								// echo back the message
								encoder.Encode(msg)
								log.Info().Msgf("MSG#%d : response sent", msgCounter)
							} else {
								log.Error().Err(err).Msg("decoder.Decode() failed")
								return err
							}
						}
					}

				})

			}
		}
	})

	listenerRunning.Wait()

	if !listenerTomb.Alive() {
		t.Fatal("listener server is not alive")
	}

	log.Info().Msgf("listener address : %s:%s", l.Addr().Network(), l.Addr().String())
	conn, err := net.Dial(l.Addr().Network(), l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	log.Info().Msg("client is connected")

	encoder := capnp.NewPackedEncoder(conn)
	decoder := capnp.NewPackedDecoder(conn)
	decoder.ReuseBuffer()

	const REQ_COUNT = 10

	for i := 0; i < REQ_COUNT; i++ {
		msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			t.Fatal(err)
		}

		req, err := message.NewRootMessage(seg)
		if err != nil {
			t.Fatal(err)
		}

		req.SetId(uint64(uid.NextUIDHash()))
		req.SetType(uint64(uid.NextUIDHash()))
		req.SetCorrelationID(uint64(uid.NextUIDHash()))
		req.SetTimestamp(uint64(time.Now().UnixNano()))
		if err := req.SetData([]byte(fmt.Sprintf("REQ #%d", i))); err != nil {
			t.Fatal(err)
		}

		// write request
		encoder.Encode(msg)

		// read response
		responseMsg, err := decoder.Decode()

		if err != nil {
			t.Fatal(err)
		}
		m, err := message.ReadRootMessage(responseMsg)
		if err != nil {
			t.Fatal(err)
		}
		data, err := m.Data()
		if err != nil {
			t.Fatal(err)
		}
		log.Info().Msgf("response message : id = 0x%x : type = 0x%x, timestamp = %v, correlationId = 0x%x, len(data) = %d, data = %s",
			m.Id(),
			m.Type(),
			time.Unix(0, int64(m.Timestamp())),
			m.CorrelationID(),
			len(data),
			string(data),
		)
	}

}
