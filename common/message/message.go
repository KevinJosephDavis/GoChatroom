// Package message provides shared types and structures for the chatroom server.
package message

//命名还不够清晰。找个时间修改一下
const (
	LoginMesType            = "LoginMes"
	LoginResMesType         = "LoginResMes"
	RegisterMesType         = "RegisterMes"
	RegisterResMesType      = "RegisterResMes"
	NotifyUserStatusMesType = "NotifyUserStatusMes"
	SmsMesType              = "SmsMes"
	SmsPrivateMesType       = "SmsPrivateMes"
	SmsPrivateResMesType    = "SmsPrivateResMes"
	OfflineMesType          = "OfflineMes"
	OfflineResMesType       = "OfflineResMes"
	OnlineMesType           = "OnlineMes"
	OnlineResMesType        = "OnlineResMes"
	DeleteAccountMesType    = "DeleteAccountMes"
	DeleteAccountResMesType = "DeleteAccountResMes"
)

//定义几个用户状态的常量
const (
	UserOnline = iota
	UserOffline
	UserBusyStatus
)

//定义两个退出情况
const (
	Abnormal = "非正常退出"
	Normal   = "正常退出"
)

type Message struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

//LoginMes 登录：用户传给服务端的消息
type LoginMes struct {
	UserID       int    `json:"userID"`
	UserPassword string `json:"userPassword"`
	UserName     string `json:"userName"`
}

type LoginResMes struct {
	Code      int      `json:"code"` // 返回状态码 500 表示该用户未注册 200 表示登陆成功
	Error     string   `json:"error"`
	UserIDs   []int    `json:"userIDs"`
	UserNames []string `json:"userNames"`
	User
}

type RegisterMes struct {
	User User `json:"user"`
}

type RegisterResMes struct {
	Code  int    `json:"code"` // 返回状态码 400 表示该用户已经占用 200 表示注册成功
	Error string `json:"error"`
}

// NotifyUserStatusMes 提醒其它用户：有用户的在线状态发生了变化
type NotifyUserStatusMes struct {
	UserID   int    `json:"userID"`
	UserName string `json:"userName"`
	Status   int    `json:"status"`
}

// GroupMes 广播：客户端发送的消息 由于返回的消息也只需要包含内容+发送者，因此无需再定义一个GroupResMes
type GroupMes struct {
	Content string `json:"content"`
	Sender  User   `json:"sender"`
}

// PrivateMes 私聊：客户端发送的消息
type PrivateMes struct {
	Content    string `json:"content"`
	Sender     User   `json:"sender"`
	ReceiverID int    `json:"receiverID"`
}

// PrivateResMes 私聊：服务端返回给客户端的消息
type PrivateResMes struct {
	Content string `json:"content"`
	Sender  User   `json:"sender"`
}

// OfflineMes 下线：客户端传给服务端的信息
type OfflineMes struct {
	UserID   int    `json:"userID"`
	UserName string `json:"userName"`
	Reason   string `json:"reason"`
	Time     int64  `json:"time"`
}

// OfflineResMes 下线：服务端传给在线用户的消息
type OfflineResMes struct {
	UserID   int    `json:"userID"`
	UserName string `json:"userName"`
	Reason   string `json:"reason"`
	Time     int64  `json:"time"`
}

// OnlineMes 上线：客户端传给服务端的信息
type OnlineMes struct {
	UserID   int    `json:"userID"`
	UserName string `json:"userName"`
}

// OnlineResMes 上线：服务端传给其它在线用户的信息
type OnlineResMes struct {
	UserID   int    `json:"userID"`
	UserName string `json:"userName"`
}

// DeleteAccountMes 注销：客户端传给服务端的信息
type DeleteAccountMes struct {
	User User  `json:"user"`
	Time int64 `json:"time"`
}

// DeleteAccountResMes 注销：服务端传给其它在线用户的信息
type DeleteAccountResMes struct {
	User User  `json:"user"`
	Time int64 `json:"time"`
}
