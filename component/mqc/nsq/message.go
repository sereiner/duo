package nsq

type NsqMessage struct {
	Message string
	HasData bool
}

func (m *NsqMessage) Ack() error {
	return nil
}

func (m *NsqMessage) Nack() error {
	return nil
}

func (m *NsqMessage) GetMessage() string {
	return m.Message
}

func (m *NsqMessage) Has() bool {
	return m.HasData
}
