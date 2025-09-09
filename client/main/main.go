package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kevinjosephdavis/chatroom/client/model"
	"github.com/kevinjosephdavis/chatroom/client/process"
)

// 定义两个变量，一个表示用户id，一个表示用户密码
var userID int
var userPassword string
var userName string
var CurUser model.CurUser

func GetCurUser() *model.CurUser {
	return &CurUser
}

func SetCurUser(user model.CurUser) {
	CurUser = user
}

// SetUpSignalHandler 设置信号处理器
func SetUpSignalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM) //捕获ctrl+C和终止信号

	go func() {
		<-c
		//fmt.Println("\n接收到信号，正在退出...")
		curUser := model.GetCurUser()
		if curUser != nil && curUser.Conn != nil && curUser.UserID != 0 {
			smsp := &process.SmsProcess{}
			err := smsp.SendLogoutMes(curUser.UserID, curUser.UserName, time.Now().Unix())
			if err != nil {
				fmt.Println("SetUpSignalHandler 发送CTRL+C下线消息失败")
				return
			}

			time.Sleep(100 * time.Millisecond)
			curUser.Conn.Close()
			model.ClearCurUser()
		}
		fmt.Println("再见！")
		os.Exit(0)
	}()
}
func main() {
	var choice int
	SetUpSignalHandler()
	for {
		fmt.Println("\t欢迎使用多人聊天系统")
		fmt.Println("\t1 登录聊天室")
		fmt.Println("\t2 注册用户")
		fmt.Println("\t3 退出系统")
		fmt.Println("\t请选择（输入1-3）：")

		fmt.Scanf("%d\n", &choice)
		switch choice {
		case 1:
			fmt.Println("登录聊天室")
			fmt.Println("请输入用户ID")
			fmt.Scanf("%d\n", &userID)
			fmt.Println("请输入用户密码")
			fmt.Scanf("%s\n", &userPassword)

			uspc := &process.UserProcess{}
			uspc.Login(userID, userPassword)
		case 2:
			fmt.Println("注册用户")
			fmt.Println("请输入用户id：")
			fmt.Scanf("%d\n", &userID)
			fmt.Println("请输入用户密码：")
			fmt.Scanf("%s\n", &userPassword)
			fmt.Println("请输入用户昵称：")
			fmt.Scanf("%s\n", &userName)

			uspc := &process.UserProcess{}
			uspc.Register(userID, userPassword, userName)
		case 3:
			fmt.Println("退出系统")
			os.Exit(0)
		default:
			fmt.Println("您的输入有误，请重新输入")
		}
	}
}
