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
	userStatus  sync.Map //维护用户状态。key为ID，value为几个用户状态。事实上维护了所有用户
	offlineMes  sync.Map //每个用户的离线消息队列
}

// InitUserMgr 完成对userMgr的初始化工作
func InitUserMgr() {
	userMgr = &UserMgr{
		onlineUsers: sync.Map{},
		userStatus:  sync.Map{},
		offlineMes:  sync.Map{},
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

// AddOnlineUser 添加上线用户
func (usmng *UserMgr) AddOnlineUser(up *UserProcess0) {
	usmng.onlineUsers.Store(up.UserID, up)
}

// DeleteOnlineUser 删除下线用户
func (usmng *UserMgr) DeleteOnlineUser(userID int) {
	usmng.onlineUsers.Delete(userID)
}

// GetAllOnlineUsers 返回所有当前在线的用户的控制器
func (usmng *UserMgr) GetAllOnlineUsers() []*UserProcess0 {
	var onlineUsers []*UserProcess0

	usmng.onlineUsers.Range(func(key, value interface{}) bool {
		onlineUsers = append(onlineUsers, value.(*UserProcess0))
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

	up, assertOk := value.(*UserProcess0)
	if !assertOk {
		err = fmt.Errorf("用户%d 数据格式错误", userID)
		return
	}
	return
}

// DeleteExistUser 将注销用户删除
func (usmng *UserMgr) DeleteExistUser(userID int) {
	usmng.userStatus.Delete(userID)
}

//需不需要写一个AddNewUser？看一下注册后登录的逻辑

// SetUserStatus 设置用户状态（实际上就是AddNewUser）
func (usmng *UserMgr) SetUserStatus(userID int, status int) {
	usmng.userStatus.Store(userID, status)
}

// GetUserStatus 获取用户状态
func (usmng *UserMgr) GetUserStatus(userID int) (int, error) {
	value, exist := usmng.userStatus.Load(userID)
	if !exist {
		return -1, fmt.Errorf("用户不存在")
	}
	status, ok := value.(int)
	if !ok {
		return -2, fmt.Errorf("GetUserStatus 类型断言错误")
	}
	return status, nil
}
