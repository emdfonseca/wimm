package banking

import (
	"context"
	"slices"
)

// Scope is which of a member's owned accounts everything drawn from
// transactions counts. The zero value is All, which is what an unset request
// means and what a member gets until they touch the control.
type Scope int

const (
	ScopeAll Scope = iota
	ScopeHousehold
	ScopeOwn
)

// scoped is a scope resolved for one member.
type scoped struct {
	// Answered is the scope actually applied: the one asked for when it is on
	// offer, otherwise All.
	Answered Scope
	// Available is empty when no control is to be shown.
	Available []Scope
	// AccountIDs are owned accounts only. An account the member merely holds a
	// grant on is in no scope (ADR 0021).
	AccountIDs []string
	// Counted and NotCounted name household accounts, and are set only under
	// Household when it counts fewer accounts than the household figure does.
	Counted, NotCounted []string
}

// scopedAccounts resolves a scope to the ids of the owned accounts it counts.
// The store applies the same ids beside its ownership join, so this decides
// which accounts and the store still refuses one that is not the member's.
func (s *Service) scopedAccounts(ctx context.Context, memberID string, asked Scope) (scoped, error) {
	visible, err := s.store.VisibleAccounts(ctx, memberID, "")
	if err != nil {
		return scoped{}, err
	}
	groups, err := s.groupAccounts(ctx, visible)
	if err != nil {
		return scoped{}, err
	}

	var household, yours, notOwned, householdNames []string
	for _, a := range visible {
		switch groups[a.ID] {
		case GroupHousehold:
			if a.Owned {
				household = append(household, a.ID)
				householdNames = append(householdNames, a.Name)
			} else {
				notOwned = append(notOwned, a.Name)
			}
		case GroupOwn:
			yours = append(yours, a.ID)
		}
	}

	out := scoped{Answered: ScopeAll}
	if len(household) > 0 && len(yours) > 0 {
		out.Available = []Scope{ScopeHousehold, ScopeOwn, ScopeAll}
		if slices.Contains(out.Available, asked) {
			out.Answered = asked
		}
	}

	switch out.Answered {
	case ScopeHousehold:
		out.AccountIDs = household
		if len(notOwned) > 0 {
			out.Counted, out.NotCounted = householdNames, notOwned
		}
	case ScopeOwn:
		out.AccountIDs = yours
	default:
		out.AccountIDs = slices.Concat(household, yours)
	}
	return out, nil
}
