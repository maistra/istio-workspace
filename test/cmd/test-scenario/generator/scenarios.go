package generator

var (
	Namespace = ""
)

// TestScenario1HTTPThreeServicesInSequence is a basic test setup with a few services
// calling each other in a chain over http. Similar to the original bookinfo example setup.
func TestScenario1HTTPThreeServicesInSequence(out Printer) {
	productpage := Entry{"productpage", "Deployment", Namespace}
	reviews := Entry{"reviews", "Deployment", Namespace}
	ratings := Entry{"ratings", "Deployment", Namespace}

	Generate(
		out,
		[]Entry{productpage, reviews, ratings},
		allSubGenerators,
		WithVersion("v1"),
		ForService(productpage, Call(HTTP(), reviews), ConnectToGateway(GatewayHost)),
		ForService(reviews, Call(HTTP(), ratings)),
		GatewayOnHost(GatewayHost),
	)
}

// TestScenario1GRPCThreeServicesInSequence is a basic test setup with a few services
// calling each other in a chain over grpc. Similar to the original bookinfo example setup.
func TestScenario1GRPCThreeServicesInSequence(out Printer) {
	productpage := Entry{"productpage", "Deployment", Namespace}
	reviews := Entry{"reviews", "Deployment", Namespace}
	ratings := Entry{"ratings", "Deployment", Namespace}

	Generate(
		out,
		[]Entry{productpage, reviews, ratings},
		allSubGenerators,
		WithVersion("v1"),
		ForService(productpage, Call(GRPC(), reviews), ConnectToGateway(GatewayHost)),
		ForService(reviews, Call(GRPC(), ratings)),
		GatewayOnHost(GatewayHost),
	)
}

// TestScenario2ThreeServicesInSequenceDeploymentConfig is a basic test setup with a
// few services calling each other in a chain. Similar to the original bookinfo example setup.
// Using DeploymentConfig.
func TestScenario2ThreeServicesInSequenceDeploymentConfig(out Printer) {
	productpage := Entry{"productpage", "DeploymentConfig", Namespace}
	reviews := Entry{"reviews", "DeploymentConfig", Namespace}
	ratings := Entry{"ratings", "DeploymentConfig", Namespace}

	Generate(
		out,
		[]Entry{productpage, reviews, ratings},
		allSubGenerators,
		WithVersion("v1"),
		ForService(productpage, Call(HTTP(), reviews), ConnectToGateway(GatewayHost)),
		ForService(reviews, Call(HTTP(), ratings)),
		GatewayOnHost(GatewayHost),
	)
}

// DemoScenario is a simple setup for demo purposes.
func DemoScenario(out Printer) {
	productpage := Entry{"productpage", "Deployment", Namespace}
	reviews := Entry{"reviews", "Deployment", Namespace}
	ratings := Entry{"ratings", "Deployment", Namespace}
	authors := Entry{"authors", "Deployment", Namespace}
	locations := Entry{"locations", "Deployment", Namespace}

	Generate(
		out,
		[]Entry{productpage, reviews, ratings, authors, locations},
		allSubGenerators,
		WithVersion("v1"),
		ForService(productpage, Call(HTTP(), reviews), Call(HTTP(), authors), ConnectToGateway("ike-demo.io")),
		ForService(reviews, Call(GRPC(), ratings)),
		ForService(authors, Call(GRPC(), locations)),
		GatewayOnHost("ike-demo.io"),
	)
}

// IncompleteMissingVirtualServices generates a scenario where there are no DestinationRules.
func IncompleteMissingDestinationRules(out Printer) {
	productpage := Entry{"productpage", "Deployment", Namespace}
	reviews := Entry{"reviews", "Deployment", Namespace}
	ratings := Entry{"ratings", "Deployment", Namespace}

	Generate(
		out,
		[]Entry{productpage, reviews, ratings},
		[]SubGenerator{Deployment, DeploymentConfig, Service, VirtualService},
		WithVersion("v1"),
		ForService(productpage, Call(HTTP(), reviews), ConnectToGateway(GatewayHost)),
		ForService(reviews, Call(HTTP(), ratings)),
		GatewayOnHost(GatewayHost),
	)
}

// IncompleteMissingVirtualServices generates a scenario where there are no VirtualServices.
func IncompleteMissingVirtualServices(out Printer) {
	productpage := Entry{"productpage", "Deployment", Namespace}
	reviews := Entry{"reviews", "Deployment", Namespace}
	ratings := Entry{"ratings", "Deployment", Namespace}

	Generate(
		out,
		[]Entry{productpage, reviews, ratings},
		[]SubGenerator{Deployment, DeploymentConfig, Service, DestinationRule},
		WithVersion("v1"),
		ForService(productpage, Call(HTTP(), reviews), ConnectToGateway(GatewayHost)),
		ForService(reviews, Call(HTTP(), ratings)),
		GatewayOnHost(GatewayHost),
	)
}
