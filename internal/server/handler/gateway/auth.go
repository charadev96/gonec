package messaging

import (
	"context"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"

	gatewaypb "github.com/charadev96/gonec/gen/gateway"
	sharedpb "github.com/charadev96/gonec/gen/shared"
	"github.com/charadev96/gonec/internal/server/service"
	shared "github.com/charadev96/gonec/internal/shared/domain"
	"github.com/charadev96/gonec/internal/shared/handler"
)

// TODO: Sanitize errors

type AuthServiceHandler struct {
	gatewaypb.UnimplementedAuthServiceServer
	Service *service.UserService
}

func (h *AuthServiceHandler) Register(ctx context.Context, req *gatewaypb.RegisterRequest) (*gatewaypb.RegisterReply, error) {
	id, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, handler.ErrArg(err)
	}
	if err := h.Service.RegisterUser(ctx, id, req.Token, req.PublicKey); err != nil {
		return nil, handler.ErrInternal(err)
	}
	return &gatewaypb.RegisterReply{}, nil
}

func (h *AuthServiceHandler) InitiateLogin(ctx context.Context, req *gatewaypb.InitiateLoginRequest) (*gatewaypb.InitiateLoginReply, error) {
	id, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, handler.ErrArg(err)
	}
	nonce, err := h.Service.CreateLoginNonce(ctx, id)
	if err != nil {
		return nil, handler.ErrInternal(err)
	}
	return &gatewaypb.InitiateLoginReply{Nonce: nonce}, nil
}

func (h *AuthServiceHandler) CompleteLogin(ctx context.Context, req *gatewaypb.CompleteLoginRequest) (*gatewaypb.CompleteLoginReply, error) {
	id, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, handler.ErrArg(err)
	}
	sess, err := h.Service.LoginUser(ctx, id, req.Signature)
	if err != nil {
		return nil, handler.ErrInternal(err)
	}
	s := &sharedpb.Session{}
	copier.Copy(s, sess)
	return &gatewaypb.CompleteLoginReply{
		Auth: s,
	}, nil
}

func (h *AuthServiceHandler) Logout(ctx context.Context, req *gatewaypb.LogoutRequest) (*gatewaypb.LogoutReply, error) {
	ids, err := handler.ParseUUIDs(req.Auth.Id, req.Auth.UserId)
	if err != nil {
		return nil, handler.ErrArg(err)
	}
	sess := shared.Session{
		ID:     ids[0],
		UserID: ids[1],
		Token:  req.Auth.Token,
	}
	if err := h.Service.LogoutUser(ctx, sess); err != nil {
		return nil, handler.ErrInternal(err)
	}
	return &gatewaypb.LogoutReply{}, nil
}
