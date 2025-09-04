package process

import (
	"fmt"

	"github.com/kevinjosephdavis/chatroom/client/model"
	"github.com/kevinjosephdavis/chatroom/common/message"
)

// 客户端要维护的map
var onlineUsers map[int]*message.User = make(map[int]*message.User, 1000) //后续考虑改成sync.Map
var CurUser model.CurUser                                                 //用户登录成功后完成对CurUser的初始化
//后续考虑客户端也维护一个所有用户的map

// 在客户端显示当前在线的用户
func outputOnlineUser() {
	fmt.Println("当前在线用户：")
	for _, user := range onlineUsers { //这里的user.UserName发生了丢包情况
		if user.UserID == CurUser.UserID {
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
