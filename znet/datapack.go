package znet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"zinx/utils"
	"zinx/ziface"
)

type DataPack struct {
}

func NewDataPack() *DataPack {
	return &DataPack{}
}

func (dp *DataPack) GetHeadLen() uint32 { //获取包头长度
	return 8
}
func (dp *DataPack) Pack(msg ziface.IMessage) ([]byte, error) { //封包方法

	//创建一个存放byte字节流的缓冲区
	dataBuf := bytes.NewBuffer([]byte{})

	//写dataLen
	if err := binary.Write(dataBuf, binary.LittleEndian, msg.GetDataLen()); err != nil {
		return nil, err
	}
	//写msgId
	if err := binary.Write(dataBuf, binary.LittleEndian, msg.GetMsgId()); err != nil {
		return nil, err
	}

	//写data数据
	if err := binary.Write(dataBuf, binary.LittleEndian, msg.GetData()); err != nil {
		return nil, err
	}
	return dataBuf.Bytes(), nil
}

func (dp *DataPack) Unpack(bData []byte) (ziface.IMessage, error) { //拆包方法
	//创建一个输入二进制数据的ioReader
	dataBuf := bytes.NewBuffer(bData)

	//只解压head的信息，得到dataLen和msgID
	msg := &Message{}

	//读dataLen
	if err := binary.Read(dataBuf, binary.LittleEndian, &msg.DataLen); err != nil {
		return nil, err
	}

	//读msgID
	if err := binary.Read(dataBuf, binary.LittleEndian, &msg.Id); err != nil {
		return nil, err
	}

	//判断dataLen的长度是否超出允许的最大包长度
	if utils.GlobalObject.MaxPacketSize > 0 && msg.DataLen > utils.GlobalObject.MaxPacketSize {
		return nil, errors.New("Too long msg data received")
	}

	//这里只需要把head的数据拆包出来就可以了，然后通过head的长度再从conn中读取一次数据
	return msg, nil
}
