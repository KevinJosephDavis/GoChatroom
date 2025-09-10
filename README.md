# 分布式聊天系统

## 项目简介

一个基于Go语言开发的分布式实时聊天系统，支持私聊、群聊、离线消息、用户状态管理等功能。

## :star2: 核心功能

### :dart: 基础功能

:white_check_mark:用户注册与登录

:white_check_mark:实时私聊与群聊（广播）

:white_check_mark:在线用户列表展示

:white_check_mark:用户上下线状态通知

### :rocket:高级功能

:white_check_mark:离线消息存储与推送

:white_check_mark:心跳检测与异常断开处理

:white_check_mark:跨网络远程通信



## :hammer_and_wrench:技术栈

### 后端技术

| 技术         | 用途         |
| ------------ | ------------ |
| TCP/IP       | 网络通讯协议 |
| JSON         | 消息序列化   |
| Redis        | 连接池/缓存  |
| sync.Map     | 并发消息存储 |
| sync.RWMutex | 读写锁       |



### 架构特性

:white_check_mark:C/S架构：客户端-服务器模式

:white_check_mark:并发处理：Goroutine处理多连接

:white_check_mark:状态同步：实时用户状态管理

:white_check_mark:连接池优化：Redis连接复用

:white_check_mark:锁机制：保证数据一致性



## :building_construction:系统架构

### :desktop_computer:组件图

```mermaid```

```mermaid
graph TD
    subgraph Client
    	SmsProcess[消息处理器]
    	UserMgr[用户管理器]
    	U1[接收用户的请求]
    	U2[接收、打印服务端返回消息]
    	U3[发送心跳检测结果]
    	U4[维护在线用户map]
    end
	
	SmsProcess --> U1
	SmsProcess --> U2
	UserMgr --> U3
	UserMgr --> U4
```

```mermaid```



```mermaid
graph TD

subgraph Server
        SmsProcess[消息处理器]
        U1[接收并处理客户端消息]
    	Redis[Redis连接池]
    	U2[数据库CRUD]
		UserMgr[用户管理器]
		U3[维护在线用户和所有用户map]
    end
    
    SmsProcess --> U1
    Redis --> U2
    UserMgr --> U3
```

```mermaid
graph TD
	subgraph trans[传输]
		transins[transfer实例]
		U1[获取客户端与服务端的连接]
		U2[读、写消息]
	end
	
	transins --> U1
	transins --> U2
```

### 消息流程

```mermaid
graph 
	subgraph 注册
		s1[用户选择注册]
		s2[客户端发送注册请求]
		s3[服务端接收并在map中增加用户]
		s4[redis增加用户]
		s5[服务端返回信息]
		s6[客户端接收信息]
	end
	
	s1 --> s2
	s2 --> s3
	s3 --> s4
	s4 --> s5
	s5 --> s6
```

```mermaid
graph
	subgraph 登录
		s1[用户选择登录]
		s2[客户端发送登录请求]
		s3[服务端接收]
		s4[redis验证]
		s5[服务端返回信息并更新map]
		s6[客户端接收信息并更新map]
		s7[客户端更新并显示在线用户列表]
	end
	
	s1 --> s2
	s2 --> s3
	s3 --> s4
	s4 --> s5
	s5 --> s6
	s6 --> s7
```

广播、私聊、下线功能同理。



## :bulb:优化

:one:系统分配ID，而不是用户自己选择ID



## :handshake:参与贡献

欢迎提交Issue和Pull Request !