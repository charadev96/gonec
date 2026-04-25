package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	gatewaypb "github.com/charadev96/gonec/gen/gateway"
	sharedpb "github.com/charadev96/gonec/gen/shared"
	shared "github.com/charadev96/gonec/internal/shared/domain"
	pb "github.com/charadev96/gonec/internal/shared/pb"
)

type ChatService struct {
	auth *AuthService
}

type MessageStream struct {
	client grpc.ServerStreamingClient[sharedpb.Message]
}

func NewChatService(a *AuthService) *ChatService {
	return &ChatService{
		auth: a,
	}
}

func (s *MessageStream) Next() (shared.Message, error) {
	m, err := s.client.Recv()
	if err != nil {
		return shared.Message{}, err
	}
	return pb.MessageFromPB(m)
}

func (s *ChatService) Send(ctx context.Context, to uuid.UUID, str string) error {
	cl, err := BindClient(s.auth, gatewaypb.NewChatServiceClient)
	if err != nil {
		return err
	}

	session, err := s.auth.Session()
	if err != nil {
		return fmt.Errorf("get active session: %w", err)
	}

	_, err = cl.Send(ctx, &gatewaypb.SendRequest{
		Auth:      pb.SessionToPB(session),
		Recipient: to.String(),
		Content:   str,
	})
	if err != nil {
		return fmt.Errorf("request send: %w", err)
	}

	return nil
}

func (s *ChatService) Listen(ctx context.Context) (<-chan shared.Packet[shared.Message], error) {
	cl, err := BindClient(s.auth, gatewaypb.NewChatServiceClient)
	if err != nil {
		return nil, err
	}

	session, err := s.auth.Session()
	if err != nil {
		return nil, fmt.Errorf("get active session: %w", err)
	}

	ctxLn, cancel := context.WithCancel(ctx)

	ch, err := cl.Listen(ctxLn, &gatewaypb.ListenRequest{
		Auth: pb.SessionToPB(session),
	})
	if err != nil {
		cancel()
		return nil, fmt.Errorf("request listen: %w", err)
	}

	ln := make(chan shared.Packet[shared.Message])

	go func() {
		defer cancel()

		for {
			msg, err := ch.Recv()
			if err != nil {
				ln <- shared.Packet[shared.Message]{Err: err}
				return
			}

			m, err := pb.MessageFromPB(msg)
			if err != nil {
				ln <- shared.Packet[shared.Message]{Err: err}
				return
			}

			select {
			case ln <- shared.Packet[shared.Message]{Msg: m}:
			case <-ctx.Done():
				ln <- shared.Packet[shared.Message]{Err: context.Cause(ctx)}
				return
			}
		}
	}()

	return ln, nil
}
