package duo

import (
	"github.com/sereiner/duo/conf"
	"os"

	"github.com/sereiner/duo/component"

	logger "github.com/sereiner/library/log"
	"github.com/wule61/log"
)

type MicroApp struct {
	logger  *logger.Logger
	xlogger logger.ILogging
	*option
	component.IComponentRegistry
	duo *Duo
}

func NewMicroApp(ops ...Option) *MicroApp {

	app := &MicroApp{
		option:             &option{},
		IComponentRegistry: component.NewServiceRegistry(),
	}
	logging := log.New(os.Stdout, "", log.Llongcolor)
	logging.SetOutputLevel(log.Ldebug)
	app.xlogger = logging

	for _, opt := range ops {
		opt(app.option)
	}

	app.logger = logger.GetSession("parrot", logger.CreateSession())

	return app
}

//Start 启动服务器
func (m *MicroApp) Start() {

	defer logger.Close()

	serverConf := conf.NewServerConf(m.ConfPath)

	if serverConf.IsDebug() {
		m.PlatName += "_debug"
	}

	m.duo = NewDuo(m.Name, m.logger, m.PlatName, m.SystemName, m.ClusterName, m.Trace, serverConf)

	s, err := m.duo.Start(m.GetServices())
	if err != nil {
		m.xlogger.Error(err)
	}

	m.xlogger.Info(s)
}
