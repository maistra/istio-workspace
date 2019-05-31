package main

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
