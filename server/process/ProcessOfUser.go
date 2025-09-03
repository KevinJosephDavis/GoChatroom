// Package process2 处理和用户相关的请求以及登录、注册、注销、用户列表管理
package process2

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/kevinjosephdavis/chatroom/common/message"
	"github.com/kevinjosephdavis/chatroom/server/model"
	"github.com/kevinjosephdavis/chatroom/server/utils"
)

type UserProcess0 struct {
	//分析应有的字段
	Conn     net.Conn
	UserID   int
	UserName string
}

// NotifyOtherOnlineUserOnline 用户上线：第二步 服务端发送变化信息给其它在线用户
func (uspc *UserProcess0) NotifyOtherOnlineUserOnline(userID int, userName string) {
	//遍历onlineUsers，然后一个一个发送
	for id, up := range userMgr.onlineUsers {
		//过滤掉自己
		if id == userID {
			continue
		}
		up.NotifyOthersOnline(userID, userName)
	}
}

func (uspc *UserProcess0) NotifyOthersOnline(userID int, userName string) {
	fmt.Printf("调试代码：广播：用户%s (ID:%d) 上线", userName, userID)
	fmt.Println()
	var mes message.Message
	mes.Type = message.NotifyUserStatusMesType

	var notifyUserStatusMes message.NotifyUserStatusMes
	notifyUserStatusMes.UserID = userID
	notifyUserStatusMes.UserName = userName
	notifyUserStatusMes.Status = message.UserOnline

	data, err := json.Marshal(notifyUserStatusMes)
	if err != nil {
		fmt.Println("NotifyOthersOnline json.Marshal err=", err)
		return
	}
	mes.Data = string(data)

	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println("NotifyOthersOnline json.Marshal err=", err)
		return
	}
	// 发送序列化后的消息给当前用户
	tf := &utils.Transfer{
		Conn: uspc.Conn,
	}
	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("NotifyOthersOnline NotifyOthers WritePkg err=", err)
	}
}

// ServerProcessRegister 处理注册请求的函数
func (uspc *UserProcess0) ServerProcessRegister(mes *message.Message) (err error) {
	var registerMes message.RegisterMes
	err = json.Unmarshal([]byte(mes.Data), &registerMes)
	if err != nil {
		fmt.Println("json.Unmarshal err=", err)
		return
	}

	var resMes message.Message
	resMes.Type = message.RegisterResMesType
	var registerResMes message.RegisterResMes

	//我们需要到redis数据库完成注册
	//使用model.MyUserDao到redis完成
	err = model.MyUserDao.Register(&registerMes.User)
	if err != nil {
		if err == model.ErrUserExists {
			registerResMes.Code = 505
			registerResMes.Error = model.ErrUserExists.Error()
		} else {
			registerResMes.Code = 506
			registerResMes.Error = "注册发生未知错误"
		}
	} else {
		registerResMes.Code = 200
	}

	data, err := json.Marshal(registerResMes)
	if err != nil {
		fmt.Println("json.Marshal err=", err)
		return
	}

	resMes.Data = string(data)

	data, err = json.Marshal(resMes)
	if err != nil {
		fmt.Println("json.Marshal err=", err)
		return
	}

	tf := &utils.Transfer{
		Conn: uspc.Conn,
	}

	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("ServerProcessRegister WritePkg err=", err)
	}
	return
}

