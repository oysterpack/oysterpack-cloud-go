// Code generated by capnpc-go. DO NOT EDIT.

package msgs

import (
	capnp "zombiezen.com/go/capnproto2"
	text "zombiezen.com/go/capnproto2/encoding/text"
	schemas "zombiezen.com/go/capnproto2/schemas"
)

type Envelope struct{ capnp.Struct }

// Envelope_TypeID is the unique identifier for the type Envelope.
const Envelope_TypeID = 0xf38cccd618967ecd

func NewEnvelope(s *capnp.Segment) (Envelope, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 16, PointerCount: 4})
	return Envelope{st}, err
}

func NewRootEnvelope(s *capnp.Segment) (Envelope, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 16, PointerCount: 4})
	return Envelope{st}, err
}

func ReadRootEnvelope(msg *capnp.Message) (Envelope, error) {
	root, err := msg.RootPtr()
	return Envelope{root.Struct()}, err
}

func (s Envelope) String() string {
	str, _ := text.Marshal(0xf38cccd618967ecd, s.Struct)
	return str
}

func (s Envelope) Id() (string, error) {
	p, err := s.Struct.Ptr(0)
	return p.Text(), err
}

func (s Envelope) HasId() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Envelope) IdBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return p.TextBytes(), err
}

func (s Envelope) SetId(v string) error {
	return s.Struct.SetText(0, v)
}

func (s Envelope) Created() int64 {
	return int64(s.Struct.Uint64(0))
}

func (s Envelope) SetCreated(v int64) {
	s.Struct.SetUint64(0, uint64(v))
}

func (s Envelope) ReplyTo() (ChannelAddress, error) {
	p, err := s.Struct.Ptr(1)
	return ChannelAddress{Struct: p.Struct()}, err
}

func (s Envelope) HasReplyTo() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s Envelope) SetReplyTo(v ChannelAddress) error {
	return s.Struct.SetPtr(1, v.Struct.ToPtr())
}

// NewReplyTo sets the replyTo field to a newly
// allocated ChannelAddress struct, preferring placement in s's segment.
func (s Envelope) NewReplyTo() (ChannelAddress, error) {
	ss, err := NewChannelAddress(s.Struct.Segment())
	if err != nil {
		return ChannelAddress{}, err
	}
	err = s.Struct.SetPtr(1, ss.Struct.ToPtr())
	return ss, err
}

func (s Envelope) Channel() (string, error) {
	p, err := s.Struct.Ptr(2)
	return p.Text(), err
}

func (s Envelope) HasChannel() bool {
	p, err := s.Struct.Ptr(2)
	return p.IsValid() || err != nil
}

func (s Envelope) ChannelBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(2)
	return p.TextBytes(), err
}

func (s Envelope) SetChannel(v string) error {
	return s.Struct.SetText(2, v)
}

func (s Envelope) MessageType() uint8 {
	return s.Struct.Uint8(8)
}

func (s Envelope) SetMessageType(v uint8) {
	s.Struct.SetUint8(8, v)
}

func (s Envelope) Message() ([]byte, error) {
	p, err := s.Struct.Ptr(3)
	return []byte(p.Data()), err
}

func (s Envelope) HasMessage() bool {
	p, err := s.Struct.Ptr(3)
	return p.IsValid() || err != nil
}

func (s Envelope) SetMessage(v []byte) error {
	return s.Struct.SetData(3, v)
}

// Envelope_List is a list of Envelope.
type Envelope_List struct{ capnp.List }

// NewEnvelope creates a new list of Envelope.
func NewEnvelope_List(s *capnp.Segment, sz int32) (Envelope_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 16, PointerCount: 4}, sz)
	return Envelope_List{l}, err
}

func (s Envelope_List) At(i int) Envelope { return Envelope{s.List.Struct(i)} }

func (s Envelope_List) Set(i int, v Envelope) error { return s.List.SetStruct(i, v.Struct) }

func (s Envelope_List) String() string {
	str, _ := text.MarshalList(0xf38cccd618967ecd, s.List)
	return str
}

// Envelope_Promise is a wrapper for a Envelope promised by a client call.
type Envelope_Promise struct{ *capnp.Pipeline }

func (p Envelope_Promise) Struct() (Envelope, error) {
	s, err := p.Pipeline.Struct()
	return Envelope{s}, err
}

func (p Envelope_Promise) ReplyTo() ChannelAddress_Promise {
	return ChannelAddress_Promise{Pipeline: p.Pipeline.GetPipeline(1)}
}

