package raknet

import (
	"bytes"
	"fmt"
	"runtime/debug"

	"github.com/pterm/pterm"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	customProtocol "simple_downloader/protocol"
)

// ...
func DecodePacket(buf []byte) (pk Packet) {
	buffer := bytes.NewBuffer(buf)
	reader := protocol.NewReader(buffer, 0, false)

	header := packet.Header{}
	header.Read(buffer)

	defer func() {
		r := recover()
		if r == nil {
			return
		}
		if pk == nil {
			pterm.Warning.Printf(
				"DecodePacket: Failed to unmarshal packet which numbered %d, and the error log is %v\n\n[Stack Info]\n%s\n",
				header.PacketID, r, string(debug.Stack()),
			)
			fmt.Println()
		} else {
			pterm.Warning.Printf(
				"DecodePacket: Failed to unmarshal packet %T, and the error log is %v\n\n[Stack Info]\n%s\n",
				pk, r, string(debug.Stack()),
			)
			fmt.Println()
		}
	}()

	pk = customProtocol.Pool()[header.PacketID]
	pk.Marshal(reader)
	return pk
}

// ...
func EncodePacket(pk Packet) []byte {
	defer func() {
		r := recover()
		if r != nil {
			pterm.Warning.Printf(
				"EncodePacket: Failed to marshal packet %T, and the error log is %v\n\n[Stack Info]\n%s\n",
				pk, r, string(debug.Stack()),
			)
			fmt.Println()
		}
	}()

	buffer := bytes.NewBuffer([]byte{})
	writer := protocol.NewWriter(buffer, 0)

	header := packet.Header{PacketID: pk.ID()}
	header.Write(buffer)

	pk.Marshal(writer)
	return buffer.Bytes()
}
