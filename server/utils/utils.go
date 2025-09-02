// Package utils 放常用函数、结构体
package utils

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"

	"github.com/kevinjosephdavis/chatroom/common/message"
)

// Transfer 将这些方法关联到结构体中
type Transfer struct {
	//分析应该有的字段
	Conn net.Conn   //连接
	Buf  [8096]byte //传输时使用的缓冲

}

// ReadPkg 读取数据
func (tf *Transfer) ReadPkg() (mes message.Message, err error) {
	//buf := make([]byte, 8096)
	fmt.Println("读取客户端发送的数据...")
	_, err = tf.Conn.Read(tf.Buf[:4]) //虽然buf是在readPkg中临时创建的，但conn.Read()会阻塞等待客户端发送数据
	//客户端发送数据后，Read()将数据写入buf的前4个字节
	if err != nil {
		//err = errors.New("read Pkg header error")
		return
	}

	//为什么要先读取4个字节，而不直接读取pkgLen个字节？
	//因为这4个字节告诉了我们数据量的大小，根据数据量的大小我们才能准确地截取有效的数据，不会截取过多或过少
	//TCP是流式协议，没有消息边界，不知道要读到哪里，所以要先发一个长度前缀。读太多可能把下一条的消息都走了，太少无需多言

	pkgLen := binary.BigEndian.Uint32(tf.Buf[0:4]) //把这个字节序列转换成uint32，这样我们才能用pkgLen这个表示长度的量

	//根据pkgLen读取消息内容
	n, err := tf.Conn.Read(tf.Buf[:pkgLen])
	if n != int(pkgLen) || err != nil {
		//err = errors.New("read Pkg body error")
		return
	}

	//把pkgLen反序列化成message.Message
	//把pkgLen反序列化，才能将其从字节流转为可用的结构体
	err = json.Unmarshal(tf.Buf[:pkgLen], &mes) //注意&
	if err != nil {
		fmt.Println("json.Unmarshall err=", err)
		return
	}

	//思维模型：
	//网络字节流（原始数据）->长度前缀协议（定义消息边界）->按长度获取（准确获取完整信息）->反序列化得到可操作对象（转为可操作结构）
	return
}

// WritePkg 发送数据
func (tf *Transfer) WritePkg(data []byte) (err error) {
	//先发送一个长度给对方
	//复用login.go中发送长度的代码
	// 先获取到data的长度->转成一个表示长度的byte切片
	pkgLen := uint32(len(data))
	//var buf [4]byte
	binary.BigEndian.PutUint32(tf.Buf[0:4], pkgLen)
	//现在发送消息长度
	n, err := tf.Conn.Write(tf.Buf[:4])
	if n != 4 || err != nil {
		fmt.Println("conn.Write(bytes) err=", err)
		return
	}

	//发送data本身
	n, err = tf.Conn.Write(data)
	if n != 4 || err != nil {
		fmt.Println("conn.Write(bytes) err=", err)
		return
	}
	return
}
