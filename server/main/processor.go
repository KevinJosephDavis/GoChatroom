// 根据客户端的请求，调用对应的处理器完成相应的任务
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/kevinjosephdavis/chatroom/common/message"
	process2 "github.com/kevinjosephdavis/chatroom/server/process"
	"github.com/kevinjosephdavis/chatroom/server/utils"
)

// Processor 服务端控制器
type Processor struct {
	Conn net.Conn
}

// ServerProcessMes 根据客户端发送消息种类不同，决定调用哪个函数来处理
func (prc *Processor) ServerProcessMes(mes *message.Message) (err error) {
	switch mes.Type {
	case message.LoginMesType:
		//调用处理登录的函数
		uspc := &process2.UserProcess0{
			Conn: prc.Conn,
		}
		err = uspc.ServerProcessLogin(mes)
	case message.RegisterMesType:
		//调用处理注册的函数
		uspc := &process2.UserProcess0{
			Conn: prc.Conn,
		}
		err = uspc.ServerProcessRegister(mes)
	case message.SmsMesType:
		//处理广播
		smsProcess := &process2.SmsProcess{}
		smsProcess.SendGroupMes(mes)
	case message.SmsPrivateMesType:
		//处理用户私聊
		smsProcess := &process2.SmsProcess{}
		smsProcess.SendPrivateMes(mes)
	case message.LogoutMesType:
		//处理用户下线
		smsProcess := &process2.SmsProcess{}
		smsProcess.SendNormalLogoutMes(mes)
	case message.DeleteAccountMesType:
		//处理用户注销
		smsProcess := &process2.SmsProcess{}
		smsProcess.SendDeleteAccountMes(mes)
	case message.HeartBeatType:
		//收到心跳检测结果
		var heartBeat struct {
			UserID int `json:"userID"`
		}
		err := json.Unmarshal([]byte(mes.Data), &heartBeat)
		if err != nil {
			fmt.Println("case message.HeartBeatType json.Unmarshal err=", err)
			break
		}
		//更新该用户最后的心跳时间
		if value, exist := process2.GetUserMgr().OnlineUsers.Load(heartBeat.UserID); exist {
			uspc := value.(*process2.UserProcess0)
			uspc.LastHeartBeat = time.Now()
			fmt.Printf("收到用户%d的心跳，更新时间：%v\n", heartBeat.UserID, uspc.LastHeartBeat)
		}
	default:
		fmt.Println("消息类型不存在，无法处理...")
	}
	return
}

func (prc *Processor) process2() (err error) {
	//循环读取客户端发送的信息
	for {
		//这里即将读取数据包，直接封装成一个函数readPkg(),返回Message,Err
		//创建一个Transfer实例，完成读包任务
		tf := &utils.Transfer{
			Conn: prc.Conn,
		}
		mes, err1 := tf.ReadPkg()
		if err1 != nil {
			if err1.Error() == "客户端关闭了连接" {
				fmt.Printf("客户端 %s 主动关闭了连接", tf.Conn.RemoteAddr())
				return
			}
			if err1 == io.EOF {
				fmt.Println("客户端退出，服务器端也退出")
				return err1
			} else {
				fmt.Println("readPkg err=", err1)
				return err1
			}
		}
		err = prc.ServerProcessMes(&mes)
		if err != nil {
			return
		}
	}
}
