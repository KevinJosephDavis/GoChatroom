package model

// User 定义一个用户的结构体
type User struct {
	UserID     int    `json:"userId"`
	UserPwd    string `json:"userPwd"`
	UserName   string `json:"userName"`
	UserStatus int    `json:"userStatus"` //用户状态
}
