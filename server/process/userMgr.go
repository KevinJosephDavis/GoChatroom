package process2

import (
	"fmt"
	"sync"
)

//因为UserMgr实例在服务器端有且仅有一个
//而在很多地方都会使用到，所以我们将其定义为全局变量

var (
	userMgr *UserMgr
)

type UserMgr struct {
	onlineUsers sync.Map //map[int]*UserProcess0
	userStatus  sync.Map //维护用户状态。key为ID，value为几个用户状态
}

// InitUserMgr 完成对userMgr的初始化工作
func InitUserMgr() {
	userMgr = &UserMgr{
		onlineUsers: sync.Map{},
		userStatus:  sync.Map{},
	}
	fmt.Println("userMgr初始化完成")
}

// GetUserMgr 获取实例（单例模式）
func GetUserMgr() *UserMgr {
	if userMgr == nil {
		InitUserMgr()
	}
	return userMgr
}

// AddOnlineUser 完成对onlineUsers的添加
func (usmng *UserMgr) AddOnlineUser(up *UserProcess0) {
	//usmng.onlineUsers[up.UserID] = up
	usmng.onlineUsers.Store(up.UserID, up)
}

// DeleteOnlineUser 删除
func (usmng *UserMgr) DeleteOnlineUser(userID int) {
	//delete(usmng.onlineUsers, userID)
	usmng.onlineUsers.Delete(userID)
}

// GetAllOnlineUsers 返回所有当前在线的用户的控制器
func (usmng *UserMgr) GetAllOnlineUsers() []*UserProcess0 {
	var onlineUsers []*UserProcess0

	usmng.onlineUsers.Range(func(key, value interface{}) bool {
		onlineUsers = append(onlineUsers, value.(*UserProcess0)) // 类型断言
		return true
	})

	return onlineUsers
}

// GetOnlineUserByID 根据ID返回对应的值
func (usmng *UserMgr) GetOnlineUserByID(userID int) (up *UserProcess0, err error) {
	value, exist := usmng.onlineUsers.Load(userID)
	if !exist {
		err = fmt.Errorf("用户%d 不在线或不存在", userID)
		return
	}

	//类型断言：将interface{}转换为*UserProcess0
	up, assertOk := value.(*UserProcess0)
	if !assertOk {
		err = fmt.Errorf("用户%d 数据格式错误", userID)
		return
	}
	return
}
