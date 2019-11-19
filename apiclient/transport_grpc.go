package apiclient

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"

	account "github.com/marselester/ddd-err"
	"github.com/marselester/ddd-err/api"
	pb "github.com/marselester/ddd-err/rpc/account"
)

// NewGRPCUserClient returns a gRPC client for a user service.
// The caller is responsible for constructing the conn, and eventually closing the underlying transport.
func NewGRPCUserClient(conn *grpc.ClientConn) account.UserService {
	c := client{}
	var ep endpoint.Endpoint
	{
		ep = grpctransport.NewClient(
			conn,
			"ddd_err.account.User",
			"FindUserByID",
			encodeGRPCFindUserByIDReq,
			decodeGRPCFindUserByIDResp,
			pb.FindUserByIDResp{},
		).Endpoint()
		c.findUserByIDEndpoint = ep
	}
	{
		ep = grpctransport.NewClient(
			conn,
			"ddd_err.account.User",
			"CreateUser",
			encodeGRPCCreateUserReq,
			decodeGRPCCreateUserResp,
			pb.CreateUserResp{},
		).Endpoint()
		c.createUserEndpoint = ep
	}
	return &c
}

// encodeGRPCFindUserByIDReq is a transport/grpc.EncodeRequestFunc that converts
// a user-domain FindUserByIDReq to a gRPC FindUserByIDReq.
func encodeGRPCFindUserByIDReq(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(api.FindUserByIDReq)
	return &pb.FindUserByIDReq{
		Id: req.ID,
	}, nil
}

// decodeGRPCFindUserByIDResp is a transport/grpc.DecodeResponseFunc that converts a
// gRPC FindUserByIDResp to a user-domain FindUserByIDResp.
func decodeGRPCFindUserByIDResp(_ context.Context, grpcResp interface{}) (interface{}, error) {
	resp := grpcResp.(*pb.FindUserByIDResp)
	if resp.Error == nil {
		return api.FindUserByIDResp{
			ID:       resp.Id,
			Username: resp.Username,
		}, nil
	}
	// Decode gRPC error into domain error.
	return api.FindUserByIDResp{
		Err: account.Error{
			Code:    resp.Error.Code,
			Message: resp.Error.Message,
		},
	}, nil
}

// encodeGRPCCreateUserReq is a transport/grpc.EncodeRequestFunc that converts
// a user-domain CreateUserReq to a gRPC CreateUserReq.
func encodeGRPCCreateUserReq(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(api.CreateUserReq)
	return &pb.CreateUserReq{
		Username: req.Username,
	}, nil
}

// decodeGRPCCreateUserResp is a transport/grpc.DecodeResponseFunc that converts a
// gRPC CreateUserResp to a user-domain CreateUserResp.
func decodeGRPCCreateUserResp(_ context.Context, grpcResp interface{}) (interface{}, error) {
	resp := grpcResp.(*pb.CreateUserResp)
	if resp.Error == nil {
		return api.CreateUserResp{}, nil
	}
	// Decode gRPC error into domain error.
	return api.CreateUserResp{
		Err: account.Error{
			Code:    resp.Error.Code,
			Message: resp.Error.Message,
		},
	}, nil
}