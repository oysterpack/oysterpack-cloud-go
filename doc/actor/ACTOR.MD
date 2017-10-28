# Message Design

    type Message interface {
        encoding.BinaryMarshaler
        encoding.BinaryUnmarshaler
    }

- Messages know how to encode/decode themselves to a binary format
    - the recommended message binary encoding is using capnp
      - define corresponding capnp message schemas
      - it's easier to work with plain old go structs, but for marshalling efficiency and performance, capnp messages are optimal
- Messages are wrapped in an Envelope when sending messages between actors.
- All message structs should define fields with public scope
  - Envelope.String() logs the message as JSON, thus if fields are not public, then they will not get picked up.
  
## Envelope

    type Envelope struct {
        id      string
        created time.Time
    
        channel string
        message Message
    
        replyTo *ChannelAddress
    }
    
- Envelope implements the Message interface as well
  - the envelope will compress the message using zlib. Thus, messages will not need to apply any compression at the message level.
    Compression will be applied at the envelope level.