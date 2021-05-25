package new

import (
	"fmt"
	"strconv"
	"testing"

	"emperror.dev/errors"
)

func TestDesign(t *testing.T) {

	dryRun := false

	// Semi static configuration?
	validators := []Validator{IsDryRun(dryRun)}
	locators := []Locator{DeploymentLocator}
	modificators := []ModificatorRegistrar{DeploymentRegistrar}
	extractModificators := func(registrars []ModificatorRegistrar) []Modificator {
		var mods []Modificator
		for _, reg := range registrars {
			_, mod := reg()
			mods = append(mods, mod)
		}
		return mods
	}
	/*
		extractTargetResourceType := func(registrars []ModificatorRegistrar) []client.Object {
			var types []client.Object
			for _, reg := range registrars {
				t, _ := reg()
				types = append(types, t)
			}
			return types
		}
	*/

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
				/* updateComponents(session.components + unique(located)) */
			},
			func(modified ModificatorStatus) {
				/* updateComponent() && addCondition(session) && callEventAPI() */

				msg := ""
				if modified.Error != nil {
					msg = modified.Error.Error()
				}
				addCondition(Condition{Type: string(modified.Action) + "-" + modified.Kind, Reason: "Required", Status: strconv.FormatBool(modified.Success), Message: msg})
				//fmt.Println("modified", modified)
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
