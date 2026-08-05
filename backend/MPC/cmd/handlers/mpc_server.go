package handlers

import (
	"context"
	pb "mpc/gen/mpc"
)

type MpcServer interface {
	StartKeyGen(ctx context.Context, req *pb.StartKeyGenRequest) (*pb.StartKeyGenResponse, error)
	GetKeyGenStatus(ctx context.Context, req *pb.GetKeyGenStatusRequest) (*pb.GetKeyGenStatusResponse, error)
}
