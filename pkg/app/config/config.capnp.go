// Code generated by capnpc-go. DO NOT EDIT.

package config

import (
	capnp "zombiezen.com/go/capnproto2"
	text "zombiezen.com/go/capnproto2/encoding/text"
	schemas "zombiezen.com/go/capnproto2/schemas"
)

type RPCServerSpec struct{ capnp.Struct }

// RPCServerSpec_TypeID is the unique identifier for the type RPCServerSpec.
const RPCServerSpec_TypeID = 0xfc13c8456771ca68

func NewRPCServerSpec(s *capnp.Segment) (RPCServerSpec, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 3})
	return RPCServerSpec{st}, err
}

func NewRootRPCServerSpec(s *capnp.Segment) (RPCServerSpec, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 3})
	return RPCServerSpec{st}, err
}

func ReadRootRPCServerSpec(msg *capnp.Message) (RPCServerSpec, error) {
	root, err := msg.RootPtr()
	return RPCServerSpec{root.Struct()}, err
}

func (s RPCServerSpec) String() string {
	str, _ := text.Marshal(0xfc13c8456771ca68, s.Struct)
	return str
}

func (s RPCServerSpec) RpcServiceSpec() (RPCServiceSpec, error) {
	p, err := s.Struct.Ptr(0)
	return RPCServiceSpec{Struct: p.Struct()}, err
}

func (s RPCServerSpec) HasRpcServiceSpec() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s RPCServerSpec) SetRpcServiceSpec(v RPCServiceSpec) error {
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewRpcServiceSpec sets the rpcServiceSpec field to a newly
// allocated RPCServiceSpec struct, preferring placement in s's segment.
func (s RPCServerSpec) NewRpcServiceSpec() (RPCServiceSpec, error) {
	ss, err := NewRPCServiceSpec(s.Struct.Segment())
	if err != nil {
		return RPCServiceSpec{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s RPCServerSpec) ServerCert() (X509KeyPair, error) {
	p, err := s.Struct.Ptr(1)
	return X509KeyPair{Struct: p.Struct()}, err
}

func (s RPCServerSpec) HasServerCert() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s RPCServerSpec) SetServerCert(v X509KeyPair) error {
	return s.Struct.SetPtr(1, v.Struct.ToPtr())
}

// NewServerCert sets the serverCert field to a newly
// allocated X509KeyPair struct, preferring placement in s's segment.
func (s RPCServerSpec) NewServerCert() (X509KeyPair, error) {
	ss, err := NewX509KeyPair(s.Struct.Segment())
	if err != nil {
		return X509KeyPair{}, err
	}
	err = s.Struct.SetPtr(1, ss.Struct.ToPtr())
	return ss, err
}

func (s RPCServerSpec) CaCert() ([]byte, error) {
	p, err := s.Struct.Ptr(2)
	return []byte(p.Data()), err
}

func (s RPCServerSpec) HasCaCert() bool {
	p, err := s.Struct.Ptr(2)
	return p.IsValid() || err != nil
}

func (s RPCServerSpec) SetCaCert(v []byte) error {
	return s.Struct.SetData(2, v)
}

// RPCServerSpec_List is a list of RPCServerSpec.
type RPCServerSpec_List struct{ capnp.List }

// NewRPCServerSpec creates a new list of RPCServerSpec.
func NewRPCServerSpec_List(s *capnp.Segment, sz int32) (RPCServerSpec_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 3}, sz)
	return RPCServerSpec_List{l}, err
}

func (s RPCServerSpec_List) At(i int) RPCServerSpec { return RPCServerSpec{s.List.Struct(i)} }

func (s RPCServerSpec_List) Set(i int, v RPCServerSpec) error { return s.List.SetStruct(i, v.Struct) }

func (s RPCServerSpec_List) String() string {
	str, _ := text.MarshalList(0xfc13c8456771ca68, s.List)
	return str
}

// RPCServerSpec_Promise is a wrapper for a RPCServerSpec promised by a client call.
type RPCServerSpec_Promise struct{ *capnp.Pipeline }

func (p RPCServerSpec_Promise) Struct() (RPCServerSpec, error) {
	s, err := p.Pipeline.Struct()
	return RPCServerSpec{s}, err
}

func (p RPCServerSpec_Promise) RpcServiceSpec() RPCServiceSpec_Promise {
	return RPCServiceSpec_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p RPCServerSpec_Promise) ServerCert() X509KeyPair_Promise {
	return X509KeyPair_Promise{Pipeline: p.Pipeline.GetPipeline(1)}
}

