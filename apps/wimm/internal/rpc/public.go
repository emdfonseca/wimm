package rpc

import (
	"context"

	"connectrpc.com/connect"

	identityv1 "github.com/emdfonseca/wimm/packages/contracts/gen/go/wimm/identity/v1"

	"github.com/emdfonseca/wimm/apps/wimm/internal/identity"
)

// PublicServer answers the surface the browser reaches.
type PublicServer struct {
	identity *identity.Service
	cookies  cookiePolicy
}

// NewPublicServer wires the handler to the domain. Cookie attributes follow
// the configured origins.
func NewPublicServer(svc *identity.Service, origins []string) *PublicServer {
	return &PublicServer{identity: svc, cookies: policyForOrigins(origins)}
}

// RedeemEnrolmentLink exchanges the value in the URL for a ticket in a cookie.
//
// The web app calls this on first load and then redirects to a path that does
// not carry the value, so what lingers in history cannot be read off the URL.
func (s *PublicServer) RedeemEnrolmentLink(
	ctx context.Context,
	req *connect.Request[identityv1.RedeemEnrolmentLinkRequest],
) (*connect.Response[identityv1.RedeemEnrolmentLinkResponse], error) {
	ticket, m, err := s.identity.RedeemEnrolmentLink(ctx, req.Msg.GetLinkValue())
	if err != nil {
		return nil, toConnectError(err)
	}

	res := connect.NewResponse(&identityv1.RedeemEnrolmentLinkResponse{
		EnrolmentTicket: ticket.Value,
		Member:          toProtoMember(m),
	})
	res.Header().Add("Set-Cookie", s.cookies.set(EnrolmentCookie, ticket.Value, ticket.ExpiresAt).String())
	return res, nil
}

func (s *PublicServer) BeginEnrolment(
	ctx context.Context,
	req *connect.Request[identityv1.BeginEnrolmentRequest],
) (*connect.Response[identityv1.BeginEnrolmentResponse], error) {
	ticket := or(req.Msg.GetEnrolmentTicket(), cookieValue(req.Header(), EnrolmentCookie))

	ceremony, m, err := s.identity.BeginEnrolment(ctx, ticket)
	if err != nil {
		return nil, toConnectError(err)
	}

	return connect.NewResponse(&identityv1.BeginEnrolmentResponse{
		Member:              toProtoMember(m),
		CeremonyId:          ceremony.ID,
		CreationOptionsJson: ceremony.OptionsJSON,
	}), nil
}

func (s *PublicServer) FinishEnrolment(
	ctx context.Context,
	req *connect.Request[identityv1.FinishEnrolmentRequest],
) (*connect.Response[identityv1.FinishEnrolmentResponse], error) {
	m, session, err := s.identity.FinishEnrolment(ctx, req.Msg.GetCeremonyId(), req.Msg.GetCredentialJson())
	if err != nil {
		return nil, toConnectError(err)
	}

	res := connect.NewResponse(&identityv1.FinishEnrolmentResponse{Member: toProtoMember(m)})
	// Enrolment signs them in: the ceremony just verified them, and a second
	// one seconds later proves nothing new.
	res.Header().Add("Set-Cookie", s.cookies.set(SessionCookie, session.Value, session.ExpiresAt).String())
	res.Header().Add("Set-Cookie", s.cookies.clear(EnrolmentCookie).String())
	return res, nil
}

func (s *PublicServer) BeginSignIn(
	ctx context.Context,
	req *connect.Request[identityv1.BeginSignInRequest],
) (*connect.Response[identityv1.BeginSignInResponse], error) {
	ceremony, err := s.identity.BeginSignIn(ctx, req.Msg.GetIntendedPath())
	if err != nil {
		return nil, toConnectError(err)
	}

	return connect.NewResponse(&identityv1.BeginSignInResponse{
		CeremonyId:         ceremony.ID,
		RequestOptionsJson: ceremony.OptionsJSON,
	}), nil
}

func (s *PublicServer) FinishSignIn(
	ctx context.Context,
	req *connect.Request[identityv1.FinishSignInRequest],
) (*connect.Response[identityv1.FinishSignInResponse], error) {
	m, session, returnPath, err := s.identity.FinishSignIn(ctx, req.Msg.GetCeremonyId(), req.Msg.GetCredentialJson())
	if err != nil {
		return nil, toConnectError(err)
	}

	res := connect.NewResponse(&identityv1.FinishSignInResponse{
		Member:     toProtoMember(m),
		ReturnPath: returnPath,
	})
	res.Header().Add("Set-Cookie", s.cookies.set(SessionCookie, session.Value, session.ExpiresAt).String())
	return res, nil
}

func (s *PublicServer) SignOut(
	ctx context.Context,
	req *connect.Request[identityv1.SignOutRequest],
) (*connect.Response[identityv1.SignOutResponse], error) {
	if err := s.identity.SignOut(ctx, cookieValue(req.Header(), SessionCookie)); err != nil {
		return nil, toConnectError(err)
	}

	res := connect.NewResponse(&identityv1.SignOutResponse{})
	res.Header().Add("Set-Cookie", s.cookies.clear(SessionCookie).String())
	return res, nil
}

func (s *PublicServer) GetCurrentMember(
	ctx context.Context,
	req *connect.Request[identityv1.GetCurrentMemberRequest],
) (*connect.Response[identityv1.GetCurrentMemberResponse], error) {
	m, err := s.identity.MemberForSession(ctx, cookieValue(req.Header(), SessionCookie))
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&identityv1.GetCurrentMemberResponse{Member: toProtoMember(m)}), nil
}

func or(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
