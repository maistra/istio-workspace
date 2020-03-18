package generator

import (
	"io"
)

var (
	Namespace = ""
)

// TestScenario1ThreeServicesInSequence is a basic test setup with a few services
// calling each other in a chain. Similar to the original bookinfo example setup.
func TestScenario1ThreeServicesInSequence(out io.Writer) {
	productpage := Entry{"productpage", "Deployment", Namespace}
	reviews := Entry{"reviews", "Deployment", Namespace}
	ratings := Entry{"ratings", "Deployment", Namespace}

	Generate(
		out,
		[]Entry{productpage, reviews, ratings},
		WithVersion("v1"),
		ForService(productpage, Call(reviews), ConnectToGateway()),
		ForService(reviews, Call(ratings)),
		GatewayOnHost(GatewayHost),
	)
}

// TestScenario2ThreeServicesInSequenceDeploymentConfig is a basic test setup with a
// few services calling each other in a chain. Similar to the original bookinfo example setup.
// Using DeploymentConfig.
func TestScenario2ThreeServicesInSequenceDeploymentConfig(out io.Writer) {
	productpage := Entry{"productpage", "DeploymentConfig", Namespace}
	reviews := Entry{"reviews", "DeploymentConfig", Namespace}
	ratings := Entry{"ratings", "DeploymentConfig", Namespace}

	Generate(
		out,
		[]Entry{productpage, reviews, ratings},
		WithVersion("v1"),
		ForService(productpage, Call(reviews), ConnectToGateway()),
		ForService(reviews, Call(ratings)),
		GatewayOnHost(GatewayHost),
	)
}

// DemoScenario is a simple setup for demo purposes
func DemoScenario(out io.Writer) {
	productpage := Entry{"productpage", "Deployment", Namespace}
	reviews := Entry{"reviews", "Deployment", Namespace}
	ratings := Entry{"ratings", "Deployment", Namespace}
	authors := Entry{"authors", "Deployment", Namespace}
	locations := Entry{"locations", "Deployment", Namespace}

	Generate(
		out,
		[]Entry{productpage, reviews, ratings, authors, locations},
		WithVersion("v1"),
		ForService(productpage, Call(reviews), Call(authors), ConnectToGateway()),
		ForService(reviews, Call(ratings)),
		ForService(authors, Call(locations)),
		GatewayOnHost("ike-demo.io"), GatewayOnHost("*.ike-demo.io"),
	)
}
