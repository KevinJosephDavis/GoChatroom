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
	GetUserMgr().OnlineUsers.Range(func(key, value interface{}) bool {
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

	// 离线留言：第一步：发送对象不存在的错误处理
	status, err := GetUserMgr().GetUserStatus(smsPrivateMes.ReceiverID)
	if err != nil {
		if status == -1 {
			//要告诉客户端，这个用户不存在
			errorMes := map[string]interface{}{
				"type":    message.ErrorResType,
				"code":    "UserNotExist",
				"message": "发送对象不存在",
			}
			smsp.SendErrorToSender(smsPrivateMes.Sender.UserID, errorMes)
		} else {
			fmt.Println("系统错误，err=", err)
		}
		return
	}

	mes.Type = message.SmsPrivateResMesType
	data, err := json.Marshal(mes)
	if err != nil {
		fmt.Println("SendPrivateMes json.Marshal err=", err)
		return
	}

	// 直接查找目标用户，不需要遍历，从O(n)优化为O(1)
	value, exist := GetUserMgr().OnlineUsers.Load(smsPrivateMes.ReceiverID)
	if !exist {
		fmt.Println("用户不在线") //由于在遍历onlineUsers之前，已经判断过用户是否存在。所以这里只可能是用户不在线
		// 离线留言：第二步：默认调用离线留言功能

		//1.先创建一个离线留言消息
		offlineMes := message.OfflineMes{
			SenderID:   smsPrivateMes.Sender.UserID,
			SenderName: smsPrivateMes.Sender.UserName,
			ReceiverID: smsPrivateMes.ReceiverID,
			Content:    smsPrivateMes.Content,
			Time:       time.Now().Unix(),
		}

		//2.将离线留言消息存储到服务端中。用户登录后自动发送查看离线留言信息的请求，这时候服务端再返回离线留言信息
		GetUserMgr().StoreOfflineMes(&offlineMes)
		fmt.Printf("用户(ID:%d) 的离线消息已存储", smsPrivateMes.ReceiverID)

		//3.存储离线消息后，要告诉发送人离线留言成功
		SucStoreOfflineMes := map[string]interface{}{
			"type":    message.ErrorResType,
			"code":    "OfflineMesStored",
			"message": "消息已存储，将在对方上线时送达，离线留言成功",
		}
		smsp.SendErrorToSender(smsPrivateMes.Sender.UserID, SucStoreOfflineMes)

		return
	}
	uspc := value.(*UserProcess0)

	// 发送私聊消息
	smsp.SendMesToSpecifiedUser(data, uspc.Conn)
}

// BroadcastLogoutNotification 下线：第二步 服务端通知其它在线用户该用户下线（正常下线与非正常下线的公共部分）
func (smsp *SmsProcess) BroadcastLogoutNotification(logoutResMes message.LogoutResMes) {
	data, err := json.Marshal(logoutResMes)
	if err != nil {
		fmt.Println("BroadcastLogoutNotification json.Marshal err=", err)
		return
	}

	resMes := message.Message{
		Type: message.LogoutResMesType,
		Data: string(data),
	}

	FinalData, err := json.Marshal(resMes)
	if err != nil {
		fmt.Println("BroadcastLogoutNotification json.Marshal err=", err)
		return
	}

	//通知所有其它在线用户
	GetUserMgr().OnlineUsers.Range(func(key, value interface{}) bool {
		id, ok := key.(int)
		if !ok {
			return true //跳过无效key
		}
		if id == logoutResMes.UserID {
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
			fmt.Println("BroadcastLogoutNotification json.Marshal err=", err)
		}
		return true
	})
}

// SendNormalLogoutMes 下线：第二步 正常下线（输入5或者键入ctrl+C）
func (smsp *SmsProcess) SendNormalLogoutMes(mes *message.Message) {
	//因为用户正常退出，因此服务端可以收到客户端发送的mes
	var normalLogoutMes message.LogoutMes
	err := json.Unmarshal([]byte(mes.Data), &normalLogoutMes)
	if err != nil {
		fmt.Println("SendNormalLogoutMes json.Unmarshal err=", err)
		return
	}

	//检查该用户是否已经心跳超时
	if value, exist := GetUserMgr().OnlineUsers.Load(normalLogoutMes.UserID); exist {
		uspc := value.(*UserProcess0)
		if time.Since(uspc.LastHeartBeat) > 10*time.Second {
			//实际上是心跳超时，按异常下线处理
			smsp.SendAbnormalLogoutResMes(normalLogoutMes.UserID, normalLogoutMes.UserName)
			return
		}
	}

	//反序列化后得到了下线用户的ID、昵称、下线时间
	//下一步要对服务端维护的两个map进行crud

	//1.把这个用户从onlineUser中删除，调用userMgr的delete函数
	GetUserMgr().DeleteOnlineUser(normalLogoutMes.UserID)

	//2.改变这个用户的状态
	GetUserMgr().userStatus.Store(normalLogoutMes.UserID, message.UserOffline)

	logoutResMes := message.LogoutResMes(normalLogoutMes) //由于两个struct字段一样，所以直接类型转换
	smsp.BroadcastLogoutNotification(logoutResMes)
}

// SendAbnormalLogoutResMes 下线：第二步 非正常下线（直接关闭终端或网络波动...）
func (smsp *SmsProcess) SendAbnormalLogoutResMes(userID int, userName string) {
	//1.把这个用户从onlineUser中删除，调用userMgr的delete函数
	GetUserMgr().DeleteOnlineUser(userID)

	//2.改变这个用户的状态
	GetUserMgr().userStatus.Store(userID, message.UserOffline)

	logoutResMes := message.LogoutResMes{
		UserID:   userID,
		UserName: userName,
		Time:     time.Now().Unix(),
	}
	smsp.BroadcastLogoutNotification(logoutResMes)
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
	if up, exists := GetUserMgr().OnlineUsers.Load(DeleteAccountMes.User.UserID); exists {
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
	GetUserMgr().OnlineUsers.Range(func(key, value interface{}) bool {
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

// SendErrorToSender 如果私聊信息的接收者不存在，或者离线留言成功了，服务端要告诉客户端
func (smsp *SmsProcess) SendErrorToSender(SenderID int, errorMes map[string]interface{}) {
	value, exist := GetUserMgr().OnlineUsers.Load(SenderID)
	if !exist {
		return //发送者也不在线，无法通知
	}

	data, err := json.Marshal(errorMes)
	if err != nil {
		fmt.Println("SendErrorToSender json.Marshal err=", err)
		return
	}

	mes := message.Message{
		Type: "ErrorRes",
		Data: string(data),
	}

	FinalData, err := json.Marshal(mes)
	if err != nil {
		fmt.Println("SendErrorToSender json.Marshal err=", err)
		return
	}

	uspc := value.(*UserProcess0)
	smsp.SendMesToSpecifiedUser(FinalData, uspc.Conn)
}
