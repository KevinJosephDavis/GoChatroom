package process

import (
	"encoding/json"
	"fmt"

	"github.com/kevinjosephdavis/chatroom/client/model"
	"github.com/kevinjosephdavis/chatroom/common/message"
	"github.com/kevinjosephdavis/chatroom/server/utils"
)

type SmsProcess struct {
}

// SendGroupMes 广播：第一步 发送者发送广播内容
func (smsp *SmsProcess) SendGroupMes(content string) (err error) {
	curUser := model.GetCurUser()
	var mes message.Message
	mes.Type = message.SmsMesType

	groupMes := &message.GroupMes{
		Content: content,
		Sender: message.User{
			UserID:     curUser.UserID,
			UserStatus: curUser.UserStatus,
			UserName:   curUser.UserName,
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
		Conn: curUser.Conn,
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
	curUser := model.GetCurUser()
	var mes message.Message
	mes.Type = message.SmsPrivateMesType

	smsPrivateMes := &message.PrivateMes{
		Content: content,
		Sender: message.User{
			UserID:     model.GetCurUser().UserID,
			UserStatus: model.GetCurUser().UserStatus,
			UserName:   model.GetCurUser().UserName,
		},
		ReceiverID: receiverID,
	}

	data, err := json.Marshal(smsPrivateMes)
	if err != nil {
		fmt.Println("SendPrivateMes json.Marshal err=", err)
		return
	}

	mes.Data = string(data)

	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println("SendPrivateMes json.Marshal err=", err)
		return
	}

	tf := &utils.Transfer{
		Conn: curUser.Conn,
	}

	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("SendPrivateMes WritePkg err=", err)
		return
	}

	return
}

// SendLogoutMes 下线：第一步 客户端向服务端发送下线信息（正常退出）
func (smsp *SmsProcess) SendLogoutMes(userID int, userName string, Time int64) (err error) {
	curUser := model.GetCurUser()
	var mes message.Message
	mes.Type = message.LogoutMesType

	var logoutMes message.LogoutMes
	logoutMes.UserID = userID
	logoutMes.UserName = userName
	logoutMes.Time = Time

	data, err := json.Marshal(logoutMes)
	if err != nil {
		fmt.Println("SendLogoutMes json.Marshal err=", err)
		return
	}

	mes.Data = string(data)

	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println("SendLogoutMes json.Marshal err=", err)
		return
	}

	tf := &utils.Transfer{
		Conn: curUser.Conn,
	}
	err = tf.WritePkg(data)
	return
}

// SendDeleteAccountMes 注销：第一步 客户端向服务端发送下线信息
func (smsp *SmsProcess) SendDeleteAccountMes(userID int, userName string, Time int64) (err error) {
	curUser := model.GetCurUser()
	var mes message.Message
	mes.Type = message.DeleteAccountMesType

	var deleteAccountMes message.DeleteAccountMes
	deleteAccountMes.User.UserID = userID
	deleteAccountMes.User.UserName = userName
	deleteAccountMes.Time = Time

	data, err := json.Marshal(deleteAccountMes)
	if err != nil {
		fmt.Println("SendLogoutMes json.Marshal err=", err)
		return
	}

	mes.Data = string(data)

	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println("SendLogoutMes json.Marshal err=", err)
		return
	}

	tf := &utils.Transfer{
		Conn: curUser.Conn,
	}
	err = tf.WritePkg(data)
	return
}
