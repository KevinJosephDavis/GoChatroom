// Package model provides data structures for the chatroom client.
package model

import (
	"net"

	"github.com/kevinjosephdavis/chatroom/common/message"
)

// CurUser 在客户端很多地方会使用到CurUser，因此将其作为全局的变量
type CurUser struct {
	Conn net.Conn
	message.User
}
