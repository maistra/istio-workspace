package model_test

import (
	"fmt"
	"strconv"
	"testing"

	"emperror.dev/errors"

	"github.com/maistra/istio-workspace/pkg/model"
)

type Condition struct {
	// Human readable reason for the change
	Message string `json:"message,omitempty"`
	// Programmatic reason for the change
	Reason string `json:"reason,omitempty"`
	// Boolean value to indicate success
	Status string `json:"status,omitempty"`
	// The type of change
	Type string `json:"type,omitempty"`
}

func TestLocatorStoreSort(t *testing.T) {
	store := model.LocatorStore{}
	store.Report(model.LocatorStatus{Resource: model.Resource{Name: "X", Kind: "X"}, Action: model.ActionCreate})
	store.Report(model.LocatorStatus{Resource: model.Resource{Name: "X", Kind: "X"}, Action: model.ActionDelete})
	store.Report(model.LocatorStatus{Resource: model.Resource{Name: "Y", Kind: "X"}, Action: model.ActionDelete})
	store.Report(model.LocatorStatus{Resource: model.Resource{Name: "Y", Kind: "X"}, Action: model.ActionCreate})

	ls := store.Store("X")
	for i, l := range ls {
		if (i == 0 || i == 1) && l.Action != model.ActionDelete {
			t.Error("should sort Delete first")
		}
		if (i == 2 || i == 3) && l.Action != model.ActionCreate {
			t.Error("should sort Delete first")
		}
		fmt.Println(l)
	}
}

func TestRefHash(t *testing.T) {
	r1 := model.Ref{KindName: model.ParseRefKindName("x"), Namespace: "Y", Strategy: "X", Deleted: true, Args: map[string]string{"X": "Y", "Y": "X"}}
	r2 := model.Ref{KindName: model.ParseRefKindName("x"), Namespace: "Y", Strategy: "X", Deleted: true, Args: map[string]string{"Y": "X", "X": "Y"}}

	refs := []model.Ref{
		{KindName: model.ParseRefKindName("x"), Namespace: "Y", Strategy: "X", Deleted: true, Args: map[string]string{"Y": "X", "X": "X"}},
		{KindName: model.ParseRefKindName("x"), Namespace: "Y", Strategy: "X", Deleted: false, Args: map[string]string{"Y": "X", "X": "Y"}},
		{KindName: model.ParseRefKindName("x"), Namespace: "Y", Strategy: "Y", Deleted: true, Args: map[string]string{"Y": "X", "X": "Y"}},
		{KindName: model.ParseRefKindName("x"), Namespace: "X", Strategy: "X", Deleted: true, Args: map[string]string{"Y": "X", "X": "Y"}},
		{KindName: model.ParseRefKindName("Y"), Namespace: "Y", Strategy: "X", Deleted: true, Args: map[string]string{"Y": "X", "X": "Y"}},
		{KindName: model.ParseRefKindName("Y"), Namespace: "Y", Strategy: "X", Deleted: true},
		{KindName: model.ParseRefKindName("Y"), Namespace: "Y", Strategy: "X"},
		{KindName: model.ParseRefKindName("Y"), Namespace: "Y"},
		{KindName: model.ParseRefKindName("Y")},
	}

	if r1.Hash() != r2.Hash() {
		t.Errorf("Should match %v %v %v", r1.Hash() == r2.Hash(), r1.Hash(), r2.Hash())
	}

	for _, r := range refs {
		if r1.Hash() == r.Hash() {
			t.Errorf("Should match %v %v %v", r1.Hash() == r.Hash(), r1.Hash(), r.Hash())
		}
	}
}

func TestDesign(t *testing.T) {
	dryRun := false

	// Semi static configuration?
	validators := []model.Validator{IsDryRun(dryRun)}
	var locators []model.Locator
	var modificators []model.ModificatorRegistrar
	extractModificators := func(registrars []model.ModificatorRegistrar) []model.Modificator {
		var mods []model.Modificator
		for _, reg := range registrars {
			_, mod := reg()
			mods = append(mods, mod)
		}

		return mods
	}

	// Determine the state of each Ref in spec vs status
	refs := []model.Ref{{KindName: model.RefKindName{Kind: "Deployment", Name: "reviews-v1"}}}

	// Create engine and sync
	sync := model.NewSync(locators, extractModificators(modificators))

	for _, ref := range refs {
		sync(model.SessionContext{}, ref,
			func(located model.LocatorStatusStore) bool {
				errs := ValidationChain(located(), validators...)
				for _, err := range errs {
					addCondition(Condition{Type: "Validation", Reason: "Failed", Status: "false", Message: err.Error()})
				}

				return len(errs) == 0
			},
			func(located model.LocatorStatusStore) {
				fmt.Println("located: ", located())
			},
			func(modified model.ModificatorStatus) {
				msg := ""
				if modified.Error != nil {
					msg = modified.Error.Error()
				}
				addCondition(Condition{Type: string(modified.Action) + "-" + modified.Kind, Reason: "Required", Status: strconv.FormatBool(modified.Success), Message: msg})
			})
		// updateRefStatus
	}
	// updateSessionStatus
}

func addCondition(condition Condition) {
	fmt.Println("Condition:", condition.Type, condition.Status, condition.Reason, condition.Message)
}

func ValidationChain(located []model.LocatorStatus, validators ...model.Validator) []error {
	var errs []error
	for _, c := range validators {
		err := c(located)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func IsDryRun(dryRun bool) model.Validator {
	return func([]model.LocatorStatus) error {
		if dryRun {
			return errors.NewPlain("In dry run mode")
		}

		return nil
	}
}
