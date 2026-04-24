package shared

import (
	"time"

	sharedpb "github.com/charadev96/gonec/gen/shared"
	shared "github.com/charadev96/gonec/internal/shared/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func SessionFromPB(pb *sharedpb.Session) (shared.Session, error) {
	id, err := UUIDFromPB(pb.Id)
	if err != nil {
		return shared.Session{}, err
	}
	userID, err := UUIDFromPB(pb.UserId)
	if err != nil {
		return shared.Session{}, err
	}
	return shared.Session{
		ID:     id,
		UserID: userID,
		Token:  pb.Token,
	}, nil
}

func SessionToPB(s shared.Session) *sharedpb.Session {
	return &sharedpb.Session{
		Id:     UUIDToPB(s.ID),
		UserId: UUIDToPB(s.UserID),
		Token:  s.Token,
	}
}

func ServerIdentityFromPB(pb *sharedpb.ServerIdentity) shared.ServerIdentity {
	return shared.ServerIdentity{
		IPAddress: pb.IpAddress,
		PublicKey: pb.PublicKey,
	}
}

func ServerIdentityToPB(s shared.ServerIdentity) *sharedpb.ServerIdentity {
	return &sharedpb.ServerIdentity{
		IpAddress: s.IPAddress,
		PublicKey: s.PublicKey,
	}
}

func InviteCredentialFromPB(pb *sharedpb.InviteCredential) (shared.InviteCredential, error) {
	var notBefore, notAfter time.Time
	if pb.NotBefore != nil {
		notBefore = pb.NotBefore.AsTime()
	}
	if pb.NotAfter != nil {
		notAfter = pb.NotAfter.AsTime()
	}
	userID, err := UUIDFromPB(pb.UserId)
	if err != nil {
		return shared.InviteCredential{}, err
	}
	return shared.InviteCredential{
		UserID:    userID,
		Token:     pb.Token,
		NotBefore: notBefore,
		NotAfter:  notAfter,
	}, nil
}

func InviteCredentialToPB(i shared.InviteCredential) *sharedpb.InviteCredential {
	return &sharedpb.InviteCredential{
		UserId:    UUIDToPB(i.UserID),
		Token:     i.Token,
		NotBefore: timestamppb.New(i.NotBefore),
		NotAfter:  timestamppb.New(i.NotAfter),
	}
}

func InviteTicketFromPB(pb *sharedpb.InviteTicket) (shared.InviteTicket, error) {
	cred, err := InviteCredentialFromPB(pb.Credential)
	if err != nil {
		return shared.InviteTicket{}, err
	}
	return shared.InviteTicket{
		Server:     ServerIdentityFromPB(pb.Server),
		Credential: cred,
	}, nil
}

func InviteTicketToPB(t shared.InviteTicket) *sharedpb.InviteTicket {
	return &sharedpb.InviteTicket{
		Server:     ServerIdentityToPB(t.Server),
		Credential: InviteCredentialToPB(t.Credential),
	}
}
