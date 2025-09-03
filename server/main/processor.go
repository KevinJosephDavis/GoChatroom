// 根据客户端的请求，调用对应的处理器完成相应的任务
package main

import (
	"fmt"
	"io"
	"net"

	"github.com/kevinjosephdavis/chatroom/common/message"
	process2 "github.com/kevinjosephdavis/chatroom/server/process"
	"github.com/kevinjosephdavis/chatroom/server/utils"
)

// Processor 先创建一个Processor结构体
type Processor struct {
	Conn net.Conn
}

// ServerProcessMes 编写一个ServerProcessMes函数
// 功能：根据客户端发送消息种类不同，决定调用哪个函数来处理
func (prc *Processor) ServerProcessMes(mes *message.Message) (err error) {
	switch mes.Type {
	case message.LoginMesType:
		//调用处理登录的函数
		//创建一个UserProcess实例
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
		//创建一个SmsProcess实例完成转发消息的功能
		smsProcess := &process2.SmsProcess{}
		smsProcess.SendGroupMes(mes)
	case message.SmsPrivateMesType:
		//创建一个SmsProcess实例完成发送私聊消息的功能
		smsProcess := &process2.SmsProcess{}
		smsProcess.SendPrivateMes(mes)
	case message.OfflineMesType:
		//创建一个SmsProcess实例完成处理用户下线的功能
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
