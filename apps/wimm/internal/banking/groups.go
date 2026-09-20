package banking

import (
	"context"

	"github.com/xuuid/wimm/apps/wimm/internal/store"
)

// Group is what an account counts towards for one member.
type Group int

const (
	// GroupNone is a left-out account: it counts towards nothing.
	GroupNone Group = iota
	// GroupHousehold is an account every member owns or has details on.
	GroupHousehold
	// GroupOwn is an account the member owns that is not household money.
	GroupOwn
	// GroupShared is anything else the member may see. It counts in no figure.
	GroupShared
)

// groupAccounts decides, once and in one place, which figure each visible
// account belongs to. The web load never classifies anything: a group worked
// out in two places is two answers waiting to differ.
//
// A household of one has nobody to hold money with, so nothing it owns is
// household money.
func (s *Service) groupAccounts(ctx context.Context, visible []store.VisibleAccount) (map[string]Group, error) {
	members, err := s.store.Members(ctx)
	if err != nil {
		return nil, err
	}
	full, err := s.store.FullAccessCounts(ctx)
	if err != nil {
		return nil, err
	}

	groups := make(map[string]Group, len(visible))
	for _, a := range visible {
		switch {
		case a.LeftOut():
			groups[a.ID] = GroupNone
		case len(members) >= 2 && full[a.ID] == len(members):
			groups[a.ID] = GroupHousehold
		case a.Owned:
			groups[a.ID] = GroupOwn
		default:
			groups[a.ID] = GroupShared
		}
	}
	return groups, nil
}
