package component

import (
	"context"
	logger "github.com/sereiner/library/log"
	"google.golang.org/grpc"
)

type IContainer interface {
	GetGrpcServer() *grpc.Server
	Log(ctx context.Context) *logger.Logger
}
