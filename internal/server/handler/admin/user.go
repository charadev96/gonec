package admin

import (
	"context"

	"github.com/google/uuid"

	adminpb "github.com/charadev96/gonec/gen/admin"
	server "github.com/charadev96/gonec/internal/server/domain"
	"github.com/charadev96/gonec/internal/server/service"
	"github.com/charadev96/gonec/internal/shared/handler"
	pb "github.com/charadev96/gonec/internal/shared/pb"
)

// TODO: Sanitize errors

type UserHandler struct {
	adminpb.UnimplementedUserServiceServer
	service *service.UserService
}

func NewUserHandler(s *service.UserService) *UserHandler {
	return &UserHandler{service: s}
}

func (h *UserHandler) CreateUser(ctx context.Context, req *adminpb.CreateUserRequest) (*adminpb.CreateUserReply, error) {
	id, err := h.service.Users().Create(ctx)
	if err != nil {
		return nil, handler.ErrInternal(err)
	}
	return &adminpb.CreateUserReply{UserId: id.String()}, nil
}

func (h *UserHandler) CreateInvite(ctx context.Context, req *adminpb.CreateInviteRequest) (*adminpb.CreateInviteReply, error) {
	id, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, handler.ErrArg(err)
	}
	opts := service.CreateInviteOptions{
		NotBefore: req.NotBefore.AsTime(),
		NotAfter:  req.NotAfter.AsTime(),
	}
	inv, err := h.service.CreateInvite(ctx, id, opts)
	if err != nil {
		return nil, handler.ErrInternal(err)
	}
	return &adminpb.CreateInviteReply{Invite: pb.InviteCredentialToPB(inv)}, nil
}

func (h *UserHandler) ExportInvite(ctx context.Context, req *adminpb.ExportInviteRequest) (*adminpb.ExportInviteReply, error) {
	id, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, handler.ErrArg(err)
	}
	tck, err := h.service.ExportInvite(ctx, id)
	if err != nil {
		return nil, handler.ErrInternal(err)
	}
	return &adminpb.ExportInviteReply{Ticket: pb.InviteTicketToPB(tck)}, nil
}

func (h *UserHandler) GetUserByID(ctx context.Context, req *adminpb.GetByIDRequest) (*adminpb.GetUserReply, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, handler.ErrArg(err)
	}
	user, err := h.service.Users().GetByID(ctx, id)
	if err != nil {
		return nil, handler.ErrInternal(err)
	}
	return &adminpb.GetUserReply{User: pb.UserToPB(user)}, nil
}

func (h *UserHandler) GetUserByName(ctx context.Context, req *adminpb.GetByNameRequest) (*adminpb.GetUserReply, error) {
	user, err := h.service.Users().GetByName(ctx, req.Name)
	if err != nil {
		return nil, handler.ErrInternal(err)
	}
	return &adminpb.GetUserReply{User: pb.UserToPB(user)}, nil
}

func (h *UserHandler) GetInviteByUserID(ctx context.Context, req *adminpb.GetByIDRequest) (*adminpb.GetInviteReply, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, handler.ErrArg(err)
	}
	invite, err := h.service.Invites().GetByUserID(ctx, id)
	if err != nil {
		return nil, handler.ErrInternal(err)
	}
	return &adminpb.GetInviteReply{Invite: pb.InviteCredentialToPB(invite)}, nil
}

func (h *UserHandler) ListUsers(ctx context.Context, req *adminpb.ListUsersRequest) (*adminpb.ListUsersReply, error) {
	var cursor uuid.UUID
	if req.Cursor != "" {
		var err error
		cursor, err = uuid.Parse(req.Cursor)
		if err != nil {
			return nil, handler.ErrArg(err)
		}
	}

	query := server.UserListQuery{
		Limit:  int(req.Limit),
		Cursor: cursor,
	}
	list, err := h.service.Users().List(ctx, query)
	if err != nil {
		return nil, handler.ErrInternal(err)
	}

	users := make([]*adminpb.User, len(list.Users))
	for i, u := range list.Users {
		users[i] = pb.UserToPB(u)
	}

	return &adminpb.ListUsersReply{
		Users:  users,
		Cursor: list.Cursor.String(),
	}, nil

}

func (h *UserHandler) DeleteUser(ctx context.Context, req *adminpb.DeleteRequest) (*adminpb.DeleteReply, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, handler.ErrArg(err)
	}
	err = h.service.DeleteUser(ctx, id)
	if err != nil {
		return nil, handler.ErrInternal(err)
	}
	return &adminpb.DeleteReply{}, nil
}

func (h *UserHandler) DeleteInvite(ctx context.Context, req *adminpb.DeleteRequest) (*adminpb.DeleteReply, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, handler.ErrArg(err)
	}
	err = h.service.Invites().Delete(ctx, id)
	if err != nil {
		return nil, handler.ErrInternal(err)
	}
	return &adminpb.DeleteReply{}, nil
}