type Address struct{ capnp.Struct }

// Address_TypeID is the unique identifier for the type Address.
const Address_TypeID = 0x9fd358f04cb684bd

func NewAddress(s *capnp.Segment) (Address, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return Address{st}, err
}

func NewRootAddress(s *capnp.Segment) (Address, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return Address{st}, err
}

func ReadRootAddress(msg *capnp.Message) (Address, error) {
	root, err := msg.RootPtr()
	return Address{root.Struct()}, err
}

func (s Address) String() string {
	str, _ := text.Marshal(0x9fd358f04cb684bd, s.Struct)
	return str
}

func (s Address) Path() (capnp.TextList, error) {
	p, err := s.Struct.Ptr(0)
	return capnp.TextList{List: p.List()}, err
}

func (s Address) HasPath() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Address) SetPath(v capnp.TextList) error {
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewPath sets the path field to a newly
// allocated capnp.TextList, preferring placement in s's segment.
func (s Address) NewPath(n int32) (capnp.TextList, error) {
	l, err := capnp.NewTextList(s.Struct.Segment(), n)
	if err != nil {
		return capnp.TextList{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

func (s Address) Id() (string, error) {
	p, err := s.Struct.Ptr(1)
	return p.Text(), err
}

func (s Address) HasId() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s Address) IdBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(1)
	return p.TextBytes(), err
}

func (s Address) SetId(v string) error {
	return s.Struct.SetText(1, v)
}

// Address_List is a list of Address.
type Address_List struct{ capnp.List }

// NewAddress creates a new list of Address.
func NewAddress_List(s *capnp.Segment, sz int32) (Address_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2}, sz)
	return Address_List{l}, err
}

func (s Address_List) At(i int) Address { return Address{s.List.Struct(i)} }

func (s Address_List) Set(i int, v Address) error { return s.List.SetStruct(i, v.Struct) }

func (s Address_List) String() string {
	str, _ := text.MarshalList(0x9fd358f04cb684bd, s.List)
	return str
}

// Address_Promise is a wrapper for a Address promised by a client call.
type Address_Promise struct{ *capnp.Pipeline }

func (p Address_Promise) Struct() (Address, error) {
	s, err := p.Pipeline.Struct()
	return Address{s}, err
}

type ChannelAddress struct{ capnp.Struct }

// ChannelAddress_TypeID is the unique identifier for the type ChannelAddress.
const ChannelAddress_TypeID = 0xd801266d9df371b7

func NewChannelAddress(s *capnp.Segment) (ChannelAddress, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return ChannelAddress{st}, err
}

func NewRootChannelAddress(s *capnp.Segment) (ChannelAddress, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return ChannelAddress{st}, err
}

func ReadRootChannelAddress(msg *capnp.Message) (ChannelAddress, error) {
	root, err := msg.RootPtr()
	return ChannelAddress{root.Struct()}, err
}

func (s ChannelAddress) String() string {
	str, _ := text.Marshal(0xd801266d9df371b7, s.Struct)
	return str
}

func (s ChannelAddress) Channel() (string, error) {
	p, err := s.Struct.Ptr(0)
	return p.Text(), err
}

func (s ChannelAddress) HasChannel() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s ChannelAddress) ChannelBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return p.TextBytes(), err
}

func (s ChannelAddress) SetChannel(v string) error {
	return s.Struct.SetText(0, v)
}

func (s ChannelAddress) Address() (Address, error) {
	p, err := s.Struct.Ptr(1)
	return Address{Struct: p.Struct()}, err
}

func (s ChannelAddress) HasAddress() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s ChannelAddress) SetAddress(v Address) error {
	return s.Struct.SetPtr(1, v.Struct.ToPtr())
}

// NewAddress sets the address field to a newly
// allocated Address struct, preferring placement in s's segment.
func (s ChannelAddress) NewAddress() (Address, error) {
	ss, err := NewAddress(s.Struct.Segment())
	if err != nil {
		return Address{}, err
	}
	err = s.Struct.SetPtr(1, ss.Struct.ToPtr())
	return ss, err
}

// ChannelAddress_List is a list of ChannelAddress.
type ChannelAddress_List struct{ capnp.List }

// NewChannelAddress creates a new list of ChannelAddress.
func NewChannelAddress_List(s *capnp.Segment, sz int32) (ChannelAddress_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2}, sz)
	return ChannelAddress_List{l}, err
}

