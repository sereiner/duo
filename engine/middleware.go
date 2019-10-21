package engine

import (
	"context"
	logger "github.com/sereiner/library/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"net"
	"runtime/debug"
	"time"
)

func LoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	start := time.Now()
	log := logger.GetSession("request", logger.CreateSession())
	ctx = context.WithValue(ctx, "__log__", log)

	ip, ok := ctx.Value("__ip__").(string)
	if !ok {
		ip = ""
	}

	ctx.Value("__log__").(*logger.Logger).Info("request: ", info.FullMethod, "from", ip, req)
	resp, err := handler(ctx, req)
	ctx.Value("__log__").(*logger.Logger).Info("response: ", info.FullMethod, time.Since(start), resp)

	ctx.Value("__log__").(*logger.Logger).Close()
	return resp, err
}

func GetClientIP(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	pr, ok := peer.FromContext(ctx)
	if !ok {
		return handler(ctx, req)
	}
	if pr.Addr == net.Addr(nil) {
		return handler(ctx, req)
	}
	ctx = context.WithValue(ctx, "__ip__", pr.Addr.String())
	return handler(ctx, req)
}

func RecoveryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			debug.PrintStack()
			err = status.Errorf(codes.Internal, "Panic err: %v", e)
		}
	}()

	return handler(ctx, req)
}
