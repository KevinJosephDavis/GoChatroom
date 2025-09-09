package process2

import (
	"fmt"
	"sync"
	"time"

	"github.com/kevinjosephdavis/chatroom/common/message"
)

//因为UserMgr实例在服务器端有且仅有一个
//而在很多地方都会使用到，所以我们将其定义为全局变量

var (
	userMgr *UserMgr
)

type UserMgr struct {
	OnlineUsers sync.Map //map[int]*UserProcess0
	userStatus  sync.Map //维护用户状态。key为ID，value为几个用户状态。事实上维护了所有用户
	offlineMes  sync.Map //每个用户的离线消息队列
}

// InitUserMgr 完成对userMgr的初始化工作
func InitUserMgr() {
	userMgr = &UserMgr{
		OnlineUsers: sync.Map{},
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
	if up.LastHeartBeat.IsZero() {
		up.LastHeartBeat = time.Now()
	}
	usmng.OnlineUsers.Store(up.UserID, up)
}

// DeleteOnlineUser 删除下线用户
func (usmng *UserMgr) DeleteOnlineUser(userID int) {
	usmng.OnlineUsers.Delete(userID)
}

// GetAllOnlineUsers 返回所有当前在线的用户的控制器
func (usmng *UserMgr) GetAllOnlineUsers() []*UserProcess0 {
	var onlineUsers []*UserProcess0

	usmng.OnlineUsers.Range(func(key, value interface{}) bool {
		onlineUsers = append(onlineUsers, value.(*UserProcess0))
		return true
	})

	return onlineUsers
}

// GetOnlineUserByID 根据ID返回对应的值
func (usmng *UserMgr) GetOnlineUserByID(userID int) (up *UserProcess0, err error) {
	value, exist := usmng.OnlineUsers.Load(userID)
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

// StoreOfflineMes 存储离线留言
func (usmng *UserMgr) StoreOfflineMes(offlineMes *message.OfflineMes) {
	var mes []*message.OfflineMes
	existingMes, exist := usmng.offlineMes.Load(offlineMes.ReceiverID) //看这个接收者是否已经存在离线留言
	if exist {
		//类型断言。因为Load出来的value是interface{}类型，不能直接append
		if existingSlice, ok := existingMes.([]*message.OfflineMes); ok {
			mes = existingSlice
		} else {
			fmt.Println("StoreOfflineMes 类型断言错误")
			mes = []*message.OfflineMes{}
		}
	}
	mes = append(mes, offlineMes)
	usmng.offlineMes.Store(offlineMes.ReceiverID, mes)
}

// GetOfflineMes 根据用户ID获取离线留言
func (usmng *UserMgr) GetOfflineMes(userID int) []*message.OfflineMes {
	value, exist := usmng.offlineMes.Load(userID)
	if !exist {
		//如果用户没有收到离线留言
		return []*message.OfflineMes{}
	}
	//如果有，就要类型断言
	if mes, ok := value.([]*message.OfflineMes); ok {
		return mes
	}

	fmt.Println("GetOfflineMes 类型断言错误")
	return []*message.OfflineMes{}
}

// ClearOfflineMes 用户读取了离线留言后，要清空离线留言列表
func (usmng *UserMgr) ClearOfflineMes(userID int) {
	usmng.offlineMes.Store(userID, []*message.OfflineMes{}) //不清空键，只清空值
}

// StartHeartBeatCheck 开始心跳检测
func (usmng *UserMgr) StartHeartBeatCheck() {
	go func() {
		ticker := time.NewTicker(5 * time.Second) //5秒检查一次
		defer ticker.Stop()

		for range ticker.C {
			usmng.CheckAllUserHeartBeat()
		}

	}()
}

// CheckAllUserHeartBeat UserManager检查所有用户的心跳
func (usmng *UserMgr) CheckAllUserHeartBeat() {
	usmng.OnlineUsers.Range(func(key, value interface{}) bool {
		userID := key.(int)
		uspc := value.(*UserProcess0)

		//如果15秒（考虑到网络波动）内没有心跳，判定为非正常下线（网络断开、关闭终端...）
		if time.Since(uspc.LastHeartBeat) > 15*time.Second {
			fmt.Printf("用户%s (ID:%d)心跳超时，最后心跳时间为：%v，判定为非正常下线\n", uspc.UserName,
				userID, uspc.LastHeartBeat.Format("2006-01-02 15:04:05"))

			//将该用户从服务端维护的onlineUsers中删除，并更改其在userStatus中的状态
			usmng.DeleteOnlineUser(userID)
			usmng.userStatus.Store(userID, message.UserOffline)

			//发送消息
			smsp := &SmsProcess{}
			smsp.SendAbnormalLogoutResMes(userID, uspc.UserName)
		}
		return true
	})
}
