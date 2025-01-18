 package transport

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"joeyyy09/P2P-FileTransfer-Go/pkg/protocol"
)

// TCPTransport implements the Transport interface using TCP protocol
// It manages peer connections and message routing in a P2P network
type TCPTransport struct {
	listenAddr string          // Address to listen for incoming connections
	listener   net.Listener    // TCP listener instance
	messageCh  chan protocol.Message    // Channel for incoming messages
	mu         sync.RWMutex    // Mutex for thread-safe operations
	peers      map[string]net.Conn // Active peer connections
	decoder    protocol.Decoder
}

// NewTCPTransport creates and initializes a new TCPTransport instance
// listenAddr: The address to listen for incoming connections (e.g., "localhost:3000")
// Returns: A configured TCPTransport instance
func NewTCPTransport(listenAddr string) *TCPTransport {
	return &TCPTransport{
		listenAddr: listenAddr,
		messageCh:  make(chan protocol.Message, 1024),
		peers:      make(map[string]net.Conn),
		decoder:    protocol.NewGobDecoder(),
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
	
	log.Printf("New peer connection established from %s", conn.RemoteAddr())
	
	t.mu.Lock()
	t.peers[conn.RemoteAddr().String()] = conn
	t.mu.Unlock()

	defer func() {
		t.mu.Lock()
		delete(t.peers, conn.RemoteAddr().String())
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
	log.Printf("Connecting to peer at %s", addr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("dial failed: %v", err)
	}

	t.mu.Lock()
	t.peers[addr] = conn
	t.mu.Unlock()

	log.Printf("Connected to peer at %s", addr)
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

// Add this method to TCPTransport
func (t *TCPTransport) Send(addr string, msg protocol.Message) error {
	t.mu.Lock()
	conn, exists := t.peers[addr]
	t.mu.Unlock()

	if !exists {
		// Connect first
		if err := t.ConnectToPeer(addr); err != nil {
			return fmt.Errorf("failed to connect to peer %s: %v", addr, err)
		}
		// Wait a bit for connection to be established
		time.Sleep(100 * time.Millisecond)
		
		// Get the connection
		t.mu.Lock()
		conn = t.peers[addr]
		t.mu.Unlock()
		
		if conn == nil {
			return fmt.Errorf("failed to establish connection with %s", addr)
		}
	}

	encoder := protocol.NewGobEncoder()
	return encoder.Encode(conn, &msg)
}

