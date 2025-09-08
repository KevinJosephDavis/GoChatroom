package process

import (
	"context"
	"fmt"

	"github.com/kevinjosephdavis/chatroom/client/model"
	"github.com/kevinjosephdavis/chatroom/common/message"
)

var (
	//CurUser           model.CurUser
	onlineUsers       = make(map[int]*message.User)
	cancelFunc        context.CancelFunc
	exitChan          = make(chan bool, 1)
	DeleteAccountChan = make(chan bool, 1)
)

// 在客户端显示当前在线的用户
func outputOnlineUser() {
	fmt.Println("当前在线用户：")
	for _, user := range onlineUsers {
		if user.UserID == model.GetCurUser().UserID {
			continue
		}
		fmt.Printf("%s (ID:%d):", user.UserName, user.UserID)
		fmt.Println()
	}
}

// updateUserStatus 上线：第三步 处理服务端返回的信息并输出
func updateUserStatus(notifyUserStatusMes *message.NotifyUserStatusMes) {

	user, ok := onlineUsers[notifyUserStatusMes.UserID]
	if !ok {
		//原来没有
		user := &message.User{
			UserID:     notifyUserStatusMes.UserID,
			UserName:   notifyUserStatusMes.UserName,
			UserStatus: notifyUserStatusMes.Status,
		}
		onlineUsers[notifyUserStatusMes.UserID] = user
	} else {
		user.UserStatus = notifyUserStatusMes.Status
		user.UserName = notifyUserStatusMes.UserName
	}
	outputOnlineUser()
}
