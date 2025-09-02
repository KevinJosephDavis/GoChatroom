package main

import (
	"fmt"
	"os"

	"github.com/kevinjosephdavis/chatroom/client/process"
)

// 定义两个变量，一个表示用户id，一个表示用户密码
var userID int
var userPassword string
var userName string

func main() {

	var choice int

	for {
		fmt.Println("\t\t欢迎使用多人聊天系统")
		fmt.Println("\t\t1 登录聊天室")
		fmt.Println("\t\t2 注册用户")
		fmt.Println("\t\t3 退出系统")
		fmt.Println("\t\t请选择（输入1-3）：")

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
