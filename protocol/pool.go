package protocol

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

// Pool ..
func Pool() map[uint32]packet.Packet {
	return map[uint32]packet.Packet{
		1: &FileSeekRequest{},
		2: &FileSeekResponse{},
		3: &FileChunkRequest{},
		4: &FileChunkResponse{},
		5: &FileSizeRequest{},
		6: &FileSizeResponse{},
	}
}
