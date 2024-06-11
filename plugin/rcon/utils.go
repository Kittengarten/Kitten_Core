package rcon

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"time"
)

// 来自 github.com/bearbin/mcgorcon

const (
	BadAuth        = -1
	PayloadMaxSize = 1460
)

const (
	PacketResponse = iota
	_
	PacketCommand
	PacketLogin
)

type MCConn struct {
	conn     net.Conn
	password string
}

type packetType int32

type RCONHeader struct {
	Size      int32
	RequestID int32
	Type      packetType
}

func (c *MCConn) Open(addr, password string) error {
	conn, err := net.DialTimeout(`tcp`, addr, 10*time.Second)
	if nil != err {
		return err
	}
	*c = MCConn{
		conn:     conn,
		password: password,
	}
	return nil
}

func (c *MCConn) Close() error {
	return c.conn.Close()
}

// SendCommand 向服务器发送命令并返回结果
func (c *MCConn) SendCommand(command string) (string, error) {
	// 发送包
	if PayloadMaxSize < len(command) {
		return ``, errors.New(`命令过长喵！`)
	}
	head, payload, err := c.sendPacket(PacketCommand, []byte(command))
	if nil != err {
		return ``, err
	}
	// 验证失败，返回错误
	if head.RequestID == BadAuth {
		return ``, errors.New(`验证失败，不能发送命令喵！`)
	}
	return string(payload), nil
}

// Authenticate 验证用户身份
func (c *MCConn) Authenticate() error {
	// 发送包
	head, _, err := c.sendPacket(PacketLogin, []byte(c.password))
	if nil != err {
		return err
	}
	// 验证失败，返回错误
	if head.RequestID == BadAuth {
		return errors.New(`验证失败喵！`)
	}
	return nil
}

// sendPacket 发送二进制包并返回响应
func (c *MCConn) sendPacket(t packetType, p []byte) (RCONHeader, []byte, error) {
	// 生成二进制包
	packet, err := packetise(t, p)
	if nil != err {
		return RCONHeader{}, nil, err
	}
	// 发送二进制包
	_, err = c.conn.Write(packet)
	if nil != err {
		return RCONHeader{}, nil, err
	}
	// 接收并解码响应
	return depacketise(c.conn)
}

// packetise 编码数据包并转换为二进制表达
func packetise(t packetType, p []byte) ([]byte, error) {
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, int32(len(p)+10))
	binary.Write(&buf, binary.LittleEndian, int32(0))
	binary.Write(&buf, binary.LittleEndian, t)
	binary.Write(&buf, binary.LittleEndian, p)
	binary.Write(&buf, binary.LittleEndian, [2]byte{})
	// 数据包太大，无法处理
	if PayloadMaxSize <= buf.Len() {
		return nil, errors.New(`数据包太大了喵！`)
	}
	// 返回数据包的字节切片
	return buf.Bytes(), nil
}

// depacketise 解码数据包
func depacketise(r io.Reader) (RCONHeader, []byte, error) {
	head := RCONHeader{}
	if err := binary.Read(r, binary.LittleEndian, &head); nil != err {
		return RCONHeader{}, nil, err
	}
	payload := make([]byte, head.Size-8)
	if _, err := io.ReadFull(r, payload); nil != err {
		return RCONHeader{}, nil, err
	}
	// 检查
	switch head.Type {
	case PacketResponse, PacketCommand:
		return head, payload[:len(payload)-2], nil
	default:
		return RCONHeader{}, nil, errors.New(`数据包类型错误喵！`)
	}
}
