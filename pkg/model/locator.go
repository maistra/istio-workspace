package model

import (
	"sort"
	"strings"
	"time"
)

// LocatorStatus is the action to perform an a given Resource as calculated by the Locators.
type LocatorStatus struct {
	Resource
	TimeStamp time.Time
	Labels    map[string]string
	Action    StatusAction
}

type LocatorStatusReporter func(LocatorStatus)
type LocatedReporter func(LocatorStatusStore)
type LocatorStatusStore func(kind ...string) []LocatorStatus

// Locator should report on Resources that need some Action performed on them to satisfy the Ref.
type Locator func(context SessionContext, ref Ref, store LocatorStatusStore, reporter LocatorStatusReporter) error

// TODO dummy impl for testing purposes.
type LocatorStore struct {
	stored []LocatorStatus
}

func (l *LocatorStore) Store(kind ...string) []LocatorStatus {
	sorter := func(s []LocatorStatus) func(i, j int) bool {
		return func(i, j int) bool {
			nextActionIsUndo := s[j].Action == ActionDelete || s[j].Action == ActionRevert
			currentActionIsUndo := s[i].Action == ActionDelete || s[i].Action == ActionRevert
			if currentActionIsUndo && !nextActionIsUndo {
				return true
			}
			if !currentActionIsUndo && nextActionIsUndo {
				return false
			}

			return true
		}
	}

	if len(kind) == 0 {
		f := l.stored
		sort.SliceStable(f, sorter(f))

		return f
	}
	var f []LocatorStatus
	for _, loc := range l.stored {
		for _, k := range kind {
			if strings.EqualFold(loc.Kind, k) {
				f = append(f, loc)

				break
			}
		}
	}
	sort.SliceStable(f, sorter(f))

	return f
}

func (l *LocatorStore) Report(status LocatorStatus) {
	replaced := false

	if status.TimeStamp.IsZero() {
		status.TimeStamp = time.Now()
	}
	for i, stored := range l.stored {
		if stored.Name == status.Name && stored.Kind == status.Kind && stored.Action == status.Action {
			l.stored[i] = status
			replaced = true
		}
	}
	if !replaced {
		l.stored = append(l.stored, status)
	}
}

func (l *LocatorStore) Clear() {
	l.stored = []LocatorStatus{}
}
