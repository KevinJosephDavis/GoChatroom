// Package process2 处理和短消息相关的请求。群聊、点对点聊天。
package process2

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/kevinjosephdavis/chatroom/common/message"
	"github.com/kevinjosephdavis/chatroom/server/model"
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
	GetUserMgr().onlineUsers.Range(func(key, value interface{}) bool {
		id := key.(int)
		up := value.(*UserProcess0)
		if id == groupMes.Sender.UserID {
			return true //继续遍历
		}
		smsp.SendMesToEachOnlineUser(data, up.Conn) //获取每个在线用户的控制器，得到其与服务端的连接，进而发送信息
		return true
	})
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
	data, err := json.Marshal(mes)
	if err != nil {
		fmt.Println("SendPrivateMes json.Marshal err=", err)
		return
	}

	// 直接查找目标用户，不需要遍历，从O(n)优化为O(1)
	value, exist := GetUserMgr().onlineUsers.Load(smsPrivateMes.ReceiverID)
	if !exist {
		fmt.Println("用户不在线或不存在，无法发送私聊消息")
		// 调用离线留言功能
		return
	}
	uspc := value.(*UserProcess0)

	// 发送私聊消息
	smsp.SendMesToSpecifiedUser(data, uspc.Conn)
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
	GetUserMgr().DeleteOnlineUser(normalOfflineMes.UserID)

	//2.改变这个用户的状态
	GetUserMgr().userStatus.Store(normalOfflineMes.UserID, message.UserOffline)

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
	GetUserMgr().onlineUsers.Range(func(key, value interface{}) bool {
		id, ok := key.(int)
		if !ok {
			return true //跳过无效key
		}
		if id == normalOfflineMes.UserID {
			return true //继续遍历，过滤掉自己
			// 其实按道理来说不会出现这种情况，因为前面已经delete该下线用户了
		}
		up, ok := value.(*UserProcess0)
		if !ok || up == nil {
			return true //跳过无效up
		}
		if up.Conn == nil {
			return true
		}
		tf := &utils.Transfer{
			Conn: up.Conn,
		}

		err = tf.WritePkg(FinalData)
		if err != nil {
			fmt.Println("SendNormalOfflineMes json.Marshal err=", err)
		}
		return true
	})
}

// SendAbnormalOfflineMes 下线：第二步 针对用户非正常退出的情况，服务端向在线用户发送该用户下线这一消息
func (smsp *SmsProcess) SendAbnormalOfflineMes(userID int, userName string) {
	//由于用户非正常退出，服务器是无法接收到客户端发过来的下线信息的
	//因此，在上层需要写一个心跳检测，每隔5秒检测用户是否与服务端保持连接
	//如果用户没有保持连接又没有发送OfflineMes，就在这个if中调用这个函数

}

// SendDeleteAccountMes 注销：第二步 服务端将注销用户从两个map中delete，并向其它在线用户发送该用户注销这一消息
func (smsp *SmsProcess) SendDeleteAccountMes(mes *message.Message) {
	var DeleteAccountMes message.DeleteAccountMes
	err := json.Unmarshal([]byte(mes.Data), &DeleteAccountMes)
	if err != nil {
		fmt.Println("SendDeleteAccountMes json.Unmarshal err=", err)
		return
	}

	//服务端要告诉其它在线用户有用户注销了
	var resMes message.Message
	resMes.Type = message.DeleteAccountResMesType

	var deleteAccountResMes message.DeleteAccountResMes
	deleteAccountResMes.User.UserID = DeleteAccountMes.User.UserID
	deleteAccountResMes.User.UserName = DeleteAccountMes.User.UserName
	deleteAccountResMes.Time = DeleteAccountMes.Time
	data, err := json.Marshal(deleteAccountResMes)
	if err != nil {
		fmt.Println("SendDeleteAccountMes json.Marshal err=", err)
		return
	}

	resMes.Data = string(data)

	FinalData, err := json.Marshal(resMes)
	if err != nil {
		fmt.Println("SendDeleteAccountMes json.Marshal err=", err)
		return
	}

	//向注销用户自己发送通知（在其被删除之前）
	if up, exists := GetUserMgr().onlineUsers.Load(DeleteAccountMes.User.UserID); exists {
		if uspc, ok := up.(*UserProcess0); ok && uspc.Conn != nil {
			tf := &utils.Transfer{
				Conn: uspc.Conn,
			}
			err := tf.WritePkg(FinalData)
			if err != nil {
				fmt.Println("向注销用户发送通知失败：", err)
			}
			//稍微延迟，确保消息发送完成
			time.Sleep(10 * time.Millisecond)
		}
	}

	//得到了下线用户的ID、昵称、下线时间
	//下一步将其从这两个map中删除
	GetUserMgr().DeleteOnlineUser(DeleteAccountMes.User.UserID)
	GetUserMgr().DeleteExistUser(DeleteAccountMes.User.UserID)

	err = model.MyUserDao.DeleteAccount(&DeleteAccountMes.User)
	if err != nil {
		fmt.Println("销户失败，err=", err)
		fmt.Println()
		return
	}

	//通知其它在线用户
	GetUserMgr().onlineUsers.Range(func(key, value interface{}) bool {
		up, ok := value.(*UserProcess0)
		if !ok || up == nil || up.Conn == nil {
			return true
		}
		tf := &utils.Transfer{
			Conn: up.Conn,
		}
		err = tf.WritePkg(FinalData)
		if err != nil {
			fmt.Println("SendDeleteAccountMes json.Marshal err=", err)
		}
		return true
	})
}
