package mqc

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/judwhite/go-svc/svc"
	"github.com/mreiferson/go-options"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"
	"github.com/nsqio/nsq/internal/app"
	"github.com/micro-plat/lib4go/logger"

	"github.com/nsqio/nsq/nsqd"
)

type NsqdServer struct {
	once sync.Once
	nsqd *nsqd.NSQD
}

func NewNsqd() *NsqdServer {
	return &NsqdServer{}
}

func (p *NsqdServer) Init(env svc.Environment) error {
	if env.IsWindowsService() {
		dir := filepath.Dir(os.Args[0])
		return os.Chdir(dir)
	}
	return nil
}

func (p *NsqdServer) Start() error {
	opts := nsqd.NewOptions()

	// 解析命令行参数(待封装)
	flagSet := nsqdFlagSet(opts)
	flagSet.Parse(os.Args[1:])

	// 保证每次随机的值不一样
	rand.Seed(time.Now().UTC().UnixNano())

	if flagSet.Lookup("version").Value.(flag.Getter).Get().(bool) {
		fmt.Println(version.String("nsqd"))
		os.Exit(0)
	}

	// 解析配置文件（关于options待封装）
	var cfg config
	configFile := flagSet.Lookup("config").Value.String()
	if configFile != "" {
		_, err := toml.DecodeFile(configFile, &cfg)
		if err != nil {
			//logFatal("failed to load config file %s - %s", configFile, err)
		}
	}
	cfg.Validate()

	// 合并默认参数和命令行参数
	options.Resolve(opts, flagSet, cfg)
	nsqd, err := nsqd.New(opts)
	if err != nil {

	}
	p.nsqd = nsqd

	err = p.nsqd.LoadMetadata()
	if err != nil {
		//logFatal("failed to load metadata - %s", err)
	}
	err = p.nsqd.PersistMetadata()
	if err != nil {
		//logFatal("failed to persist metadata - %s", err)
	}

	// 开启服务
	go func() {
		err := p.nsqd.Main()
		if err != nil {
			p.Stop()
			os.Exit(1)
		}
	}()

	return nil
}

func (p *NsqdServer) Stop() error {
	p.once.Do(func() {
		p.nsqd.Exit()
	})
	return nil
}

