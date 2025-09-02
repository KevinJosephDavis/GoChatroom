// Package message provides shared types and structures for the chatroom server.
package message

const (
	LoginMesType            = "LoginMes"
	LoginResMesType         = "LoginResMes"
	RegisterMesType         = "RegisterMes"
	RegisterResMesType      = "RegisterResMes"
	NotifyUserStatusMesType = "NotifyUserStatusMes"
	SmsMesType              = "SmsMes"
	SmsPrivateMesType       = "SmsPrivateMes"
	SmsPrivateResMesType    = "SmsPrivateResMes"
)

//定义几个用户状态的常量
const (
	UserOnline = iota
	UserOffline
	UserBusyStatus
)

type Message struct {
	Type string `json:"type"` //消息类型
	Data string `json:"data"`
}

//LoginMes 先定义两个消息，后面需要再增加
type LoginMes struct {
	UserID       int    `json:"userID"`       // 用户ID
	UserPassword string `json:"userPassword"` // 用户密码
	UserName     string `json:"userName"`     // 用户名
}

type LoginResMes struct {
	Code      int      `json:"code"`      // 返回状态码 500 表示该用户未注册 200 表示登陆成功
	Error     string   `json:"error"`     // 返回错误信息
	UserIDs   []int    `json:"userIDs"`   //保存用户ID的切片
	UserNames []string `json:"userNames"` //保存用户名的切片
	User
}

type RegisterMes struct {
	User User `json:"user"` //类型就是User结构体
}

type RegisterResMes struct {
	Code  int    `json:"code"`  // 返回状态码 400 表示该用户已经占用 200 表示注册成功
	Error string `json:"error"` // 返回错误信息
}

// NotifyUserStatusMes 为了配合服务器端推送用户上下线变化消息，新定义一个类型
type NotifyUserStatusMes struct {
	UserID int `json:"userID"`
	Status int `json:"status"`
}

// GroupMes 广播：客户端发送的消息 由于返回的消息也只需要包含内容+发送者，因此无需再定义一个GroupResMes
type GroupMes struct {
	Content string `json:"content"` //消息内容
	Sender  User   //匿名结构体继承。复用User，注意不带密码
}

// PrivateMes 私聊：客户端发送的消息
type PrivateMes struct {
	Content    string `json:"content"` //消息内容
	Sender     User   //发送人
	ReceiverID int    `json:"receiverID"` //接收人的ID
}

// PrivateResMes 私聊：服务端返回给客户端的消息
type PrivateResMes struct {
	Content string `json:"content"`
	Sender  User
}
