package process2

import "fmt"

//因为UserMgr实例在服务器端有且仅有一个
//而在很多地方都会使用到，所以我们将其定义为全局变量

var (
	userMgr *UserMgr
)

type UserMgr struct {
	onlineUsers map[int]*UserProcess0
}

//完成对userMgr的初始化工作
func init() {
	userMgr = &UserMgr{
		onlineUsers: make(map[int]*UserProcess0, 1024),
	}
}

// AddOnlineUser 完成对onlineUsers的添加
func (usmng *UserMgr) AddOnlineUser(up *UserProcess0) {
	usmng.onlineUsers[up.UserID] = up
}

// DeleteOnlineUser 删除
func (usmng *UserMgr) DeleteOnlineUser(userID int) {
	delete(usmng.onlineUsers, userID)
}

// GetAllOnlineUsers 返回所有当前在线的用户
func (usmng *UserMgr) GetAllOnlineUsers() map[int]*UserProcess0 {
	return usmng.onlineUsers
}

// GetOnlineUserByID 根据ID返回对应的值
func (usmng *UserMgr) GetOnlineUserByID(userID int) (up *UserProcess0, err error) {
	up, ok := usmng.onlineUsers[userID]
	if !ok {
		//说明当前查找的用户当前不在线
		err = fmt.Errorf("用户%d当前不在线或不存在", userID)
		return
	}
	return
}
