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
