package protocol

const (
    MessageTypeStream  uint8 = 0x1
    MessageTypeNormal  uint8 = 0x2
)

// Message represents a network message
type Message struct {
    Type     uint8
    From     string
    FromAddr string
    Payload  []byte
} 