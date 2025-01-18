package peer

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"joeyyy09/P2P-FileTransfer-Go/pkg/protocol"
)

// Peer represents a node in the P2P network that can share and receive files
type Peer struct {
	id          string            // Unique identifier for the peer
	listenAddr  string            // Network address the peer listens on
	transport   Transport         // Transport layer for network communication
	sharedDir   string           // Directory for shared files
	receivedDir string           // Directory for received files
	peers       map[string]string // Map of peer IDs to their addresses
}

// Transport defines the interface for network communication
// Implementations must provide methods for connection management and message handling
type Transport interface {
	GetListenAddress() string
	ConnectToPeer(addr string) error
	StartListening() error
	GetMessageChannel() <-chan protocol.Message
	Shutdown() error
	Send(addr string, msg protocol.Message) error
}

// New creates and initializes a new Peer instance
// id: Unique identifier for the peer
// listenAddr: Network address to listen on
// sharedDir: Directory path for shared files
// receivedDir: Directory path for received files
// transport: Implementation of the Transport interface
// Returns: Initialized peer and any error encountered
func New(id, listenAddr, sharedDir, receivedDir string, transport Transport) (*Peer, error) {
	// Create both directories
	if err := os.MkdirAll(sharedDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create shared directory: %v", err)
	}
	if err := os.MkdirAll(receivedDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create received directory: %v", err)
	}

	return &Peer{
		id:          id,
		listenAddr:  listenAddr,
		transport:   transport,
		sharedDir:   sharedDir,
		receivedDir: receivedDir,
		peers:       make(map[string]string),
	}, nil
}

// Start begins peer operation by starting the transport layer and message handler
// Returns: Error if the transport fails to start
func (p *Peer) Start() error {
	if err := p.transport.StartListening(); err != nil {
		return err
	}

	go p.handleMessages()
	return nil
}

// handleMessages processes incoming messages from the transport layer
// Continuously reads from message channel and routes to appropriate handlers
func (p *Peer) handleMessages() {
	for msg := range p.transport.GetMessageChannel() {
		switch msg.Type {
		case protocol.MessageTypeFileRequest:
			p.handleFileRequest(msg)
		case protocol.MessageTypeFileResponse:
			p.handleFileResponse(msg)
		}
	}
}

// RequestFile initiates a file transfer request to a peer
// peerAddr: Address of the peer to request the file from
// fileName: Name of the file to request
// Returns: Error if the request fails to send
func (p *Peer) RequestFile(peerAddr, fileName string) error {
	maxRetries := 5
	retryInterval := time.Second * 2

	req := &protocol.FileRequest{
		FileName: fileName,
	}
	
	msg := protocol.Message{
		Type:    protocol.MessageTypeFileRequest,
		From:    p.id,
		Payload: req,
	}
	
	// Retry loop
	for i := 0; i < maxRetries; i++ {
		err := p.transport.Send(peerAddr, msg)
		if err == nil {
			return nil
		}
		
		log.Printf("Connection attempt %d failed: %v. Retrying in %v...", 
			i+1, err, retryInterval)
		
		// Wait before retrying
		time.Sleep(retryInterval)
	}
	
	return fmt.Errorf("failed to connect after %d attempts", maxRetries)
}

// handleFileRequest processes incoming file requests
// Reads the requested file and sends it back to the requesting peer
// msg: The file request message containing the file name
func (p *Peer) handleFileRequest(msg protocol.Message) {
	req := msg.Payload.(*protocol.FileRequest)
	log.Printf("Received file request from %s for file: %s", msg.From, req.FileName)
	
	filePath := filepath.Join(p.sharedDir, req.FileName)
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("File not found: %s", req.FileName)
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		log.Printf("Error reading file stats: %v", err)
		return
	}

	content := make([]byte, fileInfo.Size())
	if _, err := file.Read(content); err != nil {
		log.Printf("Error reading file: %v", err)
		return
	}
	log.Printf("Reading file: %s (size: %d bytes)", req.FileName, fileInfo.Size())

	resp := &protocol.FileResponse{
		Name: req.FileName,
		Size: fileInfo.Size(),
		Data: content,
	}

	responseMsg := protocol.Message{
		Type:     protocol.MessageTypeFileResponse,
		From:     p.id,
		FromAddr: p.listenAddr,
		Payload:  resp,
	}
	
	log.Printf("Sending file %s to peer %s", req.FileName, msg.From)
	if err := p.transport.Send(msg.FromAddr, responseMsg); err != nil {
		log.Printf("Error sending file response: %v", err)
		return
	}
	log.Printf("Successfully sent file %s to peer %s", req.FileName, msg.From)
}

// handleFileResponse processes incoming file responses
// Saves the received file to the shared directory
// msg: The file response message containing the file data
func (p *Peer) handleFileResponse(msg protocol.Message) {
	resp := msg.Payload.(*protocol.FileResponse)
	filePath := filepath.Join(p.receivedDir, resp.Name)

	if err := os.WriteFile(filePath, resp.Data, 0644); err != nil {
		log.Printf("Error saving file: %v", err)
		return
	}

	log.Printf("File received and saved: %s", filePath)
}

// Shutdown gracefully stops the peer and its transport layer
// Returns: Error if shutdown fails
func (p *Peer) Shutdown() error {
	return p.transport.Shutdown()
}

// SendFile initiates sending a file to a requesting peer
func (p *Peer) SendFile(fileName string) error {
	filePath := filepath.Join(p.sharedDir, fileName)
	
	// Verify file exists
	if _, err := os.Stat(filePath); err != nil {
		return fmt.Errorf("file %s not found: %v", fileName, err)
	}
	
	log.Printf("Ready to send file %s to any requesting peer", fileName)
	return nil
}