type RPCClientSpec struct{ capnp.Struct }

// RPCClientSpec_TypeID is the unique identifier for the type RPCClientSpec.
const RPCClientSpec_TypeID = 0xbec6688394d29776

func NewRPCClientSpec(s *capnp.Segment) (RPCClientSpec, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 3})
	return RPCClientSpec{st}, err
}

func NewRootRPCClientSpec(s *capnp.Segment) (RPCClientSpec, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 3})
	return RPCClientSpec{st}, err
}

func ReadRootRPCClientSpec(msg *capnp.Message) (RPCClientSpec, error) {
	root, err := msg.RootPtr()
	return RPCClientSpec{root.Struct()}, err
}

func (s RPCClientSpec) String() string {
	str, _ := text.Marshal(0xbec6688394d29776, s.Struct)
	return str
}

func (s RPCClientSpec) RpcServiceSpec() (RPCServiceSpec, error) {
	p, err := s.Struct.Ptr(0)
	return RPCServiceSpec{Struct: p.Struct()}, err
}

func (s RPCClientSpec) HasRpcServiceSpec() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s RPCClientSpec) SetRpcServiceSpec(v RPCServiceSpec) error {
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewRpcServiceSpec sets the rpcServiceSpec field to a newly
// allocated RPCServiceSpec struct, preferring placement in s's segment.
func (s RPCClientSpec) NewRpcServiceSpec() (RPCServiceSpec, error) {
	ss, err := NewRPCServiceSpec(s.Struct.Segment())
	if err != nil {
		return RPCServiceSpec{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s RPCClientSpec) ClientCert() (X509KeyPair, error) {
	p, err := s.Struct.Ptr(1)
	return X509KeyPair{Struct: p.Struct()}, err
}

func (s RPCClientSpec) HasClientCert() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s RPCClientSpec) SetClientCert(v X509KeyPair) error {
	return s.Struct.SetPtr(1, v.Struct.ToPtr())
}

// NewClientCert sets the clientCert field to a newly
// allocated X509KeyPair struct, preferring placement in s's segment.
func (s RPCClientSpec) NewClientCert() (X509KeyPair, error) {
	ss, err := NewX509KeyPair(s.Struct.Segment())
	if err != nil {
		return X509KeyPair{}, err
	}
	err = s.Struct.SetPtr(1, ss.Struct.ToPtr())
	return ss, err
}

func (s RPCClientSpec) CaCert() ([]byte, error) {
	p, err := s.Struct.Ptr(2)
	return []byte(p.Data()), err
}

func (s RPCClientSpec) HasCaCert() bool {
	p, err := s.Struct.Ptr(2)
	return p.IsValid() || err != nil
}

func (s RPCClientSpec) SetCaCert(v []byte) error {
	return s.Struct.SetData(2, v)
}

// RPCClientSpec_List is a list of RPCClientSpec.
type RPCClientSpec_List struct{ capnp.List }

// NewRPCClientSpec creates a new list of RPCClientSpec.
func NewRPCClientSpec_List(s *capnp.Segment, sz int32) (RPCClientSpec_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 3}, sz)
	return RPCClientSpec_List{l}, err
}

func (s RPCClientSpec_List) At(i int) RPCClientSpec { return RPCClientSpec{s.List.Struct(i)} }

func (s RPCClientSpec_List) Set(i int, v RPCClientSpec) error { return s.List.SetStruct(i, v.Struct) }

func (s RPCClientSpec_List) String() string {
	str, _ := text.MarshalList(0xbec6688394d29776, s.List)
	return str
}

// RPCClientSpec_Promise is a wrapper for a RPCClientSpec promised by a client call.
type RPCClientSpec_Promise struct{ *capnp.Pipeline }

func (p RPCClientSpec_Promise) Struct() (RPCClientSpec, error) {
	s, err := p.Pipeline.Struct()
	return RPCClientSpec{s}, err
}

func (p RPCClientSpec_Promise) RpcServiceSpec() RPCServiceSpec_Promise {
	return RPCServiceSpec_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p RPCClientSpec_Promise) ClientCert() X509KeyPair_Promise {
	return X509KeyPair_Promise{Pipeline: p.Pipeline.GetPipeline(1)}
}

type RPCServiceSpec struct{ capnp.Struct }

// RPCServiceSpec_TypeID is the unique identifier for the type RPCServiceSpec.
const RPCServiceSpec_TypeID = 0xb6e32df5c504ebf2

