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

// NotifyOtherOnlineUser 写通知所有在线用户的方法。userID要通知其它在线用户自己上线了
func (uspc *UserProcess0) NotifyOtherOnlineUser(userID int) {
	//遍历onlineUsers，然后一个一个发送
	for id, up := range userMgr.onlineUsers {
		//过滤掉自己
		if id == userID {
			continue
		}
		//开始通知（单独写一个方法）
		up.NotifyOthers(userID)
	}
}

func (uspc *UserProcess0) NotifyOthers(userID int) {
	var mes message.Message
	mes.Type = message.NotifyUserStatusMesType

	var notifyUserStatusMes message.NotifyUserStatusMes
	notifyUserStatusMes.UserID = userID
	notifyUserStatusMes.Status = message.UserOnline

	//序列化
	data, err := json.Marshal(notifyUserStatusMes)
	if err != nil {
		fmt.Println("json.Marshal err=", err)
		return
	}
	//将序列化后的NotifyUserStatusMes赋给data
	mes.Data = string(data)

	//对mes再次序列化，准备发送
	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println("json.Marshal err=", err)
		return
	}
	// 发送序列化后的消息给当前用户
	tf := &utils.Transfer{
		Conn: uspc.Conn,
	}
	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("NotifyOthers WritePkg err=", err)
	}

}

// ServerProcessRegister 处理注册请求的函数
func (uspc *UserProcess0) ServerProcessRegister(mes *message.Message) (err error) {
	//先从mes中取出其data，并反序列化成ResigerMes
	var registerMes message.RegisterMes
	err = json.Unmarshal([]byte(mes.Data), &registerMes)
	if err != nil {
		fmt.Println("json.Unmarshal err=", err)
		return
	}

	//注意registerMes中只有一个User结构体字段
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

	//现在data：状态码+错误信息。转成了byte切片
	//将data赋值给resMes
	resMes.Data = string(data)

	//对resMes进行序列化，准备发送
	data, err = json.Marshal(resMes)
	if err != nil {
		fmt.Println("json.Marshal err=", err)
		return
	}

	//发送data
	tf := &utils.Transfer{
		Conn: uspc.Conn,
	}

	err = tf.WritePkg(data)
	return
}

// ServerProcessLogin 处理登录请求的函数
func (uspc *UserProcess0) ServerProcessLogin(mes *message.Message) (err error) {
	//1.先从mes中取出其data，并反序列化成一个LoginMes
	var loginMes message.LoginMes
	err = json.Unmarshal([]byte(mes.Data), &loginMes)
	if err != nil {
		fmt.Println("json.Unmarshal err=", err)
		return
	}

	//1 声明一个 resMes
	var resMes message.Message
	resMes.Type = message.LoginResMesType

	//2 声明一个loginResMes并完成赋值
	var loginResMes message.LoginResMes

	//我们需要到redis数据库完成验证
	//1.使用model.MyUserDao到redis完成验证
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
		uspc.NotifyOtherOnlineUser(uspc.UserID) //登录成功，就告诉其它用户自己上线了
		//将当前在线用户的ID放入到loginResMes.UserIDs
		//遍历userMgr.onlineUsers
		for id, up := range userMgr.onlineUsers {
			loginResMes.UserIDs = append(loginResMes.UserIDs, id)
			loginResMes.UserNames = append(loginResMes.UserNames, up.UserName)
		}
		fmt.Println(user, "登录成功")
	}

	//3 将loginResMes序列化
	loginResMes.UserID = user.UserID
	loginResMes.UserName = user.UserName
	loginResMes.UserPwd = user.UserPwd
	loginResMes.UserStatus = user.UserStatus
	data, err := json.Marshal(loginResMes)
	if err != nil {
		fmt.Println("json.Marshal err=", err)
		return
	}

	//4 将data赋值给resMes
	resMes.Data = string(data)

	//5 对resMes进行序列化，准备发送
	data, err = json.Marshal(resMes)
	if err != nil {
		fmt.Println("json.Marshal err=", err)
		return
	}

	//6 发送data 将其封装到writePkg() 调用writePkg()
	tf := &utils.Transfer{
		Conn: uspc.Conn,
	}
	err = tf.WritePkg(data)
	return
}
