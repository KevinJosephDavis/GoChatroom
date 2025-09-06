// Package process 1.显示登录成功界面 2.保持和服务器通讯（即启动协程） 3.当读取服务器发送的消息后，显示界面
package process

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/kevinjosephdavis/chatroom/common/message"
	"github.com/kevinjosephdavis/chatroom/server/utils"
)

// ShowMenu 显示登录成功后的界面
func ShowMenu() bool {
	fmt.Println()
	fmt.Println()
	fmt.Println("\t1.显示在线用户列表")
	fmt.Println("\t2.开启广播")
	fmt.Println("\t3.私聊")
	fmt.Println("\t4.信息列表")
	fmt.Println("\t5.退出系统")
	fmt.Println("\t6.注销账号")
	fmt.Println("\t请选择1-6")
	var choice int
	var content string

	smsp := &SmsProcess{}
	_, err := fmt.Scanf("%d\n", &choice)
	if err != nil {
		//清空输入缓冲区，避免残留的换行符影响后续输入
		var discard string
		fmt.Scanln(&discard)
		fmt.Println("输入有误，请重新输入")
		return false
	}
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
		//退回到上一级菜单
		return true

	case 6:
		fmt.Println("您选择注销用户")
		fmt.Println("您确认要注销用户吗?(y/n)：")
		var confirm string
		fmt.Scanf("%s\n", &confirm)
		if confirm != "y" && confirm != "Y" {
			fmt.Println("取消注销操作")
			return false //继续循环
		}

		//发送注销请求，等待服务器回调outputDeleteAccountMes
		currentTime := time.Now().Unix()
		err := smsp.SendDeleteAccountMes(CurUser.UserID, CurUser.UserName, currentTime)
		if err != nil {
			fmt.Println("发送注销请求失败，err=", err)
			return false
		}
		fmt.Println("注销请求已发送，请等待处理...")
		return false

	default:
		fmt.Println("输入有误，请重新输入")
		return false
	}
	return false //继续循环
}

// serverProcessMes 与服务端保持通讯
func serverProcessMes(ctx context.Context, conn net.Conn) {
	//创建一个transfer实例，读取服务器发送的消息
	tf := &utils.Transfer{
		Conn: conn,
	}
	for {
		select {
		case <-ctx.Done(): //当cancel()被调用时，这里会收到信号
			//收到取消信号，退出
			fmt.Println("收到context取消信号，协程退出")
			return
		default:
			//设置读取超时，避免永久堵塞
			tf.Conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond)) //如果超时就continue
			mes, err := tf.ReadPkg()
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					//超时，继续检查context
					continue
				}
				if strings.Contains(err.Error(), "use of closed network connection") {
					fmt.Println("连接已被关闭")
					return
				}
				fmt.Println("tf.ReadPkg() err=", err)
				return
			}
			//如果读取到了消息，接着处理
			processMessageType(mes)
		}
	}
}

// processMessageType 处理消息类型
func processMessageType(mes message.Message) {
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
	case message.DeleteAccountResMesType:
		outputDeleteAccountMes(&mes)
	default:
		fmt.Println("返回了一个未知消息类型")
	}
}
