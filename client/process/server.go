// Package process 1.显示登录成功界面 2.保持和服务器通讯（即启动协程） 3.当读取服务器发送的消息后，显示界面
package process

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/kevinjosephdavis/chatroom/common/message"
	"github.com/kevinjosephdavis/chatroom/server/utils"
)

// ShowMenu 显示登录成功后的界面
func ShowMenu() {
	fmt.Printf("\t\t用户 %s (ID:%d) 登录成功", CurUser.UserName, CurUser.UserID) //优化：只在登录成功后出现
	fmt.Println()
	fmt.Println()
	fmt.Println("\t\t1.显示在线用户列表")
	fmt.Println("\t\t2.开启广播")
	fmt.Println("\t\t3.私聊")
	fmt.Println("\t\t4.信息列表")
	fmt.Println("\t\t5.退出系统")
	fmt.Println("\t\t请选择1-5")
	var choice int
	var content string

	smsp := &SmsProcess{}
	fmt.Scanf("%d\n", &choice)
	switch choice {
	case 1:
		outputOnlineUser()
	case 2:
		fmt.Println("请输入您想要发送的内容：")
		fmt.Scanf("%s\n", &content)
		smsp.SendGroupMes(content)
	case 3:
		fmt.Println("请输入您想要发送的内容：")
		fmt.Scanf("%s\n", &content)
		var receiverID int
		fmt.Println("请输入您想要发送的对象ID：")
		fmt.Scanf("%d\n", &receiverID)
		if receiverID == CurUser.UserID {
			fmt.Println("无法自己向自己发送消息")
			break
		}
		smsp.SendPrivateMes(content, receiverID)
	case 4:
		fmt.Println("查看信息列表") //离线留言使用
	case 5:
		fmt.Println("您选择退出系统")
		smsp.SendOfflineMes(CurUser.UserID, CurUser.UserName, time.Now().Unix())
		time.Sleep(100 * time.Millisecond) //确保消息发送过去
		fmt.Println("再见！")
		os.Exit(0)
	default:
		fmt.Println("输入有误，请重新输入")
	}
}

// serverProcessMes 与服务端保持通讯
func serverProcessMes(conn net.Conn) {
	//创建一个transfer实例，读取服务器发送的消息
	tf := &utils.Transfer{
		Conn: conn,
	}
	for {
		mes, err := tf.ReadPkg()
		if err != nil {
			fmt.Println("tf.ReadPkg() err=", err)
			return
		}
		//如果读取到了消息，接着处理
		switch mes.Type {
		case message.NotifyUserStatusMesType: //有人上线了
			var notifyUserStatusMes message.NotifyUserStatusMes
			json.Unmarshal([]byte(mes.Data), &notifyUserStatusMes)
			updateUserStatus(&notifyUserStatusMes)
		case message.SmsMesType: //有人群发消息
			outputGroupMes(&mes)
		case message.SmsPrivateResMesType:
			outputPrivateMes(&mes)
		case message.OfflineResMesType:
			outputOfflineMes(&mes)
		default:
			fmt.Println("返回了一个未知消息类型")
		}
	}
}
