package protocol

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// FileSizeRequest ..
type FileSizeRequest struct{}

func (*FileSizeRequest) ID() uint32 {
	return 5
}

func (f *FileSizeRequest) Marshal(io protocol.IO) {}

// FileSizeResponse ..
type FileSizeResponse struct {
	Size int64
}

func (*FileSizeResponse) ID() uint32 {
	return 6
}

func (f *FileSizeResponse) Marshal(io protocol.IO) {
	io.Varint64(&f.Size)
}
