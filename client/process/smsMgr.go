package process

import (
	"encoding/json"
	"fmt"

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
	info := fmt.Sprintf("%s (ID:%d) \t 下线了", offlineResMes.UserName, offlineResMes.UserID)
	fmt.Println(info)
	fmt.Println()
}
