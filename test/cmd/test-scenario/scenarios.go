package main

// TestScenario1ThreeServicesInSequence is a basic test setup with a few services
// calling each other in a chain. Similar to the original bookinfo example setup.
func TestScenario1ThreeServicesInSequence() {
	productpage := Entry{"productpage", "Deployment", targetNamespace}
	reviews := Entry{"reviews", "Deployment", targetNamespace}
	ratings := Entry{"ratings", "Deployment", targetNamespace}

	Generate(
		[]Entry{productpage, reviews, ratings},
		WithVersion("v1"),
		ForService(productpage, Call(reviews), ConnectToGateway()),
		ForService(reviews, Call(ratings)),
		GatewayOnHost(gatewayHost),
	)
}

// TestScenario2ThreeServicesInSequenceDeploymentConfig is a basic test setup with a
// few services calling each other in a chain. Similar to the original bookinfo example setup.
// Using DeploymentConfig.
func TestScenario2ThreeServicesInSequenceDeploymentConfig() {
	productpage := Entry{"productpage", "DeploymentConfig", targetNamespace}
	reviews := Entry{"reviews", "DeploymentConfig", targetNamespace}
	ratings := Entry{"ratings", "DeploymentConfig", targetNamespace}

	Generate(
		[]Entry{productpage, reviews, ratings},
		WithVersion("v1"),
		ForService(productpage, Call(reviews), ConnectToGateway()),
		ForService(reviews, Call(ratings)),
		GatewayOnHost(gatewayHost),
	)
}

// DemoScenario is a simple setup for demo purposes
func DemoScenario() {
	productpage := Entry{"productpage", "Deployment", targetNamespace}
	reviews := Entry{"reviews", "Deployment", targetNamespace}
	ratings := Entry{"ratings", "Deployment", targetNamespace}
	authors := Entry{"authors", "Deployment", targetNamespace}
	locations := Entry{"locations", "Deployment", targetNamespace}
	Generate(
		[]Entry{productpage, reviews, ratings, authors, locations},
		WithVersion("v1"),
		ForService(productpage, Call(reviews), Call(authors), ConnectToGateway()),
		ForService(reviews, Call(ratings)),
		ForService(authors, Call(locations)),
		GatewayOnHost("ike-demo.io"), GatewayOnHost("*.ike-demo.io"),
	)
}
