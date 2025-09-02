package process

import (
	"fmt"

	"github.com/kevinjosephdavis/chatroom/client/model"
	"github.com/kevinjosephdavis/chatroom/common/message"
)

// 客户端要维护的map
var onlineUsers map[int]*message.User = make(map[int]*message.User, 1000)
var CurUser model.CurUser //用户登录成功后完成对CurUser的初始化

// 在客户端显示当前在线的用户
func outputOnlineUser() {
	fmt.Println("当前在线用户：")
	for _, user := range onlineUsers {
		if user.UserID == CurUser.UserID {
			continue
		}
		fmt.Printf("%s (ID:%d):", user.UserName, user.UserID)
	}
}

// 编写一个方法处理返回的信息
func updateUserStatus(notifyUserStatusMes *message.NotifyUserStatusMes) {

	user, ok := onlineUsers[notifyUserStatusMes.UserID]
	if !ok {
		//原来没有
		user := &message.User{
			UserID:     notifyUserStatusMes.UserID,
			UserStatus: notifyUserStatusMes.Status,
		}

		user.UserStatus = notifyUserStatusMes.Status
		onlineUsers[notifyUserStatusMes.UserID] = user
		outputOnlineUser()
		return
	}
	// 如果用户已存在，更新其状态
	user.UserStatus = notifyUserStatusMes.Status
	outputOnlineUser()
}
