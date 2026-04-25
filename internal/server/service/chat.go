package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	server "github.com/charadev96/gonec/internal/server/domain"
	shared "github.com/charadev96/gonec/internal/shared/domain"
)

var msgQueueSize = 128

type Lock struct{}

type MessageBroker struct {
	inboxes map[uuid.UUID]chan shared.Message
	mu      chan Lock
}

func NewMessageBroker() *MessageBroker {
	return &MessageBroker{
		inboxes: make(map[uuid.UUID]chan shared.Message),
		mu:      make(chan Lock, 1),
	}
}

func (b *MessageBroker) Get(ctx context.Context, id uuid.UUID) (chan shared.Message, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case b.mu <- Lock{}:
		defer func() { <-b.mu }()

		var ok bool
		c, ok := b.inboxes[id]
		if !ok {
			c = make(chan shared.Message, msgQueueSize)
			b.inboxes[id] = c
		}
		return c, nil
	}
}

type ChatService struct {
	users server.UserRepository
	user  *UserService

	msgs *MessageBroker
}

func NewChatService(r server.UserRepository, s *UserService) *ChatService {
	return &ChatService{
		users: r,
		user:  s,
		msgs:  NewMessageBroker(),
	}
}

func (s *ChatService) Send(ctx context.Context, auth shared.Session, to uuid.UUID, str string) error {
	err := s.user.VerifySession(ctx, auth)
	if err != nil {
		return fmt.Errorf("verify session: %w", err)
	}

	_, err = s.users.GetByID(ctx, to)
	if err != nil {
		return fmt.Errorf("get recipient: %w", err)
	}

	ch, err := s.msgs.Get(ctx, to)
	if err != nil {
		return err
	}

	select {
	case ch <- shared.Message{
		Sender:  auth.UserID,
		Content: str,
	}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *ChatService) Listen(ctx context.Context, auth shared.Session) (<-chan shared.Packet[shared.Message], error) {
	err := s.user.VerifySession(ctx, auth)

	if err != nil {
		return nil, fmt.Errorf("verify session: %w", err)
	}

	ch, err := s.msgs.Get(ctx, auth.UserID)
	if err != nil {
		return nil, err
	}

	ln := make(chan shared.Packet[shared.Message])

	go func() {
		defer close(ln)
		for {
			select {
			case <-ctx.Done():
				ln <- shared.Packet[shared.Message]{Err: context.Cause(ctx)}
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				select {
				case ln <- shared.Packet[shared.Message]{Msg: msg}:
				case <-ctx.Done():
					ln <- shared.Packet[shared.Message]{Err: context.Cause(ctx)}
					return
				}
			}
		}
	}()

	return ln, nil
}
