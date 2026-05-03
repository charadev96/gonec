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

func InviteClaimsFromPB(pb *sharedpb.InviteClaims) (shared.InviteClaims, error) {
	var notBefore, notAfter time.Time
	if pb.NotBefore != nil {
		notBefore = pb.NotBefore.AsTime()
	}
	if pb.NotAfter != nil {
		notAfter = pb.NotAfter.AsTime()
	}
	userID, err := UUIDFromPB(pb.UserId)
	if err != nil {
		return shared.InviteClaims{}, err
	}
	return shared.InviteClaims{
		UserID:    userID,
		Token:     pb.Token,
		NotBefore: notBefore,
		NotAfter:  notAfter,
	}, nil
}

func InviteClaimsToPB(i shared.InviteClaims) *sharedpb.InviteClaims {
	return &sharedpb.InviteClaims{
		UserId:    UUIDToPB(i.UserID),
		Token:     i.Token,
		NotBefore: timestamppb.New(i.NotBefore),
		NotAfter:  timestamppb.New(i.NotAfter),
	}
}

func InviteTicketFromPB(pb *sharedpb.InviteTicket) (shared.InviteTicket, error) {
	cl, err := InviteClaimsFromPB(pb.Claims)
	if err != nil {
		return shared.InviteTicket{}, err
	}
	return shared.InviteTicket{
		Server: ServerIdentityFromPB(pb.Server),
		Claims: cl,
	}, nil
}

func InviteTicketToPB(t shared.InviteTicket) *sharedpb.InviteTicket {
	return &sharedpb.InviteTicket{
		Server: ServerIdentityToPB(t.Server),
		Claims: InviteClaimsToPB(t.Claims),
	}
}

func MessageFromPB(pb *sharedpb.Message) (shared.Message, error) {
	sender, err := UUIDFromPB(pb.Sender)
	if err != nil {
		return shared.Message{}, err
	}
	return shared.Message{
		Sender:  sender,
		Content: pb.Content,
	}, nil
}

func MessageToPB(m shared.Message) *sharedpb.Message {
	return &sharedpb.Message{
		Sender:  UUIDToPB(m.Sender),
		Content: m.Content,
	}
}
