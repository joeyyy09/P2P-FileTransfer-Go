package protocol

const (
    MessageTypeStream  uint8 = 0x1
    MessageTypeNormal  uint8 = 0x2
    MessageTypeFileRequest uint8 = 0x3
    MessageTypeFileResponse uint8 = 0x4
    MessageTypeChunkRequest uint8 = 0x5
    MessageTypeChunkData uint8 = 0x6
)

// Message represents a network message
type Message struct {
    Type     uint8
    From     string
    FromAddr string
    Payload  interface{}
}

type FileRequest struct {
    FileName string
}

type FileResponse struct {
    Name     string
    Size     int64
    Data     []byte
    Checksum string
    NumChunks int
}

type ChunkRequest struct {
    FileName string
    ChunkNum int
}

type ChunkData struct {
    FileName  string
    ChunkNum  int
    Data      []byte
    IsLast    bool
} 