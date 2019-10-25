package main

import (
	"fmt"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	"github.com/sereiner/duo/example/rpc"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
)

func main() {

	reporter := zipkinhttp.NewReporter("http://127.0.0.1:9411/api/v2/spans")
	defer reporter.Close()

	// create our local service endpoint
	endpoint, err := zipkin.NewEndpoint("myService", "myservice.mydomain.com:80")
	if err != nil {
		log.Fatalf("unable to create local endpoint: %+v\n", err)
	}

	// initialize our tracer
	nativeTracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))
	if err != nil {
		log.Fatalf("unable to create tracer: %+v\n", err)
	}

	tracer := zipkinot.Wrap(nativeTracer)

	// optionally set as Global OpenTracing tracer instance
	opentracing.SetGlobalTracer(tracer)

	conn, err := grpc.DialContext(context.Background(), "127.0.0.1:8090", grpc.WithInsecure(), grpc.WithUnaryInterceptor(
		otgrpc.OpenTracingClientInterceptor(tracer, otgrpc.LogPayloads()),
	))
	if err != nil {
		panic(err)
	}
	//ticker := time.NewTicker(1 * time.Second)

	client := rpc.NewSearchServiceClient(conn)
	resp, err := client.Search(context.Background(), &rpc.SearchRequest{Request: "world fafa"})
	if err == nil {
		fmt.Printf(" Reply is %s\n", resp.Response)
	}

}