func NewRPCServiceSpec(s *capnp.Segment) (RPCServiceSpec, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 32, PointerCount: 0})
	return RPCServiceSpec{st}, err
}

func NewRootRPCServiceSpec(s *capnp.Segment) (RPCServiceSpec, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 32, PointerCount: 0})
	return RPCServiceSpec{st}, err
}

func ReadRootRPCServiceSpec(msg *capnp.Message) (RPCServiceSpec, error) {
	root, err := msg.RootPtr()
	return RPCServiceSpec{root.Struct()}, err
}

func (s RPCServiceSpec) String() string {
	str, _ := text.Marshal(0xb6e32df5c504ebf2, s.Struct)
	return str
}

func (s RPCServiceSpec) DomainID() uint64 {
	return s.Struct.Uint64(0)
}

func (s RPCServiceSpec) SetDomainID(v uint64) {
	s.Struct.SetUint64(0, v)
}

func (s RPCServiceSpec) AppId() uint64 {
	return s.Struct.Uint64(8)
}

func (s RPCServiceSpec) SetAppId(v uint64) {
	s.Struct.SetUint64(8, v)
}

func (s RPCServiceSpec) ServiceId() uint64 {
	return s.Struct.Uint64(16)
}

func (s RPCServiceSpec) SetServiceId(v uint64) {
	s.Struct.SetUint64(16, v)
}

func (s RPCServiceSpec) Port() uint16 {
	return s.Struct.Uint16(24)
}

func (s RPCServiceSpec) SetPort(v uint16) {
	s.Struct.SetUint16(24, v)
}

// RPCServiceSpec_List is a list of RPCServiceSpec.
type RPCServiceSpec_List struct{ capnp.List }

// NewRPCServiceSpec creates a new list of RPCServiceSpec.
func NewRPCServiceSpec_List(s *capnp.Segment, sz int32) (RPCServiceSpec_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 32, PointerCount: 0}, sz)
	return RPCServiceSpec_List{l}, err
}

func (s RPCServiceSpec_List) At(i int) RPCServiceSpec { return RPCServiceSpec{s.List.Struct(i)} }

func (s RPCServiceSpec_List) Set(i int, v RPCServiceSpec) error { return s.List.SetStruct(i, v.Struct) }

func (s RPCServiceSpec_List) String() string {
	str, _ := text.MarshalList(0xb6e32df5c504ebf2, s.List)
	return str
}

// RPCServiceSpec_Promise is a wrapper for a RPCServiceSpec promised by a client call.
type RPCServiceSpec_Promise struct{ *capnp.Pipeline }

func (p RPCServiceSpec_Promise) Struct() (RPCServiceSpec, error) {
	s, err := p.Pipeline.Struct()
	return RPCServiceSpec{s}, err
}

type X509KeyPair struct{ capnp.Struct }

// X509KeyPair_TypeID is the unique identifier for the type X509KeyPair.
const X509KeyPair_TypeID = 0xf4dd73213f6e70a6

func NewX509KeyPair(s *capnp.Segment) (X509KeyPair, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return X509KeyPair{st}, err
}

func NewRootX509KeyPair(s *capnp.Segment) (X509KeyPair, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return X509KeyPair{st}, err
}

func ReadRootX509KeyPair(msg *capnp.Message) (X509KeyPair, error) {
	root, err := msg.RootPtr()
	return X509KeyPair{root.Struct()}, err
}

func (s X509KeyPair) String() string {
	str, _ := text.Marshal(0xf4dd73213f6e70a6, s.Struct)
	return str
}

func (s X509KeyPair) Key() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return []byte(p.Data()), err
}

func (s X509KeyPair) HasKey() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s X509KeyPair) SetKey(v []byte) error {
	return s.Struct.SetData(0, v)
}

func (s X509KeyPair) Cert() ([]byte, error) {
	p, err := s.Struct.Ptr(1)
	return []byte(p.Data()), err
}

func (s X509KeyPair) HasCert() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s X509KeyPair) SetCert(v []byte) error {
	return s.Struct.SetData(1, v)
}

// X509KeyPair_List is a list of X509KeyPair.
type X509KeyPair_List struct{ capnp.List }

// NewX509KeyPair creates a new list of X509KeyPair.
func NewX509KeyPair_List(s *capnp.Segment, sz int32) (X509KeyPair_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2}, sz)
	return X509KeyPair_List{l}, err
}

func (s X509KeyPair_List) At(i int) X509KeyPair { return X509KeyPair{s.List.Struct(i)} }

func (s X509KeyPair_List) Set(i int, v X509KeyPair) error { return s.List.SetStruct(i, v.Struct) }

