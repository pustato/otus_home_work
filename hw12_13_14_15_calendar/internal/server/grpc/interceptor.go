package grpcserver

import (
	"context"
	"strings"
	"time"

	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/logger"
	"google.golang.org/grpc"
)

const timeLayout = "[02/Jan/2006:15:04:05 -0700]"

func unaryLoggingInterceptor(log logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		start := time.Now()

		resp, err = handler(ctx, req)

		msg := strings.Join([]string{
			start.Format(timeLayout),
			info.FullMethod,
			time.Since(start).String(),
		}, " ")
		log.Info(msg,
			"type", "access",
			"context", "grpc",
		)

		if err != nil {
			log.Error(err.Error(),
				"context", "grpc",
			)
		}

		return resp, err
	}
}
