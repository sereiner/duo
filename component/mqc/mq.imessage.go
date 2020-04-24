package mqc

type IMessage interface {
	Ack() error
	Nack() error
	Has() bool
	GetMessage() string
}
