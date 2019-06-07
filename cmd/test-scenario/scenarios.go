package main

// TestScenario1ThreeServicesInSequence is a basic test setup with a few services
// calling each other in a chain. Similar to the original bookinfo example setup.
func TestScenario1ThreeServicesInSequence() {
	services := []Entry{{"productpage", "Deployment"}, {"reviews", "Deployment"}, {"ratings", "Deployment"}}
	Generate(
		services,
		WithVersion("v1"),
		ForService("productpage", Call("reviews"), ConnectToGateway()),
		ForService("reviews", Call("ratings")),
	)
}

// TestScenario2ThreeServicesInSequenceDeploymentConfig is a basic test setup with a
// few services calling each other in a chain. Similar to the original bookinfo example setup.
// Using DeploymentConfig.
func TestScenario2ThreeServicesInSequenceDeploymentConfig() {
	services := []Entry{{"productpage", "DeploymentConfig"}, {"reviews", "DeploymentConfig"}, {"ratings", "DeploymentConfig"}}
	Generate(
		services,
		WithVersion("v1"),
		ForService("productpage", Call("reviews"), ConnectToGateway()),
		ForService("reviews", Call("ratings")),
	)
}
