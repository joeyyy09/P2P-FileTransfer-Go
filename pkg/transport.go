package pkg

import (
	"joeyyy09/P2P-FileTransfer-Go/pkg/protocol"
	"net"
)

type Peer interface {
    net.Conn
    Send([]byte) error
}

type Transport interface {
    GetListenAddress() string
    ConnectToPeer(addr string) error
    StartListening() error
    GetMessageChannel() <-chan protocol.Message
    Shutdown() error
}