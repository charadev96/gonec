package user

import (
	"context"

	userpb "github.com/charadev96/gonec/gen/user"
	"github.com/charadev96/gonec/internal/client/service"
	"github.com/charadev96/gonec/internal/shared/handler"
	pb "github.com/charadev96/gonec/internal/shared/pb"
)

// TODO: Sanitize errors

type AuthServiceHandler struct {
	userpb.UnimplementedAuthServiceServer
	Service *service.AuthService
}

func (h *AuthServiceHandler) Register(ctx context.Context, req *userpb.RegisterRequest) (*userpb.RegisterReply, error) {
	t, err := pb.InviteTicketFromPB(req.Ticket)
	if err != nil {
		return nil, handler.ErrArg(err)
	}
	err = h.Service.Register(ctx, req.ConnectionId, t)
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
