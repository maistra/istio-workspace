package model

// TODO rethink naming.
func EngineImpl(locators []Locator, modificators []Modificator) Sync {
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
				context.Log.Error(err, "Could not perform locator action", "locator", locator)
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
