package main

// TestScenario1_Three_Services_In_Sequence is a basic test setup with a few services calling each other in a chain. Similar to the original bookinfo example setup.
func TestScenario1_Three_Services_In_Sequence() {
	// Scenario 1
	services := []string{"productpage", "reviews", "ratings"}
	Generate(
		services,
		WithVersion("v1"),
		ForService("productpage", Call("reviews"), ConnectToGatway()),
		ForService("reviews", Call("ratings")),
	)
}
