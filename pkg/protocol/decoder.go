 package protocol

import (
    "encoding/gob"
    "io"
)

// Decoder interface for decoding messages from network streams
type Decoder interface {
    Decode(io.Reader, *Message) error
}

// GobDecoder implements Decoder using Go's Gob encoding
type GobDecoder struct{}

func NewGobDecoder() *GobDecoder {
    // Register Message type with gob
    gob.Register(&Message{})
    return &GobDecoder{}
}

func (dec *GobDecoder) Decode(r io.Reader, msg *Message) error {
    decoder := gob.NewDecoder(r)
    return decoder.Decode(msg)
}