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

package app_test

import (
	"fmt"
	"net"
	"testing"

	"context"
	"time"

	"bytes"

	"compress/zlib"

	"io"

	"runtime"

	"github.com/oysterpack/oysterpack.go/pkg/app"
	"github.com/oysterpack/oysterpack.go/pkg/app/capnprpc"
	"gopkg.in/tomb.v2"
	"zombiezen.com/go/capnproto2/rpc"
)

func appClient(addr net.Addr) capnprpc.App {
	clientConn, err := net.Dial(addr.Network(), addr.String())
	if err != nil {
		panic(err)
	}
	rpcClient := rpc.NewConn(rpc.StreamTransport(clientConn))
	return capnprpc.App{Client: rpcClient.Bootstrap(context.Background())}
}

func appClientConn(addr net.Addr) (capnprpc.App, net.Conn) {
	clientConn, err := net.Dial(addr.Network(), addr.String())
	if err != nil {
		panic(err)
	}
	rpcClient := rpc.NewConn(rpc.StreamTransport(clientConn))
	return capnprpc.App{Client: rpcClient.Bootstrap(context.Background())}, clientConn
}

func TestRPCAppServer_NetworkErrors(t *testing.T) {
	app.Reset()

	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	network := l.Addr().Network()
	address := l.Addr().String()
	t.Logf("%s:/%s", network, address)

	defer l.Close()
	var listenerTomb tomb.Tomb
	var rpcTomb tomb.Tomb
	connLogger := app.NewConnLogger(app.Logger())
	listenerTomb.Go(func() error {
		server := capnprpc.App_ServerToClient(app.NewAppServer())
		for {
			conn, err := l.Accept()
			if err != nil {
				return err
			}
			rpcTomb.Go(func() error {
				rpcConn := rpc.NewConn(rpc.StreamTransport(conn), rpc.MainInterface(server.Client), rpc.ConnLog(connLogger))
				err := rpcConn.Wait()
				t.Logf("rpcConn err : %[1]T : %[1]v", err)
				return nil
			})
		}
	})

	appClient := appClient(l.Addr())

	ctx := context.Background()
	if result, err := appClient.Id(ctx, func(params capnprpc.App_id_Params) error { return nil }).Struct(); err != nil {
		t.Error(err)
	} else {
		t.Logf("id : %v", result.AppId())
	}

	// When the listener is closed
	l.Close()
	t.Logf("listenerTomb err : %[1]T : %[1]v", listenerTomb.Wait())

	// Then no more connections can be made, but active connections should still continue to function
	if result, err := appClient.LogLevel(ctx, func(params capnprpc.App_logLevel_Params) error { return nil }).Struct(); err != nil {
		t.Error(err)
	} else {
		t.Logf("id : %v", result.Level())
	}

	if result, err := appClient.LogLevel(ctx, func(params capnprpc.App_logLevel_Params) error { return nil }).Struct(); err != nil {
		t.Error(err)
	} else {
		t.Logf("id : %v", result.Level())
	}

	// When the client is closed
	appClient.Client.Close()
	if _, err := appClient.LogLevel(ctx, func(params capnprpc.App_logLevel_Params) error { return nil }).Struct(); err == nil {
		t.Error("client is closed")
	} else {
		t.Logf("error on closed client : %[1]T : %[1]v", err)
	}

}

