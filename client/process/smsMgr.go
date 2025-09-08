package process

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/kevinjosephdavis/chatroom/client/model"
	"github.com/kevinjosephdavis/chatroom/common/message"
)

// outputGroupMes 广播：第三步 客户端接收服务端返回的信息，并呈现发送方的ID及接收到的信息
func outputGroupMes(mes *message.Message) {
	var groupMes message.GroupMes
	err := json.Unmarshal([]byte(mes.Data), &groupMes)
	if err != nil {
		fmt.Println("json.Unmarshal err=", err)
		return
	}

	info := fmt.Sprintf("%s (ID:%d)\t 对大家说：\t%s", groupMes.Sender.UserName, groupMes.Sender.UserID, groupMes.Content)
	fmt.Println(info)
	fmt.Println()
}

// outputPrivateMes 私聊：第三步 客户端接收服务端返回的信息，并呈现发送方的ID以及接收到的信息
func outputPrivateMes(mes *message.Message) {
	var privateMsg message.PrivateResMes
	err := json.Unmarshal([]byte(mes.Data), &privateMsg)
	if err != nil {
		fmt.Println("outputPrivateMes json.Unmarshal err=", err)
		return
	}
	info := fmt.Sprintf("%s (ID:%d)\t 对你说：\t%s", privateMsg.Sender.UserName, privateMsg.Sender.UserID, privateMsg.Content)
	fmt.Println(info)
	fmt.Println()
}

// outputLogoutMes 下线：第三步 客户端接收服务端返回的信息，呈现下线的用户ID、昵称和下线时间
func outputLogoutMes(mes *message.Message) {
	//注意，不管用户是正常下线，还是非正常下线，第三步都调用这个函数（暂时这么考虑）
	var logoutResMes message.LogoutResMes
	err := json.Unmarshal([]byte(mes.Data), &logoutResMes)
	if err != nil {
		fmt.Println("outputLogoutMes json.Unmarshal err=", err)
		return
	}
	delete(onlineUsers, logoutResMes.UserID) //将该用户从客户端维护的在线用户map中删除
	info := fmt.Sprintf("%s (ID:%d) 于 %s 下线，下线原因是：%s",
		logoutResMes.UserName,
		logoutResMes.UserID,
		getTime(logoutResMes.Time),
		getReason(logoutResMes.Reason))
	fmt.Println(info)
	fmt.Println()
}

// outputDeleteAccountMes 注销：第三步 客户端接收服务端返回的信息，呈现注销的用户ID、昵称和下线时间
func outputDeleteAccountMes(mes *message.Message) {
	var DeleteAccountResMes message.DeleteAccountResMes
	err := json.Unmarshal([]byte(mes.Data), &DeleteAccountResMes)
	if err != nil {
		fmt.Println("outputOfflineMes json.Unmarshal err=", err)
		return
	}

	//判断是否是自己的注销操作
	if DeleteAccountResMes.User.UserID == CurUser.UserID {
		fmt.Printf("您的账号 %s (ID:%d) 已成功注销\n",
			DeleteAccountResMes.User.UserName,
			DeleteAccountResMes.User.UserID)

		//先发送退出信号，让主循环先退出
		select {
		case DeleteAccountChan <- true:
			fmt.Println("已发送注销退出信号")
		default:
			fmt.Println("注销通道已满，尝试其它方式")
			select {
			case exitChan <- true:
			default:
			}
		}

		//清理资源
		go func() {
			if cancelFunc != nil {
				cancelFunc() //通知消息协程退出
			}
			if CurUser.Conn != nil {
				CurUser.Conn.Close() //关闭连接
			}
			//由于是本用户注销，因此要清理其客户端状态
			onlineUsers = make(map[int]*message.User)
			CurUser = model.CurUser{}
			fmt.Println("资源清理完成")
		}()
	} else {
		//这是其他用户注销的通知
		delete(onlineUsers, DeleteAccountResMes.User.UserID) //将该用户从客户端维护的在线用户map中删除
		info := fmt.Sprintf("%s (ID:%d) 于 %s 注销了用户",
			DeleteAccountResMes.User.UserName,
			DeleteAccountResMes.User.UserID,
			getTime(DeleteAccountResMes.Time))
		fmt.Println(info)
	}
	fmt.Println()
}

// getReason 下线：获取下线原因
func getReason(Reason string) string {
	switch Reason {
	case message.Normal:
		return "正常下线"
	case message.Abnormal:
		return "非正常下线"
	default:
		return "未知下线原因"
	}
}

// getTime 下线/注销：获取下线/注销时间
func getTime(timeStamp int64) string {
	//将Unix时间戳转换为time.Time
	t := time.Unix(timeStamp, 0)

	//格式化
	return t.Format("2006-01-02 15:04:05") //Go的特殊格式
}

// outputErrorRes 打印离线留言返回的错误信息
func outputErrorRes(mes *message.Message) {
	var errorRes map[string]interface{}
	err := json.Unmarshal([]byte(mes.Data), &errorRes)
	if err != nil {
		fmt.Println("outputErrorRes json.Unmarshal err=", err)
		return
	}

	if code, ok := errorRes["code"].(string); ok {
		if message, ok := errorRes["message"].(string); ok {
			switch code {
			case "UserNotExist":
				fmt.Printf("%s\n", message)
			case "OfflineMesStored":
				fmt.Printf("%s\n", message)
			default:
				fmt.Printf("系统错误：%s （错误码：%s）\n", message, code)
			}
		} else {
			fmt.Println("错误消息格式不正确")
		}
	} else {
		fmt.Println("错误码格式不正确")
	}
}

// outputOfflineMes 打印离线留言消息
func outputOfflineResMes(mes *message.Message) {
	var offlineResMes message.OfflineResMes
	err := json.Unmarshal([]byte(mes.Data), &offlineResMes)
	if err != nil {
		fmt.Println("outputOfflineResMes json.Unmarshal err=", err)
		return
	}
	mesTime := time.Unix(offlineResMes.Time, 0)
	info := fmt.Sprintf("用户%s (ID:%d) 于%s给您的离线留言：%s", offlineResMes.SenderName, offlineResMes.SenderID,
		mesTime.Format("2006-01-02 15:04:05"), offlineResMes.Content)
	fmt.Println(info)
	fmt.Println()
}
