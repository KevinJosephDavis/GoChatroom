package process

import (
	"encoding/json"
	"fmt"
	"time"

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

// outputOfflineMes 下线：第三步 客户端接收服务端返回的信息，呈现下线的用户ID和昵称
func outputOfflineMes(mes *message.Message) {
	//注意，不管用户是正常下线，还是非正常下线，第三步都调用这个函数（暂时这么考虑）
	var offlineResMes message.OfflineResMes
	err := json.Unmarshal([]byte(mes.Data), &offlineResMes)
	if err != nil {
		fmt.Println("outputOfflineMes json.Unmarshal err=", err)
		return
	}
	delete(onlineUsers, offlineResMes.UserID) //将该用户从客户端维护的在线用户map中删除
	info := fmt.Sprintf("%s (ID:%d) 于 %s 下线，下线原因是：%s",
		offlineResMes.UserName,
		offlineResMes.UserID,
		getOfflineTime(offlineResMes.Time),
		getOfflineReason(offlineResMes.Reason))
	fmt.Println(info)
	fmt.Println()
}

// getOfflineReason 下线：获取下线原因
func getOfflineReason(Reason string) string {
	switch Reason {
	case message.Normal:
		return "正常下线"
	case message.Abnormal:
		return "非正常下线"
	default:
		return "未知下线原因"
	}
}

// getOfflineTime 下线：获取下线时间
func getOfflineTime(timeStamp int64) string {
	//将Unix时间戳转换为time.Time
	t := time.Unix(timeStamp, 0)

	//格式化
	return t.Format("2006-01-02 15:04:05") //Go的特殊格式
}