func TestRPCAppServer(t *testing.T) {
	app.Reset()

	rpcServer, err := app.StartRPCAppServer(func() (net.Listener, error) {
		return net.Listen("tcp", ":0")
	}, 10)

	if err != nil {
		t.Fatal(err)
	}
	defer rpcServer.Kill(nil)

	ctx := context.Background()
	<-rpcServer.Started()
	addr, err := rpcServer.ListenerAddress()
	if err != nil {
		t.Fatal(err)
	}
	appClient := appClient(addr)

	t.Run("LogLevel()", func(t *testing.T) {
		if result, err := appClient.LogLevel(ctx, func(params capnprpc.App_logLevel_Params) error {
			return nil
		}).Struct(); err != nil {
			t.Error(err)
		} else {
			if level, err := app.CapnprpcLogLevel2zerologLevel(result.Level()); err != nil {
				t.Error(err)
			} else {
				t.Logf("level : %v", level)
			}
		}

		if result, err := appClient.LogLevel(ctx, func(params capnprpc.App_logLevel_Params) error {
			return nil
		}).Struct(); err != nil {
			t.Error(err)
		} else {
			t.Logf("Log Level : %v", result.Level())
		}
	})

	t.Run("Id()", func(t *testing.T) {
		if result, err := appClient.Id(ctx, func(params capnprpc.App_id_Params) error {
			return nil
		}).Struct(); err != nil {
			t.Error(err)
		} else {
			t.Logf("app id : %v", result.AppId())
			if app.AppID(result.AppId()) != app.ID() {
				t.Errorf("Returned app id did not match : %v != %v", result.AppId(), app.ID())
			}
		}
	})

	t.Run("StartedOn()", func(t *testing.T) {
		if result, err := appClient.StartedOn(ctx, func(params capnprpc.App_startedOn_Params) error {
			return nil
		}).Struct(); err != nil {
			t.Error(err)
		} else {
			startedOn := time.Unix(0, result.StartedOn())
			t.Logf("startedOn: %v", startedOn)
			if !startedOn.Equal(app.StartedOn()) {
				t.Errorf("Returned app startedOn did not match : %v != %v", startedOn, app.StartedOn())
			}
		}
	})

	t.Run("Instance()", func(t *testing.T) {
		if result, err := appClient.Instance(ctx, func(params capnprpc.App_instance_Params) error {
			return nil
		}).Struct(); err != nil {
			t.Error(err)
		} else {
			instanceId, err := result.InstanceId()
			if err != nil {
				t.Error(err)
			} else {
				t.Logf("instance id : %v", instanceId)
				if app.InstanceID(instanceId) != app.Instance() {
					t.Errorf("Returned app instance id did not match : %v != %v", instanceId, app.Instance())
				}
			}
		}
	})

	t.Run("Service() - for registered service", func(t *testing.T) {
		service := app.NewService(app.ServiceID(999))
		app.RegisterService(service)
		remoteService := appClient.Service(ctx, func(params capnprpc.App_service_Params) error {
			params.SetId(uint64(service.ID()))
			return nil
		}).Service()

		if results, err := remoteService.Id(ctx, func(params capnprpc.Service_id_Params) error {
			return nil
		}).Struct(); err != nil {
			t.Error(err)
		} else {
			t.Logf("remote service id = %v", results.ServiceId())
			if results.ServiceId() != uint64(service.ID()) {
				t.Errorf("service id did not match : %v", results.ServiceId())
			}
		}
	})

	t.Run("Service() - for unregistered service", func(t *testing.T) {
		const SERVICE_ID = app.ServiceID(987654321)

		// When an invalid ServiceID is used to lookup a Service
		// Then an error should be returned.
		if _, err := appClient.Service(ctx, func(params capnprpc.App_service_Params) error {
			params.SetId(uint64(SERVICE_ID))
			return nil
		}).Struct(); err == nil {
			t.Error("No service should have been found")
		} else {
			t.Logf("rpc error : %T : %v", err, err)
		}

		remoteService := appClient.Service(ctx, func(params capnprpc.App_service_Params) error {
			params.SetId(uint64(SERVICE_ID))
			return nil
		}).Service()

		if _, err := remoteService.Id(ctx, func(params capnprpc.Service_id_Params) error {
			return nil
		}).Struct(); err == nil {
			t.Error("The service references should be invalid")
		} else {
			t.Logf("rpc error : %T : %v", err, err)
		}

		// When the service is registered
		service1000 := app.NewService(app.ServiceID(SERVICE_ID))
		if err := app.RegisterService(service1000); err != nil {
			t.Fatal(err)
		}

		// Then the remote service reference is still invalid, because the error is cached in the pipeline
		if _, err := remoteService.Id(ctx, func(params capnprpc.Service_id_Params) error {
			return nil
		}).Struct(); err == nil {
			t.Errorf("the remote sevice reference is invalid : %T : %v", err, err)
		} else {
			t.Logf("remote error : %T : %v", err, err)
		}

		// When a new remote service reference is retrieved
		remoteService = appClient.Service(ctx, func(params capnprpc.App_service_Params) error {
			params.SetId(uint64(SERVICE_ID))
			return nil
		}).Service()

		// Then it works fine
		if result, err := remoteService.Id(ctx, func(params capnprpc.Service_id_Params) error {
			return nil
		}).Struct(); err != nil {
			t.Errorf("service should now be registered : %T : %v", err, err)
		} else {
			if app.ServiceID(result.ServiceId()) != service1000.ID() {
				t.Errorf("Service id did not match : %d", result.ServiceId())
			}
		}

		// When the service is killed
		service1000.Kill(nil)
		service1000.Wait()

		// Then the remote service reference becomes invalid
		if result, err := remoteService.LogLevel(ctx, func(params capnprpc.Service_logLevel_Params) error {
			return nil
		}).Struct(); err == nil {
			t.Errorf("The service reference should be invalid : %v", result.Level())
		} else {
			t.Logf("rpc error : %T : %v", err, err)
		}

	})

	t.Run("Service.Kill()", func(t *testing.T) {
		const SERVICE_ID = app.ServiceID(987654321)

		// Given a registered service
		service1000 := app.NewService(SERVICE_ID)
		if err := app.RegisterService(service1000); err != nil {
			t.Fatal(err)
		}
		// When the service is killed remotely
		remoteService := appClient.Service(ctx, func(params capnprpc.App_service_Params) error {
			params.SetId(uint64(service1000.ID()))
			return nil
		}).Service()
		if result, err := remoteService.Alive(ctx, func(params capnprpc.Service_alive_Params) error {
			return nil
		}).Struct(); err != nil {
			t.Error(err)
		} else if !result.Alive() {
			t.Error("service should be alive")
		}
		remoteService.Kill(ctx, func(params capnprpc.Service_kill_Params) error {
			return nil
		})

		// Then the service should no longer be alive
		if result, err := remoteService.Alive(ctx, func(params capnprpc.Service_alive_Params) error {
			return nil
		}).Struct(); err != nil {
			t.Error(err)
		} else if result.Alive() {
			t.Error("service should not be alive")
		}
	})

	t.Run("Runtime().GoVersion()", func(t *testing.T) {
		appRuntime := appClient.Runtime(ctx, func(params capnprpc.App_runtime_Params) error {
			return nil
		}).Runtime()

		if result, err := appRuntime.GoVersion(ctx, func(params capnprpc.Runtime_goVersion_Params) error { return nil }).Struct(); err != nil {
			t.Fatal(err)
		} else {
			goVersion, err := result.Version()
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("go version: %v", goVersion)
		}

	})

	t.Run("Runtime().NumGoroutine().StackDump()", func(t *testing.T) {
		appRuntime := appClient.Runtime(ctx, func(params capnprpc.App_runtime_Params) error {
			return nil
		}).Runtime()

		if result, err := appRuntime.NumGoroutine(ctx, func(params capnprpc.Runtime_numGoroutine_Params) error {
			return nil
		}).Struct(); err != nil {
			t.Errorf("NumGoroutine failure: %v", err)
		} else {
			t.Logf("num goroutine = %d", result.Count())

			if result, err := appRuntime.StackDump(ctx, func(params capnprpc.Runtime_stackDump_Params) error {
				return nil
			}).Struct(); err != nil {
				t.Errorf("Failed to get stack dump : %v", err)
			} else {
				compressedStackDump, err := result.StackDump()
				if err != nil {
					t.Error(err)
				}
				r, err := zlib.NewReader(bytes.NewReader(compressedStackDump))
				var stackDump bytes.Buffer
				io.Copy(&stackDump, r)
				t.Log(stackDump.String())
			}
		}

		if result, err := appRuntime.NumCPU(ctx, func(params capnprpc.Runtime_numCPU_Params) error {
			return nil
		}).Struct(); err != nil {
			t.Error(err)
		} else {
			t.Logf("NumCPU = %d", result.Count())
		}
	})

	t.Run("Runtime().NumCPU()", func(t *testing.T) {
		appRuntime := appClient.Runtime(ctx, func(params capnprpc.App_runtime_Params) error {
			return nil
		}).Runtime()

		if result, err := appRuntime.NumCPU(ctx, func(params capnprpc.Runtime_numCPU_Params) error {
			return nil
		}).Struct(); err != nil {
			t.Error(err)
		} else {
			t.Logf("NumCPU = %d", result.Count())
		}
	})

	t.Run("Runtime().MemStats()", func(t *testing.T) {
		appRuntime := appClient.Runtime(ctx, func(params capnprpc.App_runtime_Params) error {
			return nil
		}).Runtime()

		if result, err := appRuntime.MemStats(ctx, func(params capnprpc.Runtime_memStats_Params) error {
			return nil
		}).Struct(); err != nil {
			t.Error(err)
		} else {
			if stats, err := result.Stats(); err != nil {
				t.Error(err)
			} else {
				t.Logf("alloc : %d", stats.Alloc())
				t.Logf("totalAlloc : %d", stats.TotalAlloc())
				t.Logf("sys : %d", stats.Sys())
				t.Logf("lookups : %d", stats.Lookups())
				t.Logf("mallocs : %d", stats.Mallocs())
				t.Logf("frees : %d", stats.Frees())

				t.Logf("heapAlloc : %d", stats.HeapAlloc())
				t.Logf("heapSys : %d", stats.HeapSys())
				t.Logf("heapIdle : %d", stats.HeapIdle())
				t.Logf("heapInUse : %d", stats.HeapInUse())
				t.Logf("heapReleased : %d", stats.HeapReleased())
				t.Logf("heapObjects : %d", stats.HeapObjects())

				t.Logf("stackInUse : %d", stats.StackInUse())
				t.Logf("stackSys : %d", stats.StackSys())

				t.Logf("mSpanInUse : %d", stats.MSpanInUse())
				t.Logf("mSpanSys : %d", stats.MSpanSys())
				t.Logf("mCacheInUse : %d", stats.MCacheInUse())
				t.Logf("mCacheSys : %d", stats.MCacheSys())

				t.Logf("buckHashSys : %d", stats.BuckHashSys())
				t.Logf("gCSys : %d", stats.GCSys())
				t.Logf("otherSys : %d", stats.OtherSys())

				t.Logf("nextGC : %d", stats.NextGC())
				t.Logf("lastGC : %d", stats.LastGC())

				t.Logf("pauseTotalNs : %d", stats.PauseTotalNs())

				memstats := &runtime.MemStats{}
				if nl, err := stats.PauseNs(); err != nil {
					t.Error(err)
				} else {
					data := make([]uint64, nl.Len())
					for i := 0; i < nl.Len(); i++ {
						data[i] = nl.At(i)
					}
					t.Logf("pauseNs : %v", data)
					runtime.ReadMemStats(memstats)
					t.Logf("current pauseNs : %v", memstats.PauseNs)
				}
				if nl, err := stats.PauseEnd(); err != nil {
					t.Error(err)
				} else {
					data := make([]uint64, nl.Len())
					for i := 0; i < nl.Len(); i++ {
						data[i] = nl.At(i)
					}
					t.Logf("pauseEnd : %v", data)
					runtime.ReadMemStats(memstats)
					t.Logf("current pauseNs : %v", memstats.PauseEnd)
				}

				t.Logf("numGC : %d", stats.NumGC())
				t.Logf("numForcedGC : %d", stats.NumForcedGC())
				t.Logf("gCCPUFraction : %d", stats.GCCPUFraction())

				if nl, err := stats.BySize(); err != nil {
					t.Error(err)
				} else {
					data := make([]capnprpc.MemStats_BySize, nl.Len())
					for i := 0; i < nl.Len(); i++ {
						data[i] = nl.At(i)
					}
					t.Logf("bySize : %v", data)

				}
			}
		}
	})

}

