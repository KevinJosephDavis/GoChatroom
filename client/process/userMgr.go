package process

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kevinjosephdavis/chatroom/client/model"
	"github.com/kevinjosephdavis/chatroom/common/message"
	"github.com/kevinjosephdavis/chatroom/server/utils"
)

var (
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
		fmt.Printf("%s (ID:%d) ", user.UserName, user.UserID)
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

// StartHeartBeatSending 开始发送心跳
func StartHeartBeatSending() {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			curUser := model.GetCurUser()
			if curUser == nil || curUser.Conn == nil {
				//用户已下线
				continue
			}

			//尝试发送心跳
			heartBeatMsg := map[string]int{
				"userID": curUser.UserID,
			}

			data, err := json.Marshal(heartBeatMsg)
			if err != nil {
				fmt.Println("startHeartBeatSending json.Marshal err=", err)
				return
			}
			mes := message.Message{
				Type: message.HeartBeatType,
				Data: string(data),
			}

			FinalData, err := json.Marshal(mes)
			if err != nil {
				fmt.Println("startHeartBeatSending json.Marshal err=", err)
				return
			}
			tf := &utils.Transfer{
				Conn: curUser.Conn,
			}
			tf.WritePkg(FinalData) //发送心跳，不处理错误
		}
	}()
}