// ServerProcessLogin 处理登录请求的函数
func (uspc *UserProcess0) ServerProcessLogin(mes *message.Message) (err error) {
	var loginMes message.LoginMes
	err = json.Unmarshal([]byte(mes.Data), &loginMes)
	if err != nil {
		fmt.Println("json.Unmarshal err=", err)
		return
	}

	var resMes message.Message
	resMes.Type = message.LoginResMesType

	var loginResMes message.LoginResMes

	//到redis数据库完成验证
	user, err := model.MyUserDao.Login(loginMes.UserID, loginMes.UserPassword)
	if err != nil {
		if err == model.ErrUserNotExists {
			loginResMes.Code = 500
			loginResMes.Error = err.Error()
		} else if err == model.ErrUserPwd {
			loginResMes.Code = 403
			loginResMes.Error = err.Error()
		} else {
			loginResMes.Code = 505
			loginResMes.Error = "服务器内部错误..."
		}

	} else {
		loginResMes.Code = 200
		//因为用户登录成功，所以要把该登录成功的用户放入到UserMgr中，表示该用户上线了
		//将登录成功的用户的userID赋值给uspc
		uspc.UserID = loginMes.UserID
		uspc.UserName = user.UserName
		userMgr.AddOnlineUser(uspc)
		uspc.NotifyOtherOnlineUserOnline(uspc.UserID, uspc.UserName) //一登录成功，就告诉其它用户自己上线了
		//将当前在线用户的ID放入到loginResMes.UserIDs

		for id, up := range userMgr.onlineUsers {
			loginResMes.UserIDs = append(loginResMes.UserIDs, id)
			loginResMes.UserNames = append(loginResMes.UserNames, up.UserName)
		}
		fmt.Println(user, "登录成功")
	}

	loginResMes.UserID = user.UserID
	loginResMes.UserName = user.UserName
	loginResMes.UserPwd = user.UserPwd
	loginResMes.UserStatus = user.UserStatus
	data, err := json.Marshal(loginResMes)
	if err != nil {
		fmt.Println("json.Marshal err=", err)
		return
	}

	resMes.Data = string(data)

	data, err = json.Marshal(resMes)
	if err != nil {
		fmt.Println("json.Marshal err=", err)
		return
	}

	tf := &utils.Transfer{
		Conn: uspc.Conn,
	}
	err = tf.WritePkg(data)
	return
}

// SendNormalOfflineMes 下线：第二步 针对用户正常退出的情况，服务端通知其它在线用户该用户下线
func (uspc *UserProcess0) SendNormalOfflineMes(mes *message.Message) {
	//因为用户正常退出，因此服务端可以收到客户端发送的mes
	var normalOfflineMes message.OfflineMes
	err := json.Unmarshal([]byte(mes.Data), &normalOfflineMes)
	if err != nil {
		fmt.Println("SendNormalOfflineMes json.Unmarshal err=", err)
		return
	}

	//反序列化后得到了下线用户的ID、昵称、下线时间
	//下一步要对服务端维护的两个map进行crud

	//1.把这个用户从onlineUser中删除，调用userMng的delete函数
	userMgr.DeleteOnlineUser(normalOfflineMes.UserID)

	//2.改变这个用户的状态
	userMgr.userStatus[normalOfflineMes.UserID] = message.UserOffline

	//服务端本身已经处理完，接下来服务端要发信息，告诉其它在线用户这个用户下线了
	mes.Type = message.OfflineResMesType
	data, err := json.Marshal(mes)
	if err != nil {
		fmt.Println("SendNormalOfflineMes json.Marshal err=", err)
		return
	}

	//通知所有其它在线用户
	for id, up := range userMgr.onlineUsers {
		if id == normalOfflineMes.UserID {
			continue //过滤掉自己
		}
		tf := &utils.Transfer{
			Conn: up.Conn,
		}
		err = tf.WritePkg(data)
		if err != nil {
			fmt.Println("SendNormalOfflineMes json.Marshal err=", err)
		}
	}
}

// SendAbnormalOfflineMes 下线：第二步 针对用户非正常退出的情况，服务端向在线用户发送某用户下线的信息
func (uspc *UserProcess0) SendAbnormalOfflineMes(userID int, userName string) {
	//由于用户非正常退出，服务器是无法接收到客户端发过来的下线信息的
	//因此，在上层需要写一个心跳检测，每隔5秒检测用户是否与服务端保持连接
	//如果用户没有保持连接又没有发送OfflineMes，就在这个if中调用这个函数

}
