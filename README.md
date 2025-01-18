# P2P File Transfer

A peer-to-peer file transfer system implemented in Go.

# Initial implementation of P2P file transfer system

## Features:
- Basic peer-to-peer file transfer functionality
- Separate directories for shared and received files
- Configurable shared and received directories
- Retry mechanism for connection attempts
- Clear command-line interface
- TCP transport layer with connection management
- Detailed logging for operations

## Usage examples:
1. Start a peer in listening mode:
   go run main.go -id peer1 -port 3000

2. Send a file:
   go run main.go -id peer2 -port 3001 -send test.txt

3. Receive a file:
   go run main.go -id peer1 -port 3000 -receive test.txt -peer localhost:3001
