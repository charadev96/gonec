package gateway

import (
	"context"

	"google.golang.org/grpc"

	gatewaypb "github.com/charadev96/gonec/gen/gateway"
	sharedpb "github.com/charadev96/gonec/gen/shared"
	"github.com/charadev96/gonec/internal/server/service"
	shared "github.com/charadev96/gonec/internal/shared/domain"
	"github.com/charadev96/gonec/internal/shared/handler"
	pb "github.com/charadev96/gonec/internal/shared/pb"
)

// TODO: Sanitize errors

type ChatHandler struct {
	gatewaypb.UnimplementedChatServiceServer

	ctx     context.Context
	service *service.ChatService
}

func NewChatHandler(ctx context.Context, s *service.ChatService) *ChatHandler {
	return &ChatHandler{
		ctx:     ctx,
		service: s,
	}
}

func (h *ChatHandler) Send(ctx context.Context, req *gatewaypb.SendRequest) (*gatewaypb.SendReply, error) {
	if context.Cause(h.ctx) != nil {
		return nil, handler.ErrInternal(h.ctx.Err())
	}

	ids, err := handler.ParseUUIDs(req.Auth.Id, req.Auth.UserId, req.Recipient)
	if err != nil {
		return nil, handler.ErrArg(err)
	}
	auth := shared.Session{
		ID:     ids[0],
		UserID: ids[1],
		Token:  req.Auth.Token,
	}
	if err := h.service.Send(ctx, auth, ids[2], req.Content); err != nil {
		return nil, handler.ErrInternal(err)
	}
	return &gatewaypb.SendReply{}, nil
}

func (h *ChatHandler) Listen(req *gatewaypb.ListenRequest, stream grpc.ServerStreamingServer[sharedpb.Message]) error {
	if context.Cause(h.ctx) != nil {
		return handler.ErrInternal(h.ctx.Err())
	}

	ids, err := handler.ParseUUIDs(req.Auth.Id, req.Auth.UserId)
	if err != nil {
		return handler.ErrArg(err)
	}
	auth := shared.Session{
		ID:     ids[0],
		UserID: ids[1],
		Token:  req.Auth.Token,
	}

	ctxLn, _ := mergeCtx(h.ctx, stream.Context())
	ln, err := h.service.Listen(ctxLn, auth)
	if err != nil {
		return handler.ErrInternal(err)
	}

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

func mergeCtx(ctx1, ctx2 context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		select {
		case <-ctx1.Done():
		case <-ctx2.Done():
		case <-ctx.Done():
		}
		cancel()
	}()
	return ctx, cancel
}
