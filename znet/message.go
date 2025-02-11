package znet

type Message struct {
	Id      uint32 //消息的ID
	DataLen uint32 //消息的长度
	Data    []byte //消息的内容
}

func NewMessage(id uint32, data []byte) *Message {
	return &Message{
		Id:      id,
		DataLen: uint32(len(data)),
		Data:    data,
	}
}

func (m *Message) GetDataLen() uint32 { //获取消息的长度
	return m.DataLen
}

func (m *Message) GetMsgId() uint32 { //获取消息ID
	return m.Id
}

func (m *Message) GetData() []byte { // 获取消息内容同
	return m.Data
}

func (m *Message) SetMsgId(id uint32) { //设计消息ID
	m.Id = id
}

func (m *Message) SetData(data []byte) { //设计消息内容
	m.Data = data
}
func (m *Message) SetDataLen(dataLen uint32) { //设置消息数据段长度
	m.DataLen = dataLen
}
