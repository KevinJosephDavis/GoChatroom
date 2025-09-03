// Package process 处理和用户相关的业务，登录、注册等等
package process

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/kevinjosephdavis/chatroom/common/message"
	"github.com/kevinjosephdavis/chatroom/server/utils"
)

type UserProcess struct {
}

// Login 写一个函数，完成登录
func (uspc *UserProcess) Login(userID int, userPassword string) (err error) {

	//1.连接到服务器
	conn, err := net.Dial("tcp", "localhost:8889")
	if err != nil {
		fmt.Println("net.Dial err=", err)
		return
	}

	//延时关闭
	defer conn.Close()

	//2.准备通过conn发送消息给服务器
	var mes message.Message
	mes.Type = message.LoginMesType

	//3.创建一个LoginMes结构体
	var loginMes message.LoginMes
	loginMes.UserID = userID
	loginMes.UserPassword = userPassword

	//4.将loginMes序列化
	data, err := json.Marshal(loginMes)
	if err != nil {
		fmt.Println("json.Marshal err=", err)
		return
	}

	//5.把data赋给mes.Data字段
	mes.Data = string(data)

	//6.将mes进行序列化
	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println("json.Marshal err=", err)
		return
	}

	//7.此时data就是我们要发送的消息
	//7.1 先把data的长度发送给服务器
	// 先获取到data的长度->转成一个表示长度的byte切片
	pkgLen := uint32(len(data))
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[0:4], pkgLen)
	//现在发送消息长度
	n, err := conn.Write(buf[:4])
	if n != 4 || err != nil {
		fmt.Println("conn.Write(bytes) err=", err)
		return
	}

	//fmt.Printf("客户端发送消息的长度=%d 内容是%s", len(data), string(data))

	//发送消息本身
	_, err = conn.Write(data)
	if err != nil {
		fmt.Println("conn.Write(data) err=", err)
		return
	}

	//处理服务器端返回的消息
	//创建一个Transfer实例
	tf := &utils.Transfer{
		Conn: conn,
	}
	mes, err = tf.ReadPkg()
	//客户端输入账号密码，序列化后传给服务端，服务端读取之后将反序列化，返回mes给客户端
	if err != nil {
		fmt.Println("readPkg(conn) err=", err)
	}

	//将mes的data部分反序列化成LoginResMes，判断其状态码
	var loginResMes message.LoginResMes
	err = json.Unmarshal([]byte(mes.Data), &loginResMes)
	if loginResMes.Code == 200 {
		//fmt.Println("登录成功")
		//初始化CurUser
		CurUser.Conn = conn
		CurUser.UserID = loginResMes.UserID
		CurUser.UserStatus = message.UserOnline
		CurUser.UserName = loginResMes.UserName

		fmt.Println("当前在线用户如下：")
		for i, v := range loginResMes.UserIDs {
			if v == loginResMes.UserID {
				continue //不显示自己在线
			}
			fmt.Printf("%s (ID：%d)\t", loginResMes.UserNames[i], v)

			//完成客户端的onlineUsers的初始化
			user := &message.User{
				UserID:     v,
				UserStatus: message.UserOnline,
			}
			onlineUsers[v] = user
		}
		fmt.Println()

		//这里还需要在客户端起一个协程，保持和服务器端的通讯。如果服务器有数据推送，及时接收并显示在客户端的终端
		go serverProcessMes(conn)

		//1.显示二级菜单（登录成功后的菜单）循环显示
		for {
			ShowMenu()
		}
	} else {
		fmt.Println(loginResMes.Error)
	}
	return
}

// Register 完成用户注册
func (uspc *UserProcess) Register(userID int, userPwd string, userName string) (err error) {
	//1.连接到服务器
	conn, err := net.Dial("tcp", "localhost:8889")
	if err != nil {
		fmt.Println("net.Dial err=", err)
		return
	}
	//延时关闭
	defer conn.Close()

	//2.准备通过conn发送消息给服务器
	var mes message.Message
	mes.Type = message.RegisterMesType

	//3.创建一个LoginMes结构体
	var registerMes message.RegisterMes
	registerMes.User.UserID = userID
	registerMes.User.UserPwd = userPwd
	registerMes.User.UserName = userName

	//4.将registerMes序列化
	data, err := json.Marshal(registerMes)
	if err != nil {
		fmt.Println("json.Marshal err=", err)
		return
	}

	//5.把data赋给mes.Data字段
	mes.Data = string(data)

	//6.将mes进行序列化
	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println("json.Marshal err=", err)
		return
	}

	//创建一个Transfer实例
	tf := &utils.Transfer{
		Conn: conn,
	}

	//发送data给服务器端
	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("注册发送信息错误 err=", err)
	}

	mes, err = tf.ReadPkg()
	//客户端输入账号密码昵称，序列化后传给服务端，服务端读取之后将反序列化，返回mes给客户端
	if err != nil {
		fmt.Println("readPkg(conn) err=", err)
	}

	//将mes的data部分反序列化成ResMes，判断其状态码
	var registerResMes message.RegisterResMes
	err = json.Unmarshal([]byte(mes.Data), &registerResMes)
	if registerResMes.Code == 200 {
		fmt.Println("注册成功，请重新登录")
		os.Exit(0)
	} else {
		fmt.Println(registerResMes.Error)
		os.Exit(0)
	}
	return
}