// BenchmarkRpcAppServer_UnixSocket/ID()-8                    20000             79748 ns/op           18266 B/op         78 allocs/op
// BenchmarkRpcAppServer_UnixSocket/Instance()-8              20000             81747 ns/op           18361 B/op         79 allocs/op
// BenchmarkRpcAppServer_UnixSocket/StartedOn()-8             20000             79975 ns/op           18265 B/op         78 allocs/op
// BenchmarkRpcAppServer_UnixSocket/connect_-_close-8         10000            123790 ns/op           32753 B/op        174 allocs/op
func BenchmarkRpcAppServer_UnixSocket(b *testing.B) {
	app.Reset()
	sockAddress := fmt.Sprintf("/tmp/app-%s.sock", app.Instance())
	b.Logf("sockAddress = %s", sockAddress)

	// start the server
	l, err := net.Listen("unix", sockAddress)
	if err != nil {
		b.Fatal(err)
	}
	defer l.Close()
	var listenerTomb tomb.Tomb
	var rpcTomb tomb.Tomb
	listenerTomb.Go(func() error {
		server := capnprpc.App_ServerToClient(app.NewAppServer())
		for {
			conn, err := l.Accept()
			if err != nil {
				return err
			}
			rpcTomb.Go(func() error {
				rpcConn := rpc.NewConn(rpc.StreamTransport(conn), rpc.MainInterface(server.Client))
				return rpcConn.Wait()
			})
		}
	})

	// create client
	clientConn, err := net.Dial("unix", sockAddress)
	if err != nil {
		b.Fatal(err)
	}
	rpcClient := rpc.NewConn(rpc.StreamTransport(clientConn))
	defer rpcClient.Close()
	ctx := context.Background()
	appClient := capnprpc.App{Client: rpcClient.Bootstrap(ctx)}

	appIdParams := func(params capnprpc.App_id_Params) error { return nil }
	b.Run("ID()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if result, err := appClient.Id(ctx, appIdParams).Struct(); err != nil {
				b.Error(err)
			} else {
				result.AppId()
			}
		}
	})

	instanceParams := func(params capnprpc.App_instance_Params) error { return nil }
	b.Run("Instance()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if result, err := appClient.Instance(ctx, instanceParams).Struct(); err != nil {
				b.Error(err)
			} else {
				result.InstanceId()
			}
		}
	})

	startedOnParams := func(params capnprpc.App_startedOn_Params) error { return nil }
	b.Run("StartedOn()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if result, err := appClient.StartedOn(ctx, startedOnParams).Struct(); err != nil {
				b.Error(err)
			} else {
				result.StartedOn()
			}
		}
	})

	b.Run("connect - close", func(b *testing.B) {
		ctx := context.Background()
		for i := 0; i < b.N; i++ {
			clientConn, err := net.Dial("unix", sockAddress)
			if err != nil {
				b.Fatal(err)
			}
			rpcClient := rpc.NewConn(rpc.StreamTransport(clientConn))
			appClient := capnprpc.App{Client: rpcClient.Bootstrap(ctx)}
			appClient.Client.Close()
		}
	})
}
