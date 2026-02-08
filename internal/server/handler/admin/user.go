package admin

import (
	"context"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/charadev96/gonec/api/admin/gen"
	server "github.com/charadev96/gonec/internal/server/domain"
	"github.com/charadev96/gonec/internal/server/service"
)

// TODO: Sanitize errors

type UserServiceHandler struct {
	gen.UnimplementedUserServiceServer
	Service *service.UserService
}

func (h *UserServiceHandler) CreateUser(ctx context.Context, req *gen.CreateUserRequest) (*gen.CreateUserReply, error) {
	id, err := h.Service.Users.Create(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &gen.CreateUserReply{Id: id.String()}, nil
}

func (h *UserServiceHandler) CreateInvite(ctx context.Context, req *gen.CreateInviteRequest) (*gen.CreateInviteReply, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	opts := service.CreateInviteOptions{
		NotBefore: req.NotBefore.AsTime(),
		NotAfter:  req.NotAfter.AsTime(),
	}
	inv, err := h.Service.CreateInvite(ctx, id, opts)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	i := new(gen.UserInvite)
	copier.Copy(i, &inv)
	i.NotBefore = timestamppb.New(inv.NotBefore)
	i.NotAfter = timestamppb.New(inv.NotAfter)
	return &gen.CreateInviteReply{Invite: i}, nil
}

func (h *UserServiceHandler) ExportInvite(ctx context.Context, req *gen.ExportInviteRequest) (*gen.ExportInviteReply, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	mnf, err := h.Service.ExportInvite(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	m := new(gen.UserInviteManifest)
	copier.Copy(m, &mnf)
	return &gen.ExportInviteReply{Manifest: m}, nil
}

func (h *UserServiceHandler) GetUserByID(ctx context.Context, req *gen.GetByIDRequest) (*gen.GetUserReply, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	user, err := h.Service.Users.GetByID(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	u := new(gen.User)
	copier.Copy(u, &user)
	return &gen.GetUserReply{User: u}, nil
}

func (h *UserServiceHandler) GetUserByName(ctx context.Context, req *gen.GetByNameRequest) (*gen.GetUserReply, error) {
	user, err := h.Service.Users.GetByName(ctx, req.Name)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	u := new(gen.User)
	copier.Copy(u, &user)
	return &gen.GetUserReply{User: u}, nil
}

func (h *UserServiceHandler) GetInviteByUserID(ctx context.Context, req *gen.GetByIDRequest) (*gen.GetInviteReply, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	invite, err := h.Service.Invites.GetByUserID(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	i := new(gen.UserInvite)
	copier.Copy(i, &invite)
	return &gen.GetInviteReply{Invite: i}, nil
}

func (h *UserServiceHandler) ListUsers(ctx context.Context, req *gen.ListUsersRequest) (*gen.ListUsersReply, error) {
	var cursor uuid.UUID
	if req.Cursor != "" {
		var err error
		cursor, err = uuid.Parse(req.Cursor)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}
	query := server.UserListQuery{
		Limit:  int(req.Limit),
		Cursor: cursor,
	}
	list, err := h.Service.Users.List(ctx, query)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var users []*gen.User
	for _, u := range list.Users {
		users = append(users, &gen.User{})
		copier.Copy(users[len(users)-1], &u)
	}

	return &gen.ListUsersReply{
		Users:  users,
		Cursor: list.Cursor.String(),
	}, nil

}

func (h *UserServiceHandler) DeleteUser(ctx context.Context, req *gen.DeleteRequest) (*gen.DeleteReply, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	err = h.Service.DeleteUser(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &gen.DeleteReply{}, nil
}

func (h *UserServiceHandler) DeleteInvite(ctx context.Context, req *gen.DeleteRequest) (*gen.DeleteReply, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	err = h.Service.Invites.Delete(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &gen.DeleteReply{}, nil
}