func (s X509KeyPair_List) String() string {
	str, _ := text.MarshalList(0xf4dd73213f6e70a6, s.List)
	return str
}

// X509KeyPair_Promise is a wrapper for a X509KeyPair promised by a client call.
type X509KeyPair_Promise struct{ *capnp.Pipeline }

func (p X509KeyPair_Promise) Struct() (X509KeyPair, error) {
	s, err := p.Pipeline.Struct()
	return X509KeyPair{s}, err
}

const schema_db8274f9144abc7e = "x\xda\xac\x92=h\x14_\x14\xc5\xcfyo\xf2\xdf\x7f" +
	" Kv\x98\xed7\x8aM\x02J\"\x044M\x94M" +
	"\x8aD\x03\xf3\xb2,F\x11q\x98}\x9b\x0c\xee\xc7s" +
	"v\x12\x13Q\xc1he\xab`ig\x9b\xceJ\xed\"" +
	"\x92\xd2\xc2\"\x92F\x11\x11;\xd1B\x10Ff\xd6\xfd" +
	"\xc0\x08\x8a\xda\xdd\xb9s\xde\xbb\xbfs\xcf\x1b?\xc3\x13" +
	"bb\xa0 \x01uh\xe0\xbf\xf8\xe3\x07k\xfb\xf3\xe1" +
	"7\x8f\xa0\xb2\xb4\xe2\x1b\x8f\xe7\xf3_\xa2\xcdW\xb02" +
	"\x80\xb3\xca\xd7\xceM&\xd5u\xbe\x03\xe3\xb5\xfb/\xee" +
	"\xddZy\xf6\x14v\x96=\xe9\x80L\x14g\xc5\xae\xa3" +
	"ERyb\x0b\x8c\x1f\x9a\xc6\xf4\x81\xd6\xde\xa7\x1f\xb4" +
	"\xa9bT\xee8\x93\xe9\xa9\x09y\x05\x8cWv./" +
	"\xcf>w\xbe\xfe\xec\xde=\xb9\xeb\xbcO\xab\xb7r\x0b" +
	"\xc7b\xbf\xd9\xa8\x06\xcbG|\xe1\x99\x86\x99Zt\x8b" +
	"%\x1d\xae\x05\xbe.e\x8c\xf6]R\xe5\xa4\x05X\x04" +
	"lo\x1eP\x17%UM\xd0&\xf3L\x9a\xc1Q@" +
	"U$\x95\x11\xb4\x85\xc8S\x00v}\x11P5I\xb5" +
	".h\xcb\xa1<%`\xaf\x8e\x01\xcaH\xaak\x82q" +
	"\xa5Y\xf7\x82\xc6\xdc\x0c\x00\x0eBp\x10,x\xc6\xcc" +
	"U:_q\xab\xcd1\x07v{\xc3\xa6\x19F\xcc@" +
	"0\x03\xee'/\xd6\x02\xdd\x88JF\xb6\xc1\x87\xba\xe0" +
	"\xb3W\x015#\xa9\xdc>\xf0\x85s\x80:-\xa9\x96" +
	"\xfa\xc0\xcbSv\xb9\xa0\xd6%\xd5m\xc184~\xba" +
	"\x0dL\xfb\xbad\xb4\xcf\\/^\x90\xb9\x04\"\x9dY" +
	"\xd4\x90a\xc4\\/\xa5\xf6\xefi\xdf+\xea0R\x16" +
	"E|\xe1\xee\x03\xf5\xe4\xe5\x9dm(K\xf0d\x9e\x1c" +
	"\x02ln\xc6\xee\xec\xc2H5\xa8i\x8eT\x9ba\xdd" +
	"\x8b\x92\x8dd!\x98\xddgqir\xfc\xf8)\xbd\xe1" +
	"zA\x08$\x0e\xff\xef:\x1c=h\x8f\x16\xban:" +
	"\x16\xcbc}n2\x97\xf4\xc6\x9f\x82\x0c\xfb\xff\xce\xc5" +
	"\xf7'\xa6\xc3\xbf\x0c\x0aP\xae\xa4:\xff{9\xb5\xd2" +
	"\x91\xbf\xc8\xa9C\xfc-\x00\x00\xff\xff\x96x\xf3w"

func init() {
	schemas.Register(schema_db8274f9144abc7e,
		0xb6e32df5c504ebf2,
		0xbec6688394d29776,
		0xf4dd73213f6e70a6,
		0xfc13c8456771ca68)
}
