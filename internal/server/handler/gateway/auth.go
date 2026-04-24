package gateway

import (
	"context"

	"github.com/google/uuid"

	gatewaypb "github.com/charadev96/gonec/gen/gateway"
	"github.com/charadev96/gonec/internal/server/service"
	shared "github.com/charadev96/gonec/internal/shared/domain"
	"github.com/charadev96/gonec/internal/shared/handler"
	pb "github.com/charadev96/gonec/internal/shared/pb"
)

// TODO: Sanitize errors

type AuthHandler struct {
	gatewaypb.UnimplementedAuthServiceServer
	service *service.UserService
}

func NewAuthHandler(s *service.UserService) *AuthHandler {
	return &AuthHandler{service: s}
}

func (h *AuthHandler) Register(ctx context.Context, req *gatewaypb.RegisterRequest) (*gatewaypb.RegisterReply, error) {
	id, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, handler.ErrArg(err)
	}
	if err := h.service.RegisterUser(ctx, id, req.Token, req.PublicKey); err != nil {
		return nil, handler.ErrInternal(err)
	}
	return &gatewaypb.RegisterReply{}, nil
}

func (h *AuthHandler) InitiateLogin(ctx context.Context, req *gatewaypb.InitiateLoginRequest) (*gatewaypb.InitiateLoginReply, error) {
	id, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, handler.ErrArg(err)
	}
	nonce, err := h.service.CreateLoginNonce(ctx, id)
	if err != nil {
		return nil, handler.ErrInternal(err)
	}
	return &gatewaypb.InitiateLoginReply{Nonce: nonce}, nil
}

func (h *AuthHandler) CompleteLogin(ctx context.Context, req *gatewaypb.CompleteLoginRequest) (*gatewaypb.CompleteLoginReply, error) {
	id, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, handler.ErrArg(err)
	}
	sess, err := h.service.LoginUser(ctx, id, req.Signature)
	if err != nil {
		return nil, handler.ErrInternal(err)
	}
	return &gatewaypb.CompleteLoginReply{
		Auth: pb.SessionToPB(sess),
	}, nil
}

func (h *AuthHandler) Logout(ctx context.Context, req *gatewaypb.LogoutRequest) (*gatewaypb.LogoutReply, error) {
	ids, err := handler.ParseUUIDs(req.Auth.Id, req.Auth.UserId)
	if err != nil {
		return nil, handler.ErrArg(err)
	}
	sess := shared.Session{
		ID:     ids[0],
		UserID: ids[1],
		Token:  req.Auth.Token,
	}
	if err := h.service.LogoutUser(ctx, sess); err != nil {
		return nil, handler.ErrInternal(err)
	}
	return &gatewaypb.LogoutReply{}, nil
}
