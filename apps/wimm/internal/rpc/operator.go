package rpc

import (
	"context"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	identityv1 "github.com/xuuid/wimm/packages/contracts/gen/go/wimm/identity/v1"

	"github.com/xuuid/wimm/apps/wimm/internal/identity"
	"github.com/xuuid/wimm/apps/wimm/internal/store"
)

// OperatorServer answers the operator surface. It converts, calls the domain,
// converts back, and maps errors. No business logic.
type OperatorServer struct {
	identity *identity.Service
}

// NewOperatorServer wires the handler to the domain.
func NewOperatorServer(svc *identity.Service) *OperatorServer {
	return &OperatorServer{identity: svc}
}

func (s *OperatorServer) RegisterMember(
	ctx context.Context,
	req *connect.Request[identityv1.RegisterMemberRequest],
) (*connect.Response[identityv1.RegisterMemberResponse], error) {
	m, link, err := s.identity.RegisterMember(ctx,
		req.Msg.GetEmail().GetValue(), req.Msg.GetFirstName(), req.Msg.GetLastName())
	if err != nil {
		return nil, toConnectError(err)
	}

	return connect.NewResponse(&identityv1.RegisterMemberResponse{
		Member:        toProtoMember(m),
		EnrolmentLink: toProtoLink(link),
	}), nil
}

func (s *OperatorServer) IssueEnrolmentLink(
	ctx context.Context,
	req *connect.Request[identityv1.IssueEnrolmentLinkRequest],
) (*connect.Response[identityv1.IssueEnrolmentLinkResponse], error) {
	m, link, err := s.identity.IssueEnrolmentLink(ctx, req.Msg.GetEmail().GetValue())
	if err != nil {
		return nil, toConnectError(err)
	}

	return connect.NewResponse(&identityv1.IssueEnrolmentLinkResponse{
		Member:        toProtoMember(m),
		EnrolmentLink: toProtoLink(link),
	}), nil
}

func toProtoMember(m store.Member) *identityv1.Member {
	return &identityv1.Member{
		Id:        m.ID,
		Email:     m.Email,
		FirstName: m.FirstName,
		LastName:  m.LastName,
	}
}

func toProtoLink(l identity.EnrolmentLink) *identityv1.EnrolmentLink {
	return &identityv1.EnrolmentLink{
		Url:       l.URL,
		ExpiresAt: timestamppb.New(l.ExpiresAt),
	}
}
