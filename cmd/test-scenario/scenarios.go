package main

// TestScenario1 is a basic test setup with a few services calling each other in a chain. Similar to the original bookinfo example setup.
func TestScenario1() {
	// Scenario 1
	services := []string{"productpage", "reviews", "ratings"}
	Generate(
		services,
		WithVersion("v1"),
		ForService("productpage", Call("reviews"), ConnectToGatway()),
		ForService("reviews", Call("ratings")),
	)
}
