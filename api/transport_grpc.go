package api

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	grpctransport "github.com/go-kit/kit/transport/grpc"

	account "github.com/marselester/ddd-err"
	pb "github.com/marselester/ddd-err/rpc/account"
)

// NewGRPCUserServer makes user service available as a gRPC UserServer.
func NewGRPCUserServer(s account.UserService, logger log.Logger) pb.UserServer {
	options := []grpctransport.ServerOption{
		grpctransport.ServerErrorLogger(logger),
	}

	srv := userServer{}
	var ep endpoint.Endpoint
	{
		ep = makeFindUserByIDEndpoint(s)
		srv.findUserByIDHandler = grpctransport.NewServer(
			ep,
			decodeGRPCfindUserByIDReq,
			encodeGRPCfindUserByIDResp,
			options...,
		)
	}
	{
		ep = makeCreateUserEndpoint(s)
		srv.createUserHandler = grpctransport.NewServer(
			ep,
			decodeGRPCcreateUserReq,
			encodeGRPCcreateUserResp,
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
		return nil, err
	}
	return resp.(*pb.FindUserByIDResp), nil
}

// CreateUser creates a user.
func (srv *userServer) CreateUser(ctx context.Context, req *pb.CreateUserReq) (*pb.CreateUserResp, error) {
	_, resp, err := srv.createUserHandler.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.CreateUserResp), nil
}

// decodeGRPCfindUserByIDReq is a transport/grpc.DecodeRequestFunc that converts a
// gRPC FindUserByIDReq request to a user-domain FindUserByIDReq request.
func decodeGRPCfindUserByIDReq(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.FindUserByIDReq)
	return FindUserByIDReq{ID: req.Id}, nil
}

// encodeGRPCfindUserByIDResp is a transport/grpc.EncodeResponseFunc that converts a
// user-domain FindUserByIDResp response to a gRPC FindUserByIDResp reply.
func encodeGRPCfindUserByIDResp(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(FindUserByIDResp)
	return &pb.FindUserByIDResp{
		Id:       resp.ID,
		Username: resp.Username,
		Error:    encodeGRPCerror(resp.Err),
	}, nil
}

// decodeGRPCcreateUserReq is a transport/grpc.DecodeRequestFunc that converts a
// gRPC CreateUserReq request to a user-domain CreateUserReq request.
func decodeGRPCcreateUserReq(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.CreateUserReq)
	return CreateUserReq{Username: req.Username}, nil
}

// encodeGRPCcreateUserResp is a transport/grpc.EncodeResponseFunc that converts a
// user-domain CreateUserResp response to a gRPC CreateUserResp reply.
func encodeGRPCcreateUserResp(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(CreateUserResp)
	return &pb.CreateUserResp{
		Error: encodeGRPCerror(resp.Err),
	}, nil
}

// encodeGRPCerror encodes domain error into gRPC error.
func encodeGRPCerror(err error) *pb.Error {
	if err == nil {
		return nil
	}

	var accErr account.Error
	if !errors.As(err, &accErr) {
		accErr = account.Error{
			Code:    account.EInternal,
			Message: "An internal error has occurred.",
		}
	}

	return &pb.Error{
		Code:    accErr.Code,
		Message: accErr.Message,
	}
}
