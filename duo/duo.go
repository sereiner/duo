package duo

import (
	"fmt"
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
	isDebug     bool
	platName    string
	systemName  string
	clusterName string
	mu          sync.Mutex
	trace       string
	server      *server.ServerEngine
	done        bool
}

func NewDuo(appName string, logger *logger.Logger, isDebug bool, platName string, systemName string, clusterName string, trace string) *Duo {
	return &Duo{
		appName:     appName,
		logger:      logger,
		closeChan:   make(chan struct{}),
		interrupt:   make(chan os.Signal, 1),
		isDebug:     isDebug,
		platName:    platName,
		systemName:  systemName,
		clusterName: clusterName,
		trace:       trace,
	}
}

func (d *Duo) Start(f func(server *grpc.Server)) (s string, err error) {
	//非调试模式时设置日志写协程数为50个
	if !d.isDebug {
		logger.AddWriteThread(49)
	}

	reporter := zipkinhttp.NewReporter("http://127.0.0.1:9411/api/v2/spans")
	defer reporter.Close()

	// create our local service endpoint
	endpoint, err := zipkin.NewEndpoint("myService", "myservice.mydomain.com:80")
	if err != nil {
		d.logger.Fatalf("unable to create local endpoint: %+v\n", err)
	}

	// initialize our tracer
	nativeTracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))
	if err != nil {
		d.logger.Fatalf("unable to create tracer: %+v\n", err)
	}

	tracer := zipkinot.Wrap(nativeTracer)

	// optionally set as Global OpenTracing tracer instance
	opentracing.SetGlobalTracer(tracer)

	d.server = server.NewServiceEngine(d.appName, ":8090", server.WithServerOption([]grpc.ServerOption{
		grpc_middleware.WithUnaryServerChain(
			server.RecoveryInterceptor,
			otgrpc.OpenTracingServerInterceptor(tracer, otgrpc.LogPayloads()),
			server.GetClientIP,
			server.LoggingInterceptor,
		),
	}))

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
	d.server.Shutdown(time.Second * 1)
	return fmt.Sprintf("%s 已安全退出", d.appName), nil
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
