package mqc

type MqcServer interface {
	Consume() (err error)
	ShutDown()
}