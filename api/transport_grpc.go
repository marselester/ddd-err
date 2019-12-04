package api

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/ratelimit"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"golang.org/x/time/rate"

	account "github.com/marselester/ddd-err"
	pb "github.com/marselester/ddd-err/rpc/account"
)

// NewGRPCUserServer makes user service available as a gRPC UserServer.
func NewGRPCUserServer(s account.UserService, logger log.Logger, qps int) pb.UserServer {
	options := []grpctransport.ServerOption{
		grpctransport.ServerErrorLogger(logger),
	}
	// limiter throttles requests that exceeded qps requests per second.
	// For example, when qps is 100, there might be max 100 requests per seconds to
	// all the API endpoints combined.
	limiter := ratelimit.NewErroringLimiter(rate.NewLimiter(
		rate.Limit(qps), qps,
	))

	srv := userServer{}
	var ep endpoint.Endpoint
	{
		ep = makeFindUserByIDEndpoint(s)
		ep = limiter(ep)
		srv.findUserByIDHandler = grpctransport.NewServer(
			ep,
			decodeGRPCFindUserByIDReq,
			encodeGRPCFindUserByIDResp,
			options...,
		)
	}
	{
		ep = makeCreateUserEndpoint(s)
		ep = limiter(ep)
		srv.createUserHandler = grpctransport.NewServer(
			ep,
			decodeGRPCCreateUserReq,
			encodeGRPCCreateUserResp,
			options...,
		)
	}
	return &srv
}

// userServer is gRPC server that implements protobuf UserServer interface.
// It's like HTTP multiplexer.
type userServer struct {
	findUserByIDHandler grpctransport.Handler
	createUserHandler   grpctransport.Handler
}

// FindUserByID looks up a user by ID.
func (srv *userServer) FindUserByID(ctx context.Context, req *pb.FindUserByIDReq) (*pb.FindUserByIDResp, error) {
	_, resp, err := srv.findUserByIDHandler.ServeGRPC(ctx, req)
	if err != nil {
		return &pb.FindUserByIDResp{
			Error: encodeGRPCerror(err),
		}, nil
	}
	return resp.(*pb.FindUserByIDResp), nil
}

// CreateUser creates a user.
func (srv *userServer) CreateUser(ctx context.Context, req *pb.CreateUserReq) (*pb.CreateUserResp, error) {
	_, resp, err := srv.createUserHandler.ServeGRPC(ctx, req)
	if err != nil {
		return &pb.CreateUserResp{
			Error: encodeGRPCerror(err),
		}, nil
	}
	return resp.(*pb.CreateUserResp), nil
}

// decodeGRPCFindUserByIDReq is a transport/grpc.DecodeRequestFunc that converts a
// gRPC FindUserByIDReq request to a user-domain FindUserByIDReq request.
func decodeGRPCFindUserByIDReq(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.FindUserByIDReq)
	return FindUserByIDReq{ID: req.Id}, nil
}

// encodeGRPCFindUserByIDResp is a transport/grpc.EncodeResponseFunc that converts a
// user-domain FindUserByIDResp response to a gRPC FindUserByIDResp response.
func encodeGRPCFindUserByIDResp(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(FindUserByIDResp)
	return &pb.FindUserByIDResp{
		Id:       resp.ID,
		Username: resp.Username,
		Error:    encodeGRPCerror(resp.Err),
	}, nil
}

// decodeGRPCCreateUserReq is a transport/grpc.DecodeRequestFunc that converts a
// gRPC CreateUserReq request to a user-domain CreateUserReq request.
func decodeGRPCCreateUserReq(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.CreateUserReq)
	return CreateUserReq{Username: req.Username}, nil
}

// encodeGRPCCreateUserResp is a transport/grpc.EncodeResponseFunc that converts a
// user-domain CreateUserResp response to a gRPC CreateUserResp response.
func encodeGRPCCreateUserResp(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(CreateUserResp)
	return &pb.CreateUserResp{
		Error: encodeGRPCerror(resp.Err),
	}, nil
}

// encodeGRPCerror encodes domain error into gRPC error.
// It also encodes errors returned by grpctransport.Handler (e.g., ratelimit).
func encodeGRPCerror(err error) *pb.Error {
	if err == nil {
		return nil
	}

	var accErr account.Error
	if !errors.As(err, &accErr) {
		if errors.Is(err, ratelimit.ErrLimited) {
			accErr = account.Error{
				Code:    account.ERateLimit,
				Message: "API rate limit exceeded.",
			}
		} else {
			accErr = account.Error{
				Code:    account.EInternal,
				Message: "An internal error has occurred.",
			}
		}
	}

	return &pb.Error{
		Code:    accErr.Code,
		Message: accErr.Message,
	}
}
