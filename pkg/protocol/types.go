package protocol

const (
    MessageTypeFileRequest uint8 = 0x3
    MessageTypeFileResponse uint8 = 0x4
)

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
    Name string
    Size int64
    Data []byte
} 