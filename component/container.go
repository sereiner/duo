package component

type IContainer interface {
}


type Container struct {

}


func New() IContainer {
	return &Container{}
}