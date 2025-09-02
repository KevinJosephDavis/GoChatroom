// Package model contains data structures and error definitions for the chatroom server.
package model

import (
	"errors"
)

//根据业务逻辑的需要，自定义一些错误

var (
	ErrUserNotExists = errors.New("用户不存在")
	ErrUserExists    = errors.New("用户已存在")
	ErrUserPwd       = errors.New("密码不正确")
)
