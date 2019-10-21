package duo

import (
	"github.com/sereiner/duo/component"
	"github.com/sereiner/duo/engine"
	"github.com/sereiner/library/log"
	"os"
	"os/signal"
	"syscall"
)

type DuoApp struct {
	*option
	engine *engine.ServiceEngine
	log    log.ILogging
}

func NewDouApp(option ...Option) *DuoApp {
	app := &DuoApp{
		engine: engine.NewServiceEngine(),
		log:    log.New("duo"),
	}

	for _, op := range option {
		op(app.option)
	}

	return app
}

func (d *DuoApp) Init(i func(c component.IContainer)) {
	d.engine.Install(i)
}

func (d *DuoApp) Run() error {

	d.engine.Start()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)

	<-ch
	return nil
}
