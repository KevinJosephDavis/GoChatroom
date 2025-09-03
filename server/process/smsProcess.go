// Package process2 处理和短消息相关的请求。群聊、点对点聊天。
package process2

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/kevinjosephdavis/chatroom/common/message"
	"github.com/kevinjosephdavis/chatroom/server/utils"
)

type SmsProcess struct {
}

// SendGroupMes 广播：第二步 服务端接收发送方的信息，遍历map将信息发出
func (smsp *SmsProcess) SendGroupMes(mes *message.Message) {

	var groupMes message.GroupMes
	err := json.Unmarshal([]byte(mes.Data), &groupMes)
	if err != nil {
		fmt.Println("SendGroupMes json.Unmarshal err=", err)
		return
	}

	data, err := json.Marshal(mes)
	if err != nil {
		fmt.Println("json.Marshal err=", err)
		return
	}
	// 反序列化 mes.Data 主要是为了获取发送者的 UserID，避免给自己发消息
	for id, up := range userMgr.onlineUsers {
		//过滤掉自己
		if id == groupMes.Sender.UserID {
			continue
		}

		smsp.SendMesToEachOnlineUser(data, up.Conn) //获取每个在线用户的控制器，得到其与服务端的连接，进而发送信息
	}
}

// SendMesToEachOnlineUser 广播：给每个用户发消息
func (smsp *SmsProcess) SendMesToEachOnlineUser(data []byte, conn net.Conn) {
	tf := &utils.Transfer{
		Conn: conn,
	}
	err := tf.WritePkg(data)
	if err != nil {
		fmt.Println("转发消息失败，err=", err)
		return
	}
}

// SendMesToSpecifiedUser 私聊：给指定用户发消息
func (smsp *SmsProcess) SendMesToSpecifiedUser(data []byte, conn net.Conn) {
	tf := &utils.Transfer{
		Conn: conn,
	}
	err := tf.WritePkg(data)
	if err != nil {
		fmt.Println("私聊消息发送失败，err=", err)
		return
	}
}

// SendPrivateMes 私聊 第二步：服务器端接收发送方的信息，遍历map确定接收方后将其打包发给接收方
func (smsp *SmsProcess) SendPrivateMes(mes *message.Message) {
	//在服务器端的onlineUsers map[int]*UserProcess0 中找到对应的userID并发送消息
	var smsPrivateMes message.PrivateMes
	err := json.Unmarshal([]byte(mes.Data), &smsPrivateMes) //反序列化的目的：获取发送者的ID和接收者的ID
	if err != nil {
		fmt.Println("SendPrivateMes json.Unmarshal err=", err)
		return
	}

	mes.Type = message.SmsPrivateResMesType
	data, err := json.Marshal(mes) //序列化mes，将其转成byte切片，以便后续作为参数传入
	if err != nil {
		fmt.Println("SendPrivateMes json.Marshal err=", err)
		return
	}

	for id, uspc := range userMgr.onlineUsers {
		if id == smsPrivateMes.ReceiverID {
			//ID 匹配，检查是否在线
			_, ok := userMgr.onlineUsers[id]
			if !ok {
				//如果不在线，调用离线留言功能（后续补充）
				fmt.Println("用户不在线或不存在，无法发送私聊消息")
				return
			} else {
				smsp.SendMesToSpecifiedUser(data, uspc.Conn) //发送消息给该用户。uspc:拿到该用户与服务器的连接
				break
			}
		}
	}
}
