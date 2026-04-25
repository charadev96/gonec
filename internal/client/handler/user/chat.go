package user

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	sharedpb "github.com/charadev96/gonec/gen/shared"
	userpb "github.com/charadev96/gonec/gen/user"
	"github.com/charadev96/gonec/internal/client/service"
	"github.com/charadev96/gonec/internal/shared/handler"
	pb "github.com/charadev96/gonec/internal/shared/pb"
)

// TODO: Sanitize errors

type ChatHandler struct {
	userpb.UnimplementedChatServiceServer

	ctx     context.Context
	service *service.ChatService
}

func NewChatHandler(ctx context.Context, s *service.ChatService) *ChatHandler {
	return &ChatHandler{
		ctx:     ctx,
		service: s,
	}
}

func (h *ChatHandler) Send(ctx context.Context, req *userpb.SendRequest) (*userpb.SendReply, error) {
	if context.Cause(h.ctx) != nil {
		return nil, handler.ErrInternal(h.ctx.Err())
	}

	toid, err := uuid.Parse(req.Recipient)
	if err != nil {
		return nil, handler.ErrArg(err)
	}
	if err := h.service.Send(ctx, toid, req.Content); err != nil {
		return nil, handler.ErrInternal(err)
	}
	return &userpb.SendReply{}, nil
}

func (h *ChatHandler) Listen(req *userpb.ListenRequest, stream grpc.ServerStreamingServer[sharedpb.Message]) error {
	if context.Cause(h.ctx) != nil {
		return handler.ErrInternal(h.ctx.Err())
	}

	ln, err := h.service.Listen(stream.Context())
	if err != nil {
		return handler.ErrInternal(err)
	}

	go func() {

	}()

	for {
		select {
		case <-h.ctx.Done():
			return handler.ErrInternal(h.ctx.Err())
		case <-stream.Context().Done():
			return handler.ErrInternal(stream.Context().Err())
		case pck := <-ln:
			if pck.Err != nil {
				return handler.ErrInternal(err)
			}
			if err := stream.Send(pb.MessageToPB(pck.Msg)); err != nil {
				return handler.ErrInternal(pck.Err)
			}
		}
	}
}
