package main

import (
	"context"
	"fmt"

	pb "orcanet/market"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/record"
	"github.com/multiformats/go-multiaddr"
)

// printRoutingTable prints the current state of the DHT's routing table to the console.
// This function is useful for debugging and monitoring the local view of the network topology.
//
// Parameters:
// - dht: A pointer to the dht.IpfsDHT instance whose routing table is to be printed.
//
// Returns: None.
func printRoutingTable(dht *dht.IpfsDHT) {
	for _, peer := range dht.RoutingTable().ListPeers() {
		fmt.Println("Peer ID:", peer)
	}
}

// registerFile registers a file in the DHT, indicating that the local user holds a specific file.
// This operation makes the file discoverable to other peers searching for it through the DHT.
//
// Parameters:
// - ctx: A context.Context for controlling the function's execution lifetime.
// - dht: A pointer to the dht.IpfsDHT used for the registration.
// - req: A *pb.RegisterFileRequest containing the user's information and the file hash to register.
//
// Returns: An error if the registration fails, or nil on success.
func registerFile(ctx context.Context, dht *dht.IpfsDHT, fileHash string, envelope *record.Envelope) error {
	key := fmt.Sprintf("/market/file/%s", fileHash)

	// Serialize the envelope containing the PeerRecord
	data, err := envelope.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal PeerRecord envelope: %v", err)
	}

	// Store the serialized data in the DHT
	if err := dht.PutValue(ctx, key, data); err != nil {
		return fmt.Errorf("failed to put value in the DHT for file hash %s: %v", fileHash, err)
	}

	fmt.Printf("Successfully registered file with hash %s\n", fileHash)
	return nil
}

// checkHolders retrieves a list of users holding a specific file by querying the DHT.
// This function is part of the file discovery process, allowing peers to locate others
// that have the file they are looking for.
//
// Parameters:
// - ctx: A context.Context for controlling the function's execution lifetime.
// - dht: A pointer to the dht.IpfsDHT used for the query.
// - req: A *pb.CheckHoldersRequest containing the file hash to search for.
//
// Returns: A *pb.HoldersResponse containing the list of Users holding the file, and an error if the query fails.
func checkHolders(ctx context.Context, dht *dht.IpfsDHT, req *pb.CheckHoldersRequest) (*pb.HoldersResponse, error) {
	key := fmt.Sprintf("/market/file/%s", req.FileHash)

	// Retrieve the serialized envelope from the DHT
	envelopeBytes, err := dht.GetValue(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("error searching for file %s: %w", req.FileHash, err)
	}

	// Deserialize the envelope
	envelope, err := record.UnmarshalEnvelope(envelopeBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal envelope: %w", err)
	}

	// Assuming the envelope payload is a PeerRecord, we now extract it
	// Note: This step may vary depending on the actual payload type.
	var peerRec peer.PeerRecord
	if err := peerRec.UnmarshalRecord(envelope.RawPayload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal peer record from envelope payload: %w", err)
	}

	// Convert the PeerRecord to the protobuf response format
	// This example is simplified; adjust based on your actual protobuf structure
	holders := []*pb.PeerInfo{
		// Populate the PeerInfo based on peerRec
		{PeerId: peerRec.PeerID.String()},
		// Include multiaddresses if your pb.PeerInfo structure supports them
	}

	return &pb.HoldersResponse{Holders: holders}, nil
}

// Helper function to convert multiaddresses to strings
func convertAddrsToStrings(addrs []multiaddr.Multiaddr) []string {
	addrStrs := make([]string, len(addrs))
	for i, addr := range addrs {
		addrStrs[i] = addr.String()
	}
	return addrStrs
}
