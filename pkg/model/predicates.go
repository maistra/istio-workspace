package model

// Predicate base function to filter Resources.
type Predicate func(ResourceStatus) bool

// Kind Predicate returns true if kind matches resource.
func Kind(kind string) Predicate {
	return func(resource ResourceStatus) bool {
		return resource.Kind == kind
	}
}

// Name Predicate returns true if name matches resource.
func Name(name string) Predicate {
	return func(resource ResourceStatus) bool {
		return resource.Name == name
	}
}
