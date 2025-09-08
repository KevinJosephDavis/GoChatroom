// Package model provides data structures for the chatroom client.
package model

import (
	"net"
	"sync"

	"github.com/kevinjosephdavis/chatroom/common/message"
)

// CurUser 在客户端很多地方会使用到CurUser，因此将其作为全局的变量
type CurUser struct {
	Conn net.Conn
	message.User
}

var (
	curUserInstance *CurUser
	mu              sync.RWMutex
)

// GetCurUser 获取当前用户实例
func GetCurUser() *CurUser {
	mu.RLock()
	defer mu.RUnlock()
	return curUserInstance
}

// SetCurUser 设置当前用户 登录成功后调用
func SetCurUser(conn net.Conn, user message.User) {
	mu.Lock()
	defer mu.Unlock()
	curUserInstance = &CurUser{
		Conn: conn,
		User: user,
	}
}

// ClearCurUser 清理当前用户 退出后调用
func ClearCurUser() {
	mu.Lock()
	defer mu.Unlock()
	curUserInstance = nil
}
