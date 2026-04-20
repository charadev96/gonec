package adminpb

import (
	"context"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/types/known/timestamppb"

	adminpb "github.com/charadev96/gonec/gen/admin"
	sharedpb "github.com/charadev96/gonec/gen/shared"
	server "github.com/charadev96/gonec/internal/server/domain"
	"github.com/charadev96/gonec/internal/server/service"
	"github.com/charadev96/gonec/internal/shared/handler"
)

// TODO: Sanitize errors

type UserServiceHandler struct {
	adminpb.UnimplementedUserServiceServer
	Service *service.UserService
}

func (h *UserServiceHandler) CreateUser(ctx context.Context, req *adminpb.CreateUserRequest) (*adminpb.CreateUserReply, error) {
	id, err := h.Service.Users().Create(ctx)
	if err != nil {
		return nil, handler.ErrInternal(err)
	}
	return &adminpb.CreateUserReply{UserId: id.String()}, nil
}

func (h *UserServiceHandler) CreateInvite(ctx context.Context, req *adminpb.CreateInviteRequest) (*adminpb.CreateInviteReply, error) {
	id, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, handler.ErrArg(err)
	}
	opts := service.CreateInviteOptions{
		NotBefore: req.NotBefore.AsTime(),
		NotAfter:  req.NotAfter.AsTime(),
	}
	inv, err := h.Service.CreateInvite(ctx, id, opts)
	if err != nil {
		return nil, handler.ErrInternal(err)
	}
	i := &sharedpb.InviteCredential{}
	copier.Copy(i, &inv)
	i.NotBefore = timestamppb.New(inv.NotBefore)
	i.NotAfter = timestamppb.New(inv.NotAfter)
	return &adminpb.CreateInviteReply{Invite: i}, nil
}

func (h *UserServiceHandler) ExportInvite(ctx context.Context, req *adminpb.ExportInviteRequest) (*adminpb.ExportInviteReply, error) {
	id, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, handler.ErrArg(err)
	}
	tck, err := h.Service.ExportInvite(ctx, id)
	if err != nil {
		return nil, handler.ErrInternal(err)
	}
	t := &sharedpb.InviteTicket{}
	copier.Copy(t, &tck)
	return &adminpb.ExportInviteReply{Ticket: t}, nil
}

func (h *UserServiceHandler) GetUserByID(ctx context.Context, req *adminpb.GetByIDRequest) (*adminpb.GetUserReply, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, handler.ErrArg(err)
	}
	user, err := h.Service.Users().GetByID(ctx, id)
	if err != nil {
		return nil, handler.ErrInternal(err)
	}
	u := new(adminpb.User)
	copier.Copy(u, &user)
	return &adminpb.GetUserReply{User: u}, nil
}

func (h *UserServiceHandler) GetUserByName(ctx context.Context, req *adminpb.GetByNameRequest) (*adminpb.GetUserReply, error) {
	user, err := h.Service.Users().GetByName(ctx, req.Name)
	if err != nil {
		return nil, handler.ErrInternal(err)
	}
	u := new(adminpb.User)
	copier.Copy(u, &user)
	return &adminpb.GetUserReply{User: u}, nil
}

func (h *UserServiceHandler) GetInviteByUserID(ctx context.Context, req *adminpb.GetByIDRequest) (*adminpb.GetInviteReply, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, handler.ErrArg(err)
	}
	invite, err := h.Service.Invites().GetByUserID(ctx, id)
	if err != nil {
		return nil, handler.ErrInternal(err)
	}
	i := &sharedpb.InviteCredential{}
	copier.Copy(i, &invite)
	return &adminpb.GetInviteReply{Invite: i}, nil
}

func (h *UserServiceHandler) ListUsers(ctx context.Context, req *adminpb.ListUsersRequest) (*adminpb.ListUsersReply, error) {
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
	list, err := h.Service.Users().List(ctx, query)
	if err != nil {
		return nil, handler.ErrInternal(err)
	}

	var users []*adminpb.User
	for _, u := range list.Users {
		users = append(users, &adminpb.User{})
		copier.Copy(users[len(users)-1], &u)
	}

	return &adminpb.ListUsersReply{
		Users:  users,
		Cursor: list.Cursor.String(),
	}, nil

}

func (h *UserServiceHandler) DeleteUser(ctx context.Context, req *adminpb.DeleteRequest) (*adminpb.DeleteReply, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, handler.ErrArg(err)
	}
	err = h.Service.DeleteUser(ctx, id)
	if err != nil {
		return nil, handler.ErrInternal(err)
	}
	return &adminpb.DeleteReply{}, nil
}

func (h *UserServiceHandler) DeleteInvite(ctx context.Context, req *adminpb.DeleteRequest) (*adminpb.DeleteReply, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, handler.ErrArg(err)
	}
	err = h.Service.Invites().Delete(ctx, id)
	if err != nil {
		return nil, handler.ErrInternal(err)
	}
	return &adminpb.DeleteReply{}, nil
}