func (s ChannelAddress_List) At(i int) ChannelAddress { return ChannelAddress{s.List.Struct(i)} }

func (s ChannelAddress_List) Set(i int, v ChannelAddress) error { return s.List.SetStruct(i, v.Struct) }

func (s ChannelAddress_List) String() string {
	str, _ := text.MarshalList(0xd801266d9df371b7, s.List)
	return str
}

// ChannelAddress_Promise is a wrapper for a ChannelAddress promised by a client call.
type ChannelAddress_Promise struct{ *capnp.Pipeline }

func (p ChannelAddress_Promise) Struct() (ChannelAddress, error) {
	s, err := p.Pipeline.Struct()
	return ChannelAddress{s}, err
}

func (p ChannelAddress_Promise) Address() Address_Promise {
	return Address_Promise{Pipeline: p.Pipeline.GetPipeline(1)}
}

type MessageProcessingError struct{ capnp.Struct }

// MessageProcessingError_TypeID is the unique identifier for the type MessageProcessingError.
const MessageProcessingError_TypeID = 0xa70dd5f5d238faaa

func NewMessageProcessingError(s *capnp.Segment) (MessageProcessingError, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 3})
	return MessageProcessingError{st}, err
}

func NewRootMessageProcessingError(s *capnp.Segment) (MessageProcessingError, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 3})
	return MessageProcessingError{st}, err
}

func ReadRootMessageProcessingError(msg *capnp.Message) (MessageProcessingError, error) {
	root, err := msg.RootPtr()
	return MessageProcessingError{root.Struct()}, err
}

func (s MessageProcessingError) String() string {
	str, _ := text.Marshal(0xa70dd5f5d238faaa, s.Struct)
	return str
}

func (s MessageProcessingError) Path() (capnp.TextList, error) {
	p, err := s.Struct.Ptr(0)
	return capnp.TextList{List: p.List()}, err
}

func (s MessageProcessingError) HasPath() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s MessageProcessingError) SetPath(v capnp.TextList) error {
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewPath sets the path field to a newly
// allocated capnp.TextList, preferring placement in s's segment.
func (s MessageProcessingError) NewPath(n int32) (capnp.TextList, error) {
	l, err := capnp.NewTextList(s.Struct.Segment(), n)
	if err != nil {
		return capnp.TextList{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

func (s MessageProcessingError) Message() (MessageProcessingError_Message, error) {
	p, err := s.Struct.Ptr(1)
	return MessageProcessingError_Message{Struct: p.Struct()}, err
}

func (s MessageProcessingError) HasMessage() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s MessageProcessingError) SetMessage(v MessageProcessingError_Message) error {
	return s.Struct.SetPtr(1, v.Struct.ToPtr())
}

// NewMessage sets the message field to a newly
// allocated MessageProcessingError_Message struct, preferring placement in s's segment.
func (s MessageProcessingError) NewMessage() (MessageProcessingError_Message, error) {
	ss, err := NewMessageProcessingError_Message(s.Struct.Segment())
	if err != nil {
		return MessageProcessingError_Message{}, err
	}
	err = s.Struct.SetPtr(1, ss.Struct.ToPtr())
	return ss, err
}

func (s MessageProcessingError) Err() (string, error) {
	p, err := s.Struct.Ptr(2)
	return p.Text(), err
}

func (s MessageProcessingError) HasErr() bool {
	p, err := s.Struct.Ptr(2)
	return p.IsValid() || err != nil
}

func (s MessageProcessingError) ErrBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(2)
	return p.TextBytes(), err
}

func (s MessageProcessingError) SetErr(v string) error {
	return s.Struct.SetText(2, v)
}

// MessageProcessingError_List is a list of MessageProcessingError.
type MessageProcessingError_List struct{ capnp.List }

// NewMessageProcessingError creates a new list of MessageProcessingError.
func NewMessageProcessingError_List(s *capnp.Segment, sz int32) (MessageProcessingError_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 3}, sz)
	return MessageProcessingError_List{l}, err
}

func (s MessageProcessingError_List) At(i int) MessageProcessingError {
	return MessageProcessingError{s.List.Struct(i)}
}

func (s MessageProcessingError_List) Set(i int, v MessageProcessingError) error {
	return s.List.SetStruct(i, v.Struct)
}

func (s MessageProcessingError_List) String() string {
	str, _ := text.MarshalList(0xa70dd5f5d238faaa, s.List)
	return str
}

// MessageProcessingError_Promise is a wrapper for a MessageProcessingError promised by a client call.
type MessageProcessingError_Promise struct{ *capnp.Pipeline }

