package pkg

import (
	"io"
	"log"
	"net"
	"sync"

	"joeyyy09/P2P-FileTransfer-Go/pkg/protocol"
)

// TCPTransport implements the Transport interface using TCP protocol
// It manages peer connections and message routing in a P2P network
type TCPTransport struct {
	listenAddr string          // Address to listen for incoming connections
	listener   net.Listener    // TCP listener instance
	messageCh  chan protocol.Message    // Channel for incoming messages
	mu         sync.RWMutex    // Mutex for thread-safe operations
	peers      map[net.Addr]net.Conn // Active peer connections
	decoder    protocol.Decoder
	encoder    protocol.Encoder
}

// NewTCPTransport creates and initializes a new TCPTransport instance
// listenAddr: The address to listen for incoming connections (e.g., "localhost:3000")
// Returns: A configured TCPTransport instance
func NewTCPTransport(listenAddr string) *TCPTransport {
	return &TCPTransport{
		listenAddr: listenAddr,
		messageCh:  make(chan protocol.Message, 1024),
		peers:      make(map[net.Addr]net.Conn),
		decoder:    protocol.NewGobDecoder(),
		encoder:    protocol.NewGobEncoder(),
	}
}

// GetListenAddress returns the address this transport is listening on
func (t *TCPTransport) GetListenAddress() string {
	return t.listenAddr
}

// StartListening initializes the TCP listener and starts accepting connections
// Returns an error if the listener cannot be started
func (t *TCPTransport) StartListening() error {
	ln, err := net.Listen("tcp", t.listenAddr)
	if err != nil {
		return err
	}
	t.listener = ln
	
	go t.handleIncomingConnections()
	return nil
}

// handleIncomingConnections continuously accepts new TCP connections
// and spawns goroutines to handle each connection
func (t *TCPTransport) handleIncomingConnections() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			log.Printf("Connection accept error: %v", err)
			continue
		}
		
		go t.managePeerConnection(conn)
	}
}

// managePeerConnection handles an individual peer connection
// It reads messages from the connection and forwards them to the message channel
func (t *TCPTransport) managePeerConnection(conn net.Conn) {
	defer conn.Close()
	
	t.mu.Lock()
	t.peers[conn.RemoteAddr()] = conn
	t.mu.Unlock()

	defer func() {
		t.mu.Lock()
		delete(t.peers, conn.RemoteAddr())
		t.mu.Unlock()
	}()

	msg := &protocol.Message{}
	for {
		err := t.decoder.Decode(conn, msg)
		if err != nil {
			if err != io.EOF {
				log.Printf("Decode error: %v", err)
			}
			return
		}

		msg.FromAddr = conn.RemoteAddr().String()
		t.messageCh <- *msg
	}
}

// ConnectToPeer establishes a connection to a remote peer
// addr: The address of the remote peer to connect to
// Returns an error if the connection fails
func (t *TCPTransport) ConnectToPeer(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	t.mu.Lock()
	t.peers[conn.RemoteAddr()] = conn
	t.mu.Unlock()

	go t.managePeerConnection(conn)
	return nil
}

// GetMessageChannel returns a receive-only channel for consuming messages
func (t *TCPTransport) GetMessageChannel() <-chan protocol.Message {
	return t.messageCh
}

// Shutdown gracefully closes all connections and resources
func (t *TCPTransport) Shutdown() error {
	if t.listener != nil {
		t.listener.Close()
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	
	for _, conn := range t.peers {
		conn.Close()
	}
	
	close(t.messageCh)
	return nil
}

// TCPPeer represents a connected peer in the network
type TCPPeer struct {
	conn    net.Conn
	encoder protocol.Encoder
}

// NewTCPPeer creates a new TCPPeer instance
func NewTCPPeer(conn net.Conn) *TCPPeer {
	return &TCPPeer{
		conn:    conn,
		encoder: protocol.NewGobEncoder(),
	}
}

// SendMessage sends data to the peer
// payload: The data to send
// Returns an error if the send operation fails
func (p *TCPPeer) SendMessage(payload []byte, msgType uint8) error {
	msg := &protocol.Message{
		Type:    msgType,
		Payload: payload,
	}
	return p.encoder.Encode(p.conn, msg)
}

func (p *TCPPeer) Send(payload []byte) error {
	msg := &protocol.Message{
		Type:    protocol.MessageTypeNormal,
		Payload: payload,
	}
	return p.encoder.Encode(p.conn, msg)
}

