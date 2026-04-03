package messaging

import (
	"context"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	gatewaypb "github.com/charadev96/gonec/gen/gateway"
	sharedpb "github.com/charadev96/gonec/gen/shared"
	"github.com/charadev96/gonec/internal/server/service"
	shared "github.com/charadev96/gonec/internal/shared/domain"
)

// TODO: Sanitize errors

type AuthServiceHandler struct {
	gatewaypb.UnimplementedAuthServiceServer
	Service *service.UserService
}

func (h *AuthServiceHandler) Register(ctx context.Context, req *gatewaypb.RegisterRequest) (*gatewaypb.RegisterReply, error) {
	id, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := h.Service.RegisterUser(ctx, id, req.Token, req.PublicKey); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &gatewaypb.RegisterReply{}, nil
}

func (h *AuthServiceHandler) InitiateLogin(ctx context.Context, req *gatewaypb.InitiateLoginRequest) (*gatewaypb.InitiateLoginReply, error) {
	id, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	nonce, err := h.Service.CreateLoginNonce(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &gatewaypb.InitiateLoginReply{Nonce: nonce}, nil
}

func (h *AuthServiceHandler) CompleteLogin(ctx context.Context, req *gatewaypb.CompleteLoginRequest) (*gatewaypb.CompleteLoginReply, error) {
	id, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	sess, err := h.Service.LoginUser(ctx, id, req.Signature)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	s := &sharedpb.Session{}
	copier.Copy(s, sess)
	return &gatewaypb.CompleteLoginReply{
		Auth: s,
	}, nil
}

func (h *AuthServiceHandler) Logout(ctx context.Context, req *gatewaypb.LogoutRequest) (*gatewaypb.LogoutReply, error) {
	id, err := uuid.Parse(req.Auth.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	uid, err := uuid.Parse(req.Auth.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	sess := shared.Session{
		ID:     id,
		UserID: uid,
		Token:  req.Auth.Token,
	}
	if err := h.Service.LogoutUser(ctx, sess); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &gatewaypb.LogoutReply{}, nil
}
