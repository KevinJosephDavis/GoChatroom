package process

import (
	"encoding/json"
	"fmt"

	"github.com/kevinjosephdavis/chatroom/common/message"
	"github.com/kevinjosephdavis/chatroom/server/utils"
)

type SmsProcess struct {
}

// SendGroupMes 广播：第一步 发送者发送广播内容
func (smsp *SmsProcess) SendGroupMes(content string) (err error) {
	var mes message.Message
	mes.Type = message.SmsMesType

	groupMes := &message.GroupMes{
		Content: content,
		Sender: message.User{
			UserID:     CurUser.UserID,
			UserStatus: CurUser.UserStatus,
			UserName:   CurUser.UserName,
		},
	}
	data, err := json.Marshal(groupMes)
	if err != nil {
		fmt.Println("SendGroupMes json.Marshal err=", err)
		return
	}

	mes.Data = string(data)

	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println("SendGroupMes json.Marshal err=", err)
		return
	}

	tf := &utils.Transfer{
		Conn: CurUser.Conn,
	}

	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("SendGroupMes WritePkg err=", err)
		return
	}

	return
}

// SendPrivateMes 私聊：第一步 发送者发送私聊消息
func (smsp *SmsProcess) SendPrivateMes(content string, receiverID int) (err error) {
	var mes message.Message
	mes.Type = message.SmsPrivateMesType

	smsPrivateMes := &message.PrivateMes{
		Content: content,
		Sender: message.User{
			UserID:     CurUser.UserID,
			UserStatus: CurUser.UserStatus,
			UserName:   CurUser.UserName,
		},
		ReceiverID: receiverID,
	}

	data, err := json.Marshal(smsPrivateMes)
	if err != nil {
		fmt.Println("SendGroupMes json.Marshal err=", err)
		return
	}

	mes.Data = string(data)

	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println("SendGroupMes json.Marshal err=", err)
		return
	}

	tf := &utils.Transfer{
		Conn: CurUser.Conn,
	}

	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("SendGroupMes WritePkg err=", err)
		return
	}

	return
}

// SendOfflineMes 下线：第一步 客户端向服务端发送下线信息（正常退出）
func (smsp *SmsProcess) SendOfflineMes(userID int, userName string, Time int64) (err error) {
	var mes message.Message
	mes.Type = message.OfflineMesType

	var offlineMes message.OfflineMes
	offlineMes.UserID = userID
	offlineMes.UserName = userName
	offlineMes.Time = Time
	offlineMes.Reason = message.Normal //能够发送offlineMes，属于正常退出
	//正常退出的几种情况：1.用户选择5 退出系统 2.用户输入exit 或 ctrl+C
	//后续要在上层写 收集用户键入信息的函数
	//调用完这个SendOfflineMes函数后，要关闭连接

	data, err := json.Marshal(offlineMes)
	if err != nil {
		fmt.Println("SendOfflineMes json.Marshal err=", err)
		return
	}

	mes.Data = string(data)

	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println("SendOfflineMes json.Marshal err=", err)
		return
	}

	tf := &utils.Transfer{
		Conn: CurUser.Conn,
	}
	err = tf.WritePkg(data)
	return
}
