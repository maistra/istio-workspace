package model

// Sync is the entry point for ensuring the desired state for the given Ref is up to date.
type Sync func(SessionContext, Ref, ModificatorController, LocatedReporter, ModificatorStatusReporter)

func NewSync(locators []Locator, modificators []Modificator) Sync {
	return func(context SessionContext, ref Ref, modify ModificatorController, locatedReporter LocatedReporter, modificationReporter ModificatorStatusReporter) {
		located := LocatorStore{}
		for _, locator := range locators {
			err := locator(
				context,
				ref,
				located.Store,
				located.Report,
			)

			if err != nil {
				context.Log.Error(err, "locating failed", "locator", locator)
			}
		}
		if !modify(located.Store) {
			return
		}

		locatedReporter(located.Store)

		for _, modificator := range modificators {
			modificator(context, ref, located.Store, modificationReporter)
		}
	}
}
