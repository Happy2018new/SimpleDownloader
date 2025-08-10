package protocol

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// FileSeekRequest ..
type FileSeekRequest struct {
	Offset int64
}

func (*FileSeekRequest) ID() uint32 {
	return 1
}

func (f *FileSeekRequest) Marshal(io protocol.IO) {
	io.Varint64(&f.Offset)
}

// FileSeekResponse ..
type FileSeekResponse struct {
	Success   bool
	ErrorInfo string
}

func (*FileSeekResponse) ID() uint32 {
	return 2
}

func (f *FileSeekResponse) Marshal(io protocol.IO) {
	io.Bool(&f.Success)
	io.String(&f.ErrorInfo)
}
