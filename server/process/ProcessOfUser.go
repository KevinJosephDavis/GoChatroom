// Package process2 处理和用户相关的请求以及登录、注册、注销、用户列表管理
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

type UserProcess0 struct {
	//分析应有的字段
	Conn          net.Conn
	UserID        int
	UserName      string
	LastHeartBeat time.Time //最后一次心跳的时间
	IsOnline      bool
}

// NotifyOtherOnlineUserOnline 用户上线：第二步 服务端发送变化信息给其它在线用户
func (uspc *UserProcess0) NotifyOtherOnlineUserOnline(userID int, userName string) {
	//遍历onlineUsers，然后一个一个发送
	GetUserMgr().OnlineUsers.Range(func(key, value interface{}) bool {
		id, ok := key.(int)
		if !ok {
			return true
		}
		up, ok0 := value.(*UserProcess0)
		if !ok0 || up == nil {
			return true
		}
		//过滤掉自己
		if id == userID {
			return true
		}
		up.NotifyOthersOnline(userID, userName)
		return true
	})
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

// ServerProcessLogin 登录：第二步
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
		switch err {
		case model.ErrUserNotExists:
			loginResMes.Code = 500
			loginResMes.Error = err.Error()
		case model.ErrUserPwd:
			loginResMes.Code = 403
			loginResMes.Error = err.Error()
		default:
			loginResMes.Code = 505
			loginResMes.Error = "服务器内部错误..."
		}
	} else {
		//添加安全检查
		if user == nil {
			loginResMes.Code = 505
			loginResMes.Error = "服务器内部错误：用户数据为空"
		} else {
			loginResMes.Code = 200
			//因为用户登录成功，所以要把该登录成功的用户放入到UserMgr中，表示该用户上线了
			//将登录成功的用户的userID赋值给uspc
			uspc.UserID = loginMes.UserID
			uspc.UserName = user.UserName
			uspc.LastHeartBeat = time.Now()
			uspc.IsOnline = true
			GetUserMgr().AddOnlineUser(uspc)
			GetUserMgr().SetUserStatus(uspc.UserID, message.UserOnline)
			uspc.NotifyOtherOnlineUserOnline(uspc.UserID, uspc.UserName) //一登录成功，就告诉其它用户自己上线了
			//将当前在线用户的ID放入到loginResMes.UserIDs
			GetUserMgr().OnlineUsers.Range(func(key, value interface{}) bool {
				id, ok := key.(int)
				if !ok {
					return true
				}
				up, ok0 := value.(*UserProcess0)
				if !ok0 || up == nil {
					return true
				}
				loginResMes.UserIDs = append(loginResMes.UserIDs, id)
				loginResMes.UserNames = append(loginResMes.UserNames, up.UserName)
				return true
			})
			fmt.Println(user, "登录成功")

			go func() {
				//用户登录成功后，自动查询离线留言列表
				time.Sleep(300 * time.Millisecond) //确保客户端正确显示了菜单
				offlineMes := GetUserMgr().GetOfflineMes(uspc.UserID)
				if len(offlineMes) > 0 {
					//如果有离线留言
					for _, mes := range offlineMes {
						offlineResMes := message.OfflineResMes{
							SenderID:   mes.SenderID,
							SenderName: mes.SenderName,
							Content:    mes.Content,
							Time:       mes.Time,
						}

						//序列化，准备发送
						resData, err := json.Marshal(offlineResMes)
						if err != nil {
							fmt.Println("离线留言消息序列化错误，err=", err)
							continue
						}

						msg := message.Message{
							Type: message.OfflineResMesType,
							Data: string(resData),
						}

						FinalData, err := json.Marshal(msg)
						if err != nil {
							fmt.Println("离线留言消息序列化错误，err=", err)
							continue
						}
						smsp := SmsProcess{}
						smsp.SendMesToSpecifiedUser(FinalData, uspc.Conn)
					}

					//投递完毕，清空离线留言列表
					fmt.Printf("用户%s (ID:%d) 的 %d条离线消息已发送", uspc.UserName, uspc.UserID,
						len(offlineMes))

					GetUserMgr().ClearOfflineMes(uspc.UserID)
				}
			}()
		}

	}

	if user != nil {
		loginResMes.UserID = user.UserID
		loginResMes.UserName = user.UserName
		loginResMes.UserPwd = user.UserPwd
		loginResMes.UserStatus = user.UserStatus
	}

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
