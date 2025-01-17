package protocol

import (
	"encoding/gob"
	"io"
)

// Encoder interface for encoding messages
type Encoder interface {
	Encode(io.Writer, *Message) error
}

// GobEncoder implements Encoder using Go's Gob encoding
type GobEncoder struct{}

func NewGobEncoder() *GobEncoder {
    // Register Message type with gob
    gob.Register(&Message{})
    return &GobEncoder{}
}

func (enc *GobEncoder) Encode(w io.Writer, msg *Message) error {
    encoder := gob.NewEncoder(w)
    return encoder.Encode(msg)
}