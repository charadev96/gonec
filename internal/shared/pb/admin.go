package shared

import (
	adminpb "github.com/charadev96/gonec/gen/admin"
	server "github.com/charadev96/gonec/internal/server/domain"
)

func UserFromPB(pb *adminpb.User) (server.User, error) {
	id, err := UUIDFromPB(pb.Id)
	if err != nil {
		return server.User{}, err
	}
	return server.User{
		ID:        id,
		Name:      pb.Name,
		PublicKey: pb.PublicKey,
		State:     userStateFromPB(pb.State),
	}, nil
}

func UserToPB(u server.User) *adminpb.User {
	return &adminpb.User{
		Id:        UUIDToPB(u.ID),
		Name:      u.Name,
		PublicKey: u.PublicKey,
		State:     userStateToPB(u.State),
	}
}

func userStateFromPB(pb adminpb.UserState) server.UserState {
	switch pb {
	case adminpb.UserState_USER_STATE_REGISTERED:
		return server.StateRegistered
	case adminpb.UserState_USER_STATE_ACTIVE:
		return server.StateActive
	default:
		return server.StatePending
	}
}

func userStateToPB(s server.UserState) adminpb.UserState {
	switch s {
	case server.StateRegistered:
		return adminpb.UserState_USER_STATE_REGISTERED
	case server.StateActive:
		return adminpb.UserState_USER_STATE_ACTIVE
	default:
		return adminpb.UserState_USER_STATE_PENDING_UNSPECIFIED
	}
}