func (p MessageProcessingError_Promise) Struct() (MessageProcessingError, error) {
	s, err := p.Pipeline.Struct()
	return MessageProcessingError{s}, err
}

func (p MessageProcessingError_Promise) Message() MessageProcessingError_Message_Promise {
	return MessageProcessingError_Message_Promise{Pipeline: p.Pipeline.GetPipeline(1)}
}

type MessageProcessingError_Message struct{ capnp.Struct }

// MessageProcessingError_Message_TypeID is the unique identifier for the type MessageProcessingError_Message.
const MessageProcessingError_Message_TypeID = 0xd187ae75f5896d22

func NewMessageProcessingError_Message(s *capnp.Segment) (MessageProcessingError_Message, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 16, PointerCount: 3})
	return MessageProcessingError_Message{st}, err
}

func NewRootMessageProcessingError_Message(s *capnp.Segment) (MessageProcessingError_Message, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 16, PointerCount: 3})
	return MessageProcessingError_Message{st}, err
}

func ReadRootMessageProcessingError_Message(msg *capnp.Message) (MessageProcessingError_Message, error) {
	root, err := msg.RootPtr()
	return MessageProcessingError_Message{root.Struct()}, err
}

func (s MessageProcessingError_Message) String() string {
	str, _ := text.Marshal(0xd187ae75f5896d22, s.Struct)
	return str
}

func (s MessageProcessingError_Message) Id() (string, error) {
	p, err := s.Struct.Ptr(0)
	return p.Text(), err
}

func (s MessageProcessingError_Message) HasId() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s MessageProcessingError_Message) IdBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return p.TextBytes(), err
}

func (s MessageProcessingError_Message) SetId(v string) error {
	return s.Struct.SetText(0, v)
}

func (s MessageProcessingError_Message) Created() int64 {
	return int64(s.Struct.Uint64(0))
}

func (s MessageProcessingError_Message) SetCreated(v int64) {
	s.Struct.SetUint64(0, uint64(v))
}

func (s MessageProcessingError_Message) ReplyTo() (ChannelAddress, error) {
	p, err := s.Struct.Ptr(1)
	return ChannelAddress{Struct: p.Struct()}, err
}

func (s MessageProcessingError_Message) HasReplyTo() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s MessageProcessingError_Message) SetReplyTo(v ChannelAddress) error {
	return s.Struct.SetPtr(1, v.Struct.ToPtr())
}

// NewReplyTo sets the replyTo field to a newly
// allocated ChannelAddress struct, preferring placement in s's segment.
func (s MessageProcessingError_Message) NewReplyTo() (ChannelAddress, error) {
	ss, err := NewChannelAddress(s.Struct.Segment())
	if err != nil {
		return ChannelAddress{}, err
	}
	err = s.Struct.SetPtr(1, ss.Struct.ToPtr())
	return ss, err
}

func (s MessageProcessingError_Message) Channel() (string, error) {
	p, err := s.Struct.Ptr(2)
	return p.Text(), err
}

func (s MessageProcessingError_Message) HasChannel() bool {
	p, err := s.Struct.Ptr(2)
	return p.IsValid() || err != nil
}

func (s MessageProcessingError_Message) ChannelBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(2)
	return p.TextBytes(), err
}

func (s MessageProcessingError_Message) SetChannel(v string) error {
	return s.Struct.SetText(2, v)
}

func (s MessageProcessingError_Message) MessageType() uint8 {
	return s.Struct.Uint8(8)
}

func (s MessageProcessingError_Message) SetMessageType(v uint8) {
	s.Struct.SetUint8(8, v)
}

// MessageProcessingError_Message_List is a list of MessageProcessingError_Message.
type MessageProcessingError_Message_List struct{ capnp.List }

// NewMessageProcessingError_Message creates a new list of MessageProcessingError_Message.
func NewMessageProcessingError_Message_List(s *capnp.Segment, sz int32) (MessageProcessingError_Message_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 16, PointerCount: 3}, sz)
	return MessageProcessingError_Message_List{l}, err
}

func (s MessageProcessingError_Message_List) At(i int) MessageProcessingError_Message {
	return MessageProcessingError_Message{s.List.Struct(i)}
}

func (s MessageProcessingError_Message_List) Set(i int, v MessageProcessingError_Message) error {
	return s.List.SetStruct(i, v.Struct)
}

