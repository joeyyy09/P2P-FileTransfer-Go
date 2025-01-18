package main

import (
	"flag"
	"log"
	"path/filepath"

	"joeyyy09/P2P-FileTransfer-Go/peer"
	"joeyyy09/P2P-FileTransfer-Go/pkg"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	
	// Basic peer setup flags
	peerID := flag.String("id", "", "Peer ID (peer1 or peer2)")
	port := flag.String("port", "", "Port to listen on (3000 or 3001)")
	
	// File operation flags
	sendFile := flag.String("send", "", "Path to file to send (relative to shared directory)")
	receiveFile := flag.String("receive", "", "Name of file to receive")
	targetPeer := flag.String("peer", "", "Address of peer to connect to (e.g., localhost:3000)")
	
	// Directory flags
	sharedDir := flag.String("shared", "", "Directory for shared files (default: ./shared{id})")
	receivedDir := flag.String("received", "", "Directory for received files (default: ./received{id})")
	
	flag.Parse()

	if *peerID == "" || *port == "" {
		log.Fatal("Please provide -id and -port flags")
	}

	// Set default directories if not specified
	if *sharedDir == "" {
		*sharedDir = filepath.Join(".", "shared"+(*peerID)[4:])
	}
	if *receivedDir == "" {
		*receivedDir = filepath.Join(".", "received"+(*peerID)[4:])
	}

	// Create and start peer
	transport := pkg.NewTCPTransport("localhost:" + *port)
	p, err := peer.New(*peerID, "localhost:"+*port, *sharedDir, *receivedDir, transport)
	if err != nil {
		log.Fatal(err)
	}

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}

	// Handle file operations
	if *receiveFile != "" {
		if *targetPeer == "" {
			log.Fatal("Please specify peer address with -peer flag")
		}
		if err := p.RequestFile(*targetPeer, *receiveFile); err != nil {
			log.Printf("File receive error: %v", err)
		}
	} else if *sendFile != "" {
		if err := p.SendFile(*sendFile); err != nil {
			log.Printf("File send error: %v", err)
		} else {
			log.Printf("Ready to send file: %s", *sendFile)
		}
	} else {
		log.Printf("Peer %s listening on %s", *peerID, *port)
		log.Printf("Shared directory: %s", *sharedDir)
		log.Printf("Received files directory: %s", *receivedDir)
	}

	// Keep program running
	select {}
} 