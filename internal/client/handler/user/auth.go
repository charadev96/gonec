package handler

import (
	"context"

	"github.com/jinzhu/copier"

	userpb "github.com/charadev96/gonec/gen/user"
	"github.com/charadev96/gonec/internal/client/service"
	shared "github.com/charadev96/gonec/internal/shared/domain"
	"github.com/charadev96/gonec/internal/shared/handler"
)

// TODO: Sanitize errors

type AuthServiceHandler struct {
	userpb.UnimplementedAuthServiceServer
	Service *service.AuthService
}

func (h *AuthServiceHandler) Register(ctx context.Context, req *userpb.RegisterRequest) (*userpb.RegisterReply, error) {
	t := shared.InviteTicket{}
	copier.Copy(&t, &req.Ticket)

	err := h.Service.Register(ctx, req.ConnectionId, t)
	if err != nil {
		return nil, handler.ErrInternal(err)
	}
	return &userpb.RegisterReply{}, nil
}

func (h *AuthServiceHandler) Login(ctx context.Context, req *userpb.LoginRequest) (*userpb.LoginReply, error) {
	err := h.Service.Login(ctx, req.ConnectionId)
	if err != nil {
		return nil, handler.ErrInternal(err)
	}
	return &userpb.LoginReply{}, nil
}

func (h *AuthServiceHandler) Logout(ctx context.Context, req *userpb.LogoutRequest) (*userpb.LogoutReply, error) {
	err := h.Service.Logout(ctx)
	if err != nil {
		return nil, handler.ErrInternal(err)
	}
	return &userpb.LogoutReply{}, nil
}