func (s MessageProcessingError_Message_List) String() string {
	str, _ := text.MarshalList(0xd187ae75f5896d22, s.List)
	return str
}

// MessageProcessingError_Message_Promise is a wrapper for a MessageProcessingError_Message promised by a client call.
type MessageProcessingError_Message_Promise struct{ *capnp.Pipeline }

func (p MessageProcessingError_Message_Promise) Struct() (MessageProcessingError_Message, error) {
	s, err := p.Pipeline.Struct()
	return MessageProcessingError_Message{s}, err
}

func (p MessageProcessingError_Message_Promise) ReplyTo() ChannelAddress_Promise {
	return ChannelAddress_Promise{Pipeline: p.Pipeline.GetPipeline(1)}
}

const schema_87fc44aa3255d2ae = "x\xda\xecU]h\x1cU\x14>\xe7\xde\xddM\x1e\x92" +
	"\xa6\xc3,X\x8ap/\xc5H\x1b\xd2\xd2\xb4\x16\xa3/" +
	"5\xdb\xe6\xa5\xa4\x90\xcb&\xd8\x06\x11nf\xae\xdd\x91" +
	"\xdd\x99\xc9\xdcYc\xc4\xba/B\xa3X\xf0\xa7\x0a\x15" +
	"\xa5\xfa \x92 \x15\x04\x7f(\x04\x14\xccC\xa0>X" +
	"\xb1\xf4\xb5\xe0\x93\xf8 \x94\x82Ha\xe4\xce\xecN\x96" +
	"t\xb3\xbe\x8bow\xe7\x9cs\xf7|\xdf\xf9\xbes\x8f" +
	"~D\x9e!\x13\xc5e\x0a \x0e\x16K\xc9\xc6\xeb\xdf" +
	"\xcc\xfcy\xee\x97k`\xed\xc1\xe4\xfa\xad\xf9c\xeb\xa7" +
	"\x1f\\\x82\"\x19\x00\xb0?\xc5\xaf\xec54\xa7\xcfp" +
	"\x190Y\xff{\xf2\xd6\xfd_\x87?\xdf\x91K\x07\x00" +
	"\x8eK2\x86\xf6\x12y\x04\xc0^!_\x02&\x07\x1a" +
	"o\xdco^\xbf\xf43\x88Q$\xdb\xa5Y\xf6\xd7\xf4" +
	"E\xb4\xb7\xd2\xe3&e\x08\x98|\xbbt\xef\xe3\xc6\xe3" +
	"x\xa7W\x1f\xc5\xd2\xef\xb6U2\xa7\xe1\x92\xe9\xe3\xa7" +
	"\xd7>\xd8w\xfb\xe6[\xf7@\xecA\xd2\x95\\0)" +
	"\x1b\xa5\x1b\xf6\xa6I>\xfeC\xe9Y\x84\xc9\xa4\xa1\xb4" +
	"\x96\x17\xd4\x11td\xe8\x87OO\xb9\xcc\x8d\x94\xd6\xb3" +
	"\x88b\x90\x16\x00\x0a\x08`\x1d\x1a\xb3\x0e11CQ" +
	"\xd4\x08Z\x88e4_\xd5~K1\xb1JQ\\!" +
	"8\x12\xca\xb8&\x0aH\x92\xe7\xdf\xbb&6n\xbf\xb9" +
	"\x09\xa2@p\xaa\x8c8\x04`\xe1B\"\x9d8\x88x" +
	"(\x81\xc65\x00\xdc\x038KM\x90\x98#\xf5\xdc>" +
	"\xc5g\xda\xc5\x9e\x0b\x90\x95\x0c\x01\xe6\xad\xd3\xac\xf5\xb3" +
	"\xd9\xcf\xd9(p\x94\xd6\x9e\x7faz$\x8a\x82H\x14" +
	"\xb0\x8bn\x0b+\xadv\xa2\x18\xca\xf1M\x8fY\xd3L" +
	"\xc4\x14\xc5\xdb]\xf8.W\xac\xcbL|OQ\xdc$" +
	"h\x11RF\x02`m\x1d\xb0\xb6\x98\xf8\x83\xa2\xf8k" +
	"w\xd4<m|\x02\x8fa\x0e\x9b\xc65~\x98/\xd7" +
	"T\xa4x\\S\\\x99\xe6x\xe08\xac\x19E\xca}" +
	"\x88\x91V\x1b^\x8f\xdb\x1fk\xdf\x1ea2\xdd\xbe\xa5" +
	"\xe0\xa4\xb7\xf0\xe5\x9aWW<\xcc9\xe0q\xcd\xd3\xbc" +
	"\xa14K\xa9\x02\xc0\xbd\xdbd\x00\xe2^\xc0\x01\x15E" +
	"}\xb8\x8f\xda\xff\xd1P\xc0\xd2\x86\x00\x1e\x9a@a\xb7" +
	"\x09\x98\xca#g\x95\x1eI\xbf#\x8ar\xce\xf9\xc5\xfd" +
	"\xd6E&>\xa1(\xbe \xd8\xa1|\xadb\xad1\xf1" +
	"\x80bu\x10\x0d\xe7\x98rn\x17\xb1b\x17\x91U\xc7" +
	"\x91bu\xd2D()#\x05\xb0O`\xc5>\x81\xac" +
	"Z3\x91\xd8D\x0a\x83e, \xdaK\xb8h7\x91" +
	"U\xbf3\x91\x1f\x91\xfc\x8b\xc4\x16:`8\xd0Tf" +
	"m\x8c-'R2V\xbdJOw\xc6@\xf3\xda!" +
	"\xe5\xbf\xa4\xeaA\xa8xZ\xe6\x05>\x8f\xbd\x86\xd2\xb1" +
	"l\x84\xfc0\x97\x9aK>\xef{/\xa7_\xc7S\x1d" +
	"\xf8\xcd\xc6\xa2\x8ax\xf0\x02\xd7\xca\x09|WsU\x97" +
	"\xa1V.\xd7\x9e\xef(~F\xfaM\x19\xad\xf0\x89q" +
	">\xf1\xd4\x93\xec(\x9f\x9f;\x05\x80E X\x04l" +
	"E*\xac\xaf\xcc\x05=\xda\xdb\xd7Fv#\x09B\xd3" +
	"\x89\xac#O\xb3\xb9<\xe9\xa6&O\xd5\x90\xaf\x96L" +
	"\x0d-\xa7&}_\xd5\xfb\xc8n\x01\x13\xd3\xb8I," +
	"\xf8\xaa\x9e\xa2\xe8\x10\xe0i.\xb5\x0e\x1c\xcfp\xc6\x97" +
	"\xbd\xcc\xeb;\xf42\x07\x03+a/a?\xd1a\x94" +
	"\xe4\x8c\x0e\xc6+azmS\xa7\xf2V>o4\xeb" +
	"\xb1\x17\xd6\x157!m\x98k'k\xeeH\x9f/*" +
	"\xae\x95\x1f\xf3\xc0\xe7\x92;5\x96\xc2\x01\xc0\x12\x10," +
	"u\x89\x96d\xa2=\x95\xe1\x9dr\xdd\x91\x1e\x8b\xafb" +
	"\x16\xdf,E\xf1\\\xd7b8_\xb1\xce3\xf1*E" +
	"\xb1J\xfa\x11\xd6\x99\xc0z\x8e\x06\xdb\xc9|\xc4\x97\x8d" +
	".'\xb5d6\x91\xbe>\xcc6\x89t\x81mO/" +
	"\x7f\xa0\xb2\xe9\xed\\\xe7\xd3\xfe\xc9L\x8f\x06\xd6\xbe\x1c" +
	"\xd6\xd5\xfd\xd6\xd5|\xb3uPmU\xac-V}\xd4" +
	"\xd8\xe5`\xb7\xf9F\xb1b\x8f\"\xab\xce\x98\xc8\xb9n" +
	"\xf3\xcdc\xc5\x9eGV]5\x91+\xdd\xe6{\x07\x17" +
	"\xed\xf7\x91U\xef\x98\xc8o&R\xa4e,\x02\xd8w" +
	"\xb1b\xdfEV\x1d$\x14\xabe\xb2\x9b-;\xd4\xbd" +
	"\x9b4}o\xa9\xa9x\x03;\x0a\xc3\xff\xdd\xf9\xdfu" +
	"g\x9fW\xaf\xc3\xd8\x87\x89V\x91'\xeb\xde+\xa8\xdc" +
	"\xecr\x9a\xbeJ\xc3@p\x18\xf0\x9f\x00\x00\x00\xff\xff" +
	"@\x81\xb7\xe2"

func init() {
	schemas.Register(schema_87fc44aa3255d2ae,
		0x9fd358f04cb684bd,
		0xa70dd5f5d238faaa,
		0xd187ae75f5896d22,
		0xd801266d9df371b7,
		0xf38cccd618967ecd)
}
