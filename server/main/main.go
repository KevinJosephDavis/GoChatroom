// 1.监听 2.等待客户端连接 3.做初始化工作
package main

import (
	"fmt"
	"net"
	"time"

	"github.com/kevinjosephdavis/chatroom/server/model"
)

// process 处理和客户端的通讯
func process(conn net.Conn) {
	//延时关闭conn
	defer conn.Close()

	//这里调用总控，创建一个Processor实例
	processor := &Processor{
		Conn: conn,
	}
	err := processor.process2()
	if err != nil {
		fmt.Println("客户端和服务器端通讯协程错误 err=", err)
		return
	}
}

// 编写一个函数，完成对UserDao的初始化任务
func initUserDao() {
	//这里的pool本身就是一个全局的变量。先initPool再initUserDao
	model.MyUserDao = model.NewUserDao(pool)
}

func init() {
	//当服务器启动时，我们就去初始化连接池
	initPool("localhost:6379", 16, 0, 300*time.Second)
	initUserDao()
}
func main() {

	//提示信息
	fmt.Println("服务器在8889端口监听....")
	listen, err := net.Listen("tcp", "0.0.0.0:8889")

	if err != nil {
		fmt.Println("net.Listen err=", err)
		return
	}
	defer listen.Close()

	//如果监听成功，等待客户端来连接服务器
	for {
		fmt.Println("等待客户端来连接服务器....")
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("listen.Accpet err=", err)
		}

		//如果连接成功，则启动一个协程和客户端保持通讯
		go process(conn)
	}
}
