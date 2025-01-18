package protocol

import (
	"encoding/gob"
)

func init() {
	gob.Register(&FileRequest{})
	gob.Register(&FileResponse{})
	gob.Register([]byte{})
} 