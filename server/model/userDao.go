package model

import (
	"encoding/json"
	"fmt"

	"github.com/gomodule/redigo/redis"
	"github.com/kevinjosephdavis/chatroom/common/message"
)

// 在服务器启动后就初始化一个UserDao实例，把它做成全局的变量。在需要和redis操作时，直接使用即可
var (
	MyUserDao *UserDao
)

type UserDao struct {
	pool *redis.Pool
}

// NewUserDao 使用工厂模式创建一个UserDao实例
func NewUserDao(pool *redis.Pool) (userDao *UserDao) {
	userDao = &UserDao{
		pool: pool,
	}
	return
}

// 1.根据用户ID返回User实例+err
func (usd *UserDao) getUserByID(conn redis.Conn, id int) (user *User, err error) {
	//通过给定的id去redis查询该用户
	res, err := redis.String(conn.Do("HGet", "users", id))
	if err != nil {
		if err == redis.ErrNil {
			//在users这个哈希中没有找到对应的id
			err = ErrUserNotExists
		}
		return
	}

	user = &User{}

	//把res反序列化成users实例
	err = json.Unmarshal([]byte(res), user)
	if err != nil {
		fmt.Println("json.Unmarshall err=", err)
	}

	return
}

// Login 完成对用户的验证
func (usd *UserDao) Login(userID int, userPwd string) (user *User, err error) {
	//先从UserDao的连接池中取出一根连接
	conn := usd.pool.Get()
	defer conn.Close()
	user, err = usd.getUserByID(conn, userID)
	if err != nil {
		return
	}

	//此时至少证明用户获取到了，但密码未必正确
	if user.UserPwd != userPwd {
		err = ErrUserPwd
		return
	}
	return
}

// Register 完成用户注册
func (usd *UserDao) Register(user *message.User) (err error) {
	//先从UserDao的连接池中取出一根连接
	conn := usd.pool.Get()
	defer conn.Close()
	_, err = usd.getUserByID(conn, user.UserID)
	if err == nil {
		err = ErrUserExists
		return
	}

	//此时说明该ID还不存在Redis中，则可以完成注册
	data, err := json.Marshal((user))
	if err != nil {
		return
	}

	//存入数据库
	_, err = conn.Do("HSet", "users", user.UserID, string(data))
	if err != nil {
		fmt.Println("保存注册用户出错 err=", err)
		return
	}
	return
}

// DeleteAccount 完成用户注销
func (usd *UserDao) DeleteAccount(user *message.User) (err error) {
	conn := usd.pool.Get()
	defer conn.Close()

	_, err = usd.getUserByID(conn, user.UserID)
	if err != nil {
		fmt.Println("用户不存在，无法删除 err=", err)
		return
	}

	_, err = conn.Do("HDEL", "users", user.UserID)
	if err != nil {
		fmt.Println("从数据库中删除用户出错 err=", err)
		return
	}
	fmt.Printf("用户%d已从数据库销户", user.UserID)
	return
}
