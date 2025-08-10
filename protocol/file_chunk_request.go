package protocol

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// FileChunkRequest ..
type FileChunkRequest struct{}

func (*FileChunkRequest) ID() uint32 {
	return 3
}

func (f *FileChunkRequest) Marshal(io protocol.IO) {}

// FileChunkResponse ..
type FileChunkResponse struct {
	IsFinalChunk bool
	ChunkData    []byte
}

func (*FileChunkResponse) ID() uint32 {
	return 4
}

func (f *FileChunkResponse) Marshal(io protocol.IO) {
	io.Bool(&f.IsFinalChunk)
	io.Bytes(&f.ChunkData)
}
