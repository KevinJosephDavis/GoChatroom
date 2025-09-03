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

// SendNormalOfflineMes 下线：第二步 针对用户正常退出的情况，服务端通知其它在线用户该用户下线
func (smsp *SmsProcess) SendNormalOfflineMes(mes *message.Message) {
	//因为用户正常退出，因此服务端可以收到客户端发送的mes
	var normalOfflineMes message.OfflineMes
	err := json.Unmarshal([]byte(mes.Data), &normalOfflineMes)
	if err != nil {
		fmt.Println("SendNormalOfflineMes json.Unmarshal err=", err)
		return
	}

	//反序列化后得到了下线用户的ID、昵称、下线时间
	//下一步要对服务端维护的两个map进行crud

	//1.把这个用户从onlineUser中删除，调用userMgr的delete函数
	userMgr.DeleteOnlineUser(normalOfflineMes.UserID)

	//2.改变这个用户的状态
	//userMgr.userStatus[normalOfflineMes.UserID] = message.UserOffline
	//使用sync.map，避免初始化，保证线程安全
	userMgr.userStatus.Store(normalOfflineMes.UserID, message.UserOffline)

	//服务端自身已经处理完，接下来服务端要发信息，告诉其它在线用户这个用户下线了
	var resMes message.Message
	resMes.Type = message.OfflineResMesType

	var offlineResMes message.OfflineResMes
	offlineResMes.UserID = normalOfflineMes.UserID
	offlineResMes.UserName = normalOfflineMes.UserName
	offlineResMes.Time = normalOfflineMes.Time
	offlineResMes.Reason = normalOfflineMes.Reason
	data, err := json.Marshal(offlineResMes)
	if err != nil {
		fmt.Println("SendNormalOfflineMes json.Marshal err=", err)
		return
	}

	resMes.Data = string(data)

	FinalData, err := json.Marshal(resMes)
	if err != nil {
		fmt.Println("SendNormalOfflineMes json.Marshal err=", err)
		return
	}

	//通知所有其它在线用户
	for id, up := range userMgr.onlineUsers {
		if id == normalOfflineMes.UserID {
			continue //过滤掉自己。其实按道理来说不会出现这种情况，因为前面已经delete这个下线用户了
		}
		tf := &utils.Transfer{
			Conn: up.Conn,
		}
		err = tf.WritePkg(FinalData)
		if err != nil {
			fmt.Println("SendNormalOfflineMes json.Marshal err=", err)
		}
	}
}

// SendAbnormalOfflineMes 下线：第二步 针对用户非正常退出的情况，服务端向在线用户发送某用户下线的信息
func (smsp *SmsProcess) SendAbnormalOfflineMes(userID int, userName string) {
	//由于用户非正常退出，服务器是无法接收到客户端发过来的下线信息的
	//因此，在上层需要写一个心跳检测，每隔5秒检测用户是否与服务端保持连接
	//如果用户没有保持连接又没有发送OfflineMes，就在这个if中调用这个函数

}
