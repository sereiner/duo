package mqc

type IMessage interface {
	Ack() error
	Nack() error
	GetMessage() string
}
