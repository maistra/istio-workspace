package new_test

import (
	"fmt"
	"strconv"
	"testing"

	"emperror.dev/errors"

	. "github.com/maistra/istio-workspace/pkg/model/new"
)

func TestLocatorStoreSort(t *testing.T) {
	store := LocatorStore{}
	store.Report(LocatorStatus{Resource: Resource{Name: "X", Kind: "X"}, Action: ActionCreate})
	store.Report(LocatorStatus{Resource: Resource{Name: "X", Kind: "X"}, Action: ActionDelete})
	store.Report(LocatorStatus{Resource: Resource{Name: "Y", Kind: "X"}, Action: ActionDelete})
	store.Report(LocatorStatus{Resource: Resource{Name: "Y", Kind: "X"}, Action: ActionCreate})

	ls := store.Store("X")
	for i, l := range ls {
		if (i == 0 || i == 1) && l.Action != ActionDelete {
			t.Error("should sort Delete first")
		}
		if (i == 2 || i == 3) && l.Action != ActionCreate {
			t.Error("should sort Delete first")
		}
		fmt.Println(l)
	}
}

func TestRefHash(t *testing.T) {
	r1 := Ref{KindName: ParseRefKindName("x"), Namespace: "Y", Strategy: "X", Deleted: true, Args: map[string]string{"X": "Y", "Y": "X"}}
	r2 := Ref{KindName: ParseRefKindName("x"), Namespace: "Y", Strategy: "X", Deleted: true, Args: map[string]string{"Y": "X", "X": "Y"}}

	refs := []Ref{
		{KindName: ParseRefKindName("x"), Namespace: "Y", Strategy: "X", Deleted: true, Args: map[string]string{"Y": "X", "X": "X"}},
		{KindName: ParseRefKindName("x"), Namespace: "Y", Strategy: "X", Deleted: false, Args: map[string]string{"Y": "X", "X": "Y"}},
		{KindName: ParseRefKindName("x"), Namespace: "Y", Strategy: "Y", Deleted: true, Args: map[string]string{"Y": "X", "X": "Y"}},
		{KindName: ParseRefKindName("x"), Namespace: "X", Strategy: "X", Deleted: true, Args: map[string]string{"Y": "X", "X": "Y"}},
		{KindName: ParseRefKindName("Y"), Namespace: "Y", Strategy: "X", Deleted: true, Args: map[string]string{"Y": "X", "X": "Y"}},
		{KindName: ParseRefKindName("Y"), Namespace: "Y", Strategy: "X", Deleted: true},
		{KindName: ParseRefKindName("Y"), Namespace: "Y", Strategy: "X"},
		{KindName: ParseRefKindName("Y"), Namespace: "Y"},
		{KindName: ParseRefKindName("Y")},
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
	validators := []Validator{IsDryRun(dryRun)}
	var locators []Locator
	var modificators []ModificatorRegistrar
	extractModificators := func(registrars []ModificatorRegistrar) []Modificator {
		var mods []Modificator
		for _, reg := range registrars {
			_, mod := reg()
			mods = append(mods, mod)
		}

		return mods
	}

	// Determine the state of each Ref in spec vs status
	refs := []Ref{{KindName: RefKindName{Kind: "Deployment", Name: "reviews-v1"}}}

	// Create engine and sync
	sync := EngineImpl(locators, extractModificators(modificators))

	for _, ref := range refs {
		sync(SessionContext{}, ref,
			func(located LocatorStatusStore) bool {
				errs := ValidationChain(located(), validators...)
				for _, err := range errs {
					addCondition(Condition{Type: "Validation", Reason: "Failed", Status: "false", Message: err.Error()})
				}

				return len(errs) == 0
			},
			func(located LocatorStatusStore) {
				fmt.Println("located: ", located())
			},
			func(modified ModificatorStatus) {
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

func ValidationChain(located []LocatorStatus, validators ...Validator) []error {
	var errs []error
	for _, c := range validators {
		err := c(located)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func IsDryRun(dryRun bool) Validator {
	return func([]LocatorStatus) error {
		if dryRun {
			return errors.NewPlain("In dry run mode")
		}

		return nil
	}
}