//在研究这里怎么封装
func nsqdFlagSet(opts *nsqd.Options) *flag.FlagSet {
	flagSet := flag.NewFlagSet("nsqd", flag.ExitOnError)

	// basic options
	flagSet.Bool("version", false, "print version string")
	flagSet.String("config", "", "path to config file")

	logLevel := opts.LogLevel
	flagSet.Var(&logLevel, "log-level", "set log verbosity: debug, info, warn, error, or fatal")
	flagSet.String("log-prefix", "[nsqd] ", "log message prefix")
	flagSet.Bool("verbose", false, "[deprecated] has no effect, use --log-level")

	flagSet.Int64("node-id", opts.ID, "unique part for message IDs, (int) in range [0,1024) (default is hash of hostname)")
	flagSet.Bool("worker-id", false, "[deprecated] use --node-id")

	flagSet.String("https-address", opts.HTTPSAddress, "<addr>:<port> to listen on for HTTPS clients")
	flagSet.String("http-address", opts.HTTPAddress, "<addr>:<port> to listen on for HTTP clients")
	flagSet.String("tcp-address", opts.TCPAddress, "<addr>:<port> to listen on for TCP clients")
	authHTTPAddresses := app.StringArray{}
	flagSet.Var(&authHTTPAddresses, "auth-http-address", "<addr>:<port> to query auth server (may be given multiple times)")
	flagSet.String("broadcast-address", opts.BroadcastAddress, "address that will be registered with lookupd (defaults to the OS hostname)")
	lookupdTCPAddrs := app.StringArray{}
	flagSet.Var(&lookupdTCPAddrs, "lookupd-tcp-address", "lookupd TCP address (may be given multiple times)")
	flagSet.Duration("http-client-connect-timeout", opts.HTTPClientConnectTimeout, "timeout for HTTP connect")
	flagSet.Duration("http-client-request-timeout", opts.HTTPClientRequestTimeout, "timeout for HTTP request")

	// diskqueue options
	flagSet.String("data-path", opts.DataPath, "path to store disk-backed messages")
	flagSet.Int64("mem-queue-size", opts.MemQueueSize, "number of messages to keep in memory (per topic/channel)")
	flagSet.Int64("max-bytes-per-file", opts.MaxBytesPerFile, "number of bytes per diskqueue file before rolling")
	flagSet.Int64("sync-every", opts.SyncEvery, "number of messages per diskqueue fsync")
	flagSet.Duration("sync-timeout", opts.SyncTimeout, "duration of time per diskqueue fsync")

	flagSet.Int("queue-scan-worker-pool-max", opts.QueueScanWorkerPoolMax, "max concurrency for checking in-flight and deferred message timeouts")
	flagSet.Int("queue-scan-selection-count", opts.QueueScanSelectionCount, "number of channels to check per cycle (every 100ms) for in-flight and deferred timeouts")

	// msg and command options
	flagSet.Duration("msg-timeout", opts.MsgTimeout, "default duration to wait before auto-requeing a message")
	flagSet.Duration("max-msg-timeout", opts.MaxMsgTimeout, "maximum duration before a message will timeout")
	flagSet.Int64("max-msg-size", opts.MaxMsgSize, "maximum size of a single message in bytes")
	flagSet.Duration("max-req-timeout", opts.MaxReqTimeout, "maximum requeuing timeout for a message")
	flagSet.Int64("max-body-size", opts.MaxBodySize, "maximum size of a single command body")

	// client overridable configuration options
	flagSet.Duration("max-heartbeat-interval", opts.MaxHeartbeatInterval, "maximum client configurable duration of time between client heartbeats")
	flagSet.Int64("max-rdy-count", opts.MaxRdyCount, "maximum RDY count for a client")
	flagSet.Int64("max-output-buffer-size", opts.MaxOutputBufferSize, "maximum client configurable size (in bytes) for a client output buffer")
	flagSet.Duration("max-output-buffer-timeout", opts.MaxOutputBufferTimeout, "maximum client configurable duration of time between flushing to a client")
	flagSet.Duration("min-output-buffer-timeout", opts.MinOutputBufferTimeout, "minimum client configurable duration of time between flushing to a client")
	flagSet.Duration("output-buffer-timeout", opts.OutputBufferTimeout, "default duration of time between flushing data to clients")
	flagSet.Int("max-channel-consumers", opts.MaxChannelConsumers, "maximum channel consumer connection count per nsqd instance (default 0, i.e., unlimited)")

	// statsd integration options
	flagSet.String("statsd-address", opts.StatsdAddress, "UDP <addr>:<port> of a statsd daemon for pushing stats")
	flagSet.Duration("statsd-interval", opts.StatsdInterval, "duration between pushing to statsd")
	flagSet.Bool("statsd-mem-stats", opts.StatsdMemStats, "toggle sending memory and GC stats to statsd")
	flagSet.String("statsd-prefix", opts.StatsdPrefix, "prefix used for keys sent to statsd (%s for host replacement)")
	flagSet.Int("statsd-udp-packet-size", opts.StatsdUDPPacketSize, "the size in bytes of statsd UDP packets")

	// End to end percentile flags
	e2eProcessingLatencyPercentiles := app.FloatArray{}
	flagSet.Var(&e2eProcessingLatencyPercentiles, "e2e-processing-latency-percentile", "message processing time percentiles (as float (0, 1.0]) to track (can be specified multiple times or comma separated '1.0,0.99,0.95', default none)")
	flagSet.Duration("e2e-processing-latency-window-time", opts.E2EProcessingLatencyWindowTime, "calculate end to end latency quantiles for this duration of time (ie: 60s would only show quantile calculations from the past 60 seconds)")

	// TLS config
	flagSet.String("tls-cert", opts.TLSCert, "path to certificate file")
	flagSet.String("tls-key", opts.TLSKey, "path to key file")
	flagSet.String("tls-client-auth-policy", opts.TLSClientAuthPolicy, "client certificate auth policy ('require' or 'require-verify')")
	flagSet.String("tls-root-ca-file", opts.TLSRootCAFile, "path to certificate authority file")
	tlsRequired := tlsRequiredOption(opts.TLSRequired)
	tlsMinVersion := tlsMinVersionOption(opts.TLSMinVersion)
	flagSet.Var(&tlsRequired, "tls-required", "require TLS for client connections (true, false, tcp-https)")
	flagSet.Var(&tlsMinVersion, "tls-min-version", "minimum SSL/TLS version acceptable ('ssl3.0', 'tls1.0', 'tls1.1', or 'tls1.2')")

	// compression
	flagSet.Bool("deflate", opts.DeflateEnabled, "enable deflate feature negotiation (client compression)")
	flagSet.Int("max-deflate-level", opts.MaxDeflateLevel, "max deflate compression level a client can negotiate (> values == > nsqd CPU usage)")
	flagSet.Bool("snappy", opts.SnappyEnabled, "enable snappy feature negotiation (client compression)")

	return flagSet
}
