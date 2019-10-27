package duo

import (
	"fmt"
	"github.com/sereiner/duo/conf"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"

	"google.golang.org/grpc"

	"github.com/sereiner/duo/server"

	logger "github.com/sereiner/library/log"
)

type Duo struct {
	appName     string
	logger      *logger.Logger
	closeChan   chan struct{}
	interrupt   chan os.Signal
	platName    string
	systemName  string
	clusterName string
	mu          sync.Mutex
	trace       string
	server      *server.ServerEngine
	done        bool
	conf.IServerConf
}

func NewDuo(appName string, logger *logger.Logger, platName string, systemName string, clusterName string, trace string, conf conf.IServerConf) *Duo {
	return &Duo{
		appName:     appName,
		logger:      logger,
		closeChan:   make(chan struct{}),
		interrupt:   make(chan os.Signal, 1),
		platName:    platName,
		systemName:  systemName,
		clusterName: clusterName,
		trace:       trace,
		IServerConf: conf,
	}
}

func (d *Duo) Start(f func(server *grpc.Server)) (s string, err error) {

	if !d.IsDebug() {
		logger.AddWriteThread(49)
	}

	filter := grpc_middleware.WithUnaryServerChain(
		server.RecoveryInterceptor,
		//otgrpc.OpenTracingServerInterceptor(tracer, otgrpc.LogPayloads()),
		server.GetClientIP,
		server.LoggingInterceptor,
	)

	var tracer opentracing.Tracer
	if d.ZipKinEnable() {
		tracer, err = d.GetTracer(d.IServerConf)
		if err != nil {
			return "", err
		}

		filter = grpc_middleware.WithUnaryServerChain(
			server.RecoveryInterceptor,
			otgrpc.OpenTracingServerInterceptor(tracer, otgrpc.LogPayloads()),
			server.GetClientIP,
			server.LoggingInterceptor,
		)
	}

	d.server = server.NewServiceEngine(d.appName, d.IServerConf, server.WithServerOption([]grpc.ServerOption{filter}))

	d.server.ServiceFunc = f

	go d.server.Run()

	//定时释放内存
	go d.freeMemory()

	//堵塞当前进程，等服务结束
	signal.Notify(d.interrupt, os.Interrupt, os.Kill, syscall.SIGTERM) //, syscall.SIGUSR1) //9:kill/SIGKILL,15:SIGTEM,20,SIGTOP 2:interrupt/syscall.SIGINT
LOOP:
	for {
		select {
		case <-d.interrupt:
			d.done = true
			break LOOP
		}
	}
	d.logger.Infof("%s 正在退出...", d.appName)
	d.server.Shutdown(time.Millisecond * 500)
	return fmt.Sprintf("%s 已安全退出", d.appName), nil
}

func (d *Duo) GetTracer(conf conf.IServerConf) (tracer opentracing.Tracer, err error) {

	reporter := zipkinhttp.NewReporter(conf.GetZipKinReportURL())
	defer reporter.Close()

	endpoint, err := zipkin.NewEndpoint(conf.AppName(), "myservice.mydomain.com:80")
	if err != nil {
		return nil, fmt.Errorf("unable to create local endpoint: %+v\n", err)
	}

	nativeTracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))
	if err != nil {
		return nil, fmt.Errorf("unable to create tracer: %+v\n", err)
	}

	tracer = zipkinot.Wrap(nativeTracer)

	opentracing.SetGlobalTracer(tracer)

	return tracer, nil
}

func (d *Duo) freeMemory() {
	for {
		select {
		case <-d.closeChan:
			return
		case <-time.After(time.Second * 120):
			debug.FreeOSMemory()
		}
	}
}

func (d *Duo) Shutdown() {
	d.done = true
	close(d.closeChan)
	d.interrupt <- syscall.SIGTERM
}
