package raknet

import (
	"context"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// 描述一个简单的，
// 但不基于 Minecraft
// 通信协议的基本 Raknet 连接实例
type Raknet struct {
	connection *websocket.Conn

	context context.Context
	cancel  context.CancelFunc

	closed     bool
	closedLock sync.Mutex

	decodePacket func(buf []byte) (pk Packet)
	encodePacket func(pk Packet) (buf []byte)

	packet chan (Packet)
}

// 描述 Raknet 数据包
type Packet packet.Packet
