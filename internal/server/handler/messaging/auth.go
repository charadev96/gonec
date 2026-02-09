package messaging

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	//"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/charadev96/gonec/api/messaging/gen"
	server "github.com/charadev96/gonec/internal/server/domain"
	"github.com/charadev96/gonec/internal/server/service"
)

// TODO: Sanitize errors

type AuthServiceHandler struct {
	gen.UnimplementedAuthServiceServer
	Service *service.UserService
}

func (h *AuthServiceHandler) Register(ctx context.Context, req *gen.RegisterRequest) (*gen.RegisterReply, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := h.Service.RegisterUser(ctx, id, req.Token, req.PublicKey); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &gen.RegisterReply{}, nil
}

func (h *AuthServiceHandler) InitiateLogin(ctx context.Context, req *gen.InitiateLoginRequest) (*gen.InitiateLoginReply, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	nonce, err := h.Service.CreateLoginNonce(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &gen.InitiateLoginReply{ChallengeNonce: nonce}, nil
}

func (h *AuthServiceHandler) CompleteLogin(ctx context.Context, req *gen.CompleteLoginRequest) (*gen.CompleteLoginReply, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	sess, err := h.Service.LoginUser(ctx, id, req.ChallengeSigned)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &gen.CompleteLoginReply{
		SessionId:    sess.ID.String(),
		SessionToken: sess.Token,
	}, nil
}

func (h *AuthServiceHandler) Logout(ctx context.Context, req *gen.LogoutRequest) (*gen.LogoutReply, error) {
	id, err := uuid.Parse(req.Auth.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	uid, err := uuid.Parse(req.Auth.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	sess := server.UserSession{
		ID:     id,
		UserID: uid,
		Token:  req.Auth.Token,
	}
	if err := h.Service.LogoutUser(ctx, sess); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &gen.LogoutReply{}, nil
}
