package test

import (
	"net/http"
	"testing"

	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/server/grpc/pb"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	grpcHost = "grpc:50051"
	httpHost = "http://http:8000"
)

func TestHttpSuite(t *testing.T) {
	suite.Run(t, &HTTPTestSuite{
		events: make([]int64, 0),
		client: http.Client{},
		host:   httpHost,
	})
}

func TestGRPCSuite(t *testing.T) {
	cc, err := grpc.Dial(grpcHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	suite.Run(t, &GRPCTestSuite{
		events: make([]int64, 0),
		client: pb.NewCalendarClient(cc),
	})
}
