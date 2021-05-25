package new

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

/*
	 Locator
	* Mutate
	* Revert

 	Mutator/Revertor

	ComponentLogger
	ConditionLogger
	EventLogger

	Accumulation Status




Session
	Ref
		- Y

	Status
		- Y



SessionController -> Convert(Session->Model) -> Engine -> Locator|Mutator -> ConditionLogger(Convert(Model->Session))
                                                                          -> ComponentLogger(Convert(Model->Session))
                                                                          -> EventLogger(Convert(Model->Event API))


Locator <-- Object | List
	Service
	Gateway
	VitualService
	DestinationRule
	Deployment
	DeploymentConfig
	...

Modificator <-- Object | List
	Gateway
	VitualService
	DestinationRule
	Deployment
	DeploymentConfig

*/

/*
 *  MODEL API - Level
 */

// ResourceAction describes which type of operation was done/attempted to the target resource. Used to determine how to undo it.
type StatusAction string

const (
	// ActionCreated imply the whole Named Kind was created and can be deleted.
	ActionCreated StatusAction = "created"
	// ActionModified imply the Named Kind has been modified and needs to be reverted to get back to original state.
	ActionModified StatusAction = "modified"
	// ActionLocated imply the resource was found, but nothing was changed.
	ActionLocated StatusAction = "located"
)

type SessionContext struct {
	context.Context

	Name      string
	Namespace string
	Client    client.Client
	Log       logr.Logger
}

type Ref struct {
	KindName RefKindName
	Deleted  bool
}

type RefKindName struct {
	Kind string
	Name string
}

type LocatorStatus struct {
	Namespace string
	Kind      string
	Name      string
	Action    StatusAction // Create, Modify, Located
	Labels    map[string]string
}

type LocatorStatusReporter func(LocatorStatus)
type LocatedReporter func(LocatorStatusStore)
type LocatorStatusStore func(kind ...string) []LocatorStatus

type Locator func(context SessionContext, ref Ref, store LocatorStatusStore, reporter LocatorStatusReporter)

type ModificatorStatusReporter func(ModificatorStatus)

type ModificatorController func(LocatorStatusStore) bool

type ModificatorStatus struct {
	LocatorStatus
	Error   error
	Success bool
}

type ModificatorRegistrar func() (targetResourceType client.Object, modificator Modificator)
type Modificator func(context SessionContext, ref Ref, store LocatorStatusStore, reporter ModificatorStatusReporter)

type Sync func(SessionContext, Ref, ModificatorController, LocatedReporter, ModificatorStatusReporter)

type LocatorStore struct {
	stored []LocatorStatus
}

func (l *LocatorStore) Store() LocatorStatusStore {
	return func(kind ...string) []LocatorStatus {
		if len(kind) == 0 {
			return l.stored
		}
		var f []LocatorStatus
		for _, loc := range l.stored {
			for _, k := range kind {
				if loc.Kind == k {
					f = append(f, loc)
					break
				}
			}
		}
		return f
	}
}

func (l *LocatorStore) Report() LocatorStatusReporter {
	return func(status LocatorStatus) {
		l.stored = append(l.stored, status)
	}
}

/*
 *  Session Controller - Level
 */

type Validator func([]LocatorStatus) error

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

/*
 *  Impl
 */

func EngineImpl(locators []Locator, modificators []Modificator) Sync {

	return func(context SessionContext, ref Ref, modify ModificatorController, locatedReporter LocatedReporter, modificationReporter ModificatorStatusReporter) {
		located := LocatorStore{}
		for _, locator := range locators {
			locator(
				context,
				ref,
				located.Store(),
				located.Report(),
			)
		}
		if !modify(located.Store()) {
			return
		}

		locatedReporter(located.Store())

		for _, modificator := range modificators {
			modificator(context, ref, located.Store(), modificationReporter)
		}
	}

}

func DeploymentLocator(context SessionContext, ref Ref, store LocatorStatusStore, report LocatorStatusReporter) {
	//deployment, err := getDeployment(ctx, ctx.Namespace, ref.KindName.Name)

	//report(LocatorStatus{Kind: "Deployment", name: deployment.Name, Namespace: ctx.Namespace, Create: true})
	report(LocatorStatus{Kind: "Deployment", Name: "test", Namespace: "namespace-test", Action: "Create"})
}

func DeploymentRegistrar() (client.Object, Modificator) {
	return nil, DeploymentModificator
}

func DeploymentModificator(context SessionContext, ref Ref, store LocatorStatusStore, reporter ModificatorStatusReporter) {
	for _, located := range store("Deployment") {

		// get Deployment
		// clone

		if located.Action == "Create" {
			/*
				// create clone.. contexxt.Client.Create(clone...)
				if err != nil {
					reporter(ModificatorStatus{LocatorStatus: located, Details: err.Error(), Success: false})
				}
			*/
			reporter(ModificatorStatus{LocatorStatus: located, Success: true})
		}
	}
}

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
