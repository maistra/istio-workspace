package generator

import (
	"io"
	"os"
)

var (
	Namespace = ""
)

// TestScenario1HTTPThreeServicesInSequence is a basic test setup with a few services
// calling each other in a chain over http. Similar to the original bookinfo example setup.
func TestScenario1HTTPThreeServicesInSequence(out io.Writer) {
	productpage := Entry{"productpage", "Deployment", Namespace}
	reviews := Entry{"reviews", "Deployment", Namespace}
	ratings := Entry{"ratings", "Deployment", Namespace}

	Generate(
		out,
		[]Entry{productpage, reviews, ratings},
		WithVersion("v1"),
		ForService(productpage, Call(HTTP(), reviews), ConnectToGateway(GatewayHost)),
		ForService(reviews, Call(HTTP(), ratings)),
		GatewayOnHost(GatewayHost),
	)
}

// TestScenario1GRPCThreeServicesInSequence is a basic test setup with a few services
// calling each other in a chain over grpc. Similar to the original bookinfo example setup.
func TestScenario1GRPCThreeServicesInSequence(out io.Writer) {
	productpage := Entry{"productpage", "Deployment", Namespace}
	reviews := Entry{"reviews", "Deployment", Namespace}
	ratings := Entry{"ratings", "Deployment", Namespace}

	Generate(
		out,
		[]Entry{productpage, reviews, ratings},
		WithVersion("v1"),
		ForService(productpage, Call(GRPC(), reviews), ConnectToGateway(GatewayHost)),
		ForService(reviews, Call(GRPC(), ratings)),
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
		ForService(productpage, Call(HTTP(), reviews), ConnectToGateway(GatewayHost)),
		ForService(reviews, Call(HTTP(), ratings)),
		GatewayOnHost(GatewayHost),
	)
}

// TestScenarioMutationHookChe is a basic test setup with a
// few services calling each other in a chain. Similar to the original bookinfo example setup
// and a single Deployment simulating a Che Deployment which should trigger the WebHook.
// Using Deployment.
func TestScenarioMutationHookChe(out io.Writer) {
	productpage := Entry{"productpage", "Deployment", Namespace}
	reviews := Entry{"reviews", "Deployment", Namespace}
	ratings := Entry{"ratings", "Deployment", Namespace}
	che := Entry{"che-workspace", "Deployment", Namespace}

	Generate(
		out,
		[]Entry{productpage, reviews, ratings},
		ForService(productpage, WithVersion("v1")),
		ForService(reviews, WithVersion("v1")),
		ForService(ratings, WithVersion("v1")),
		ForService(productpage, Call(HTTP(), reviews), ConnectToGateway(GatewayHost)),
		ForService(reviews, Call(HTTP(), ratings)),
		GatewayOnHost(GatewayHost),
	)
	Do(out, che, Deployment, WithAnnotations(map[string]string{
		"ike.target":  "reviews-v1",
		"ike.session": os.Getenv("IKE_SESSION"),
		"ike.route":   "header:x-test-suite=smoke"}))
}

// TestScenarioMutationHookCheOnly is a basic test setup with a
// few services calling each other in a chain. Similar to the original bookinfo example setup
// and a single Deployment simulating a Che Deployment which should trigger the WebHook.
// Using Deployment.
func TestScenarioMutationHookCheOnly(out io.Writer) {
	che := Entry{"che-workspace", "Deployment", Namespace}

	Do(out, che, Deployment, WithAnnotations(map[string]string{
		"ike.target":  "reviews-v1",
		"ike.session": os.Getenv("IKE_SESSION"),
		"ike.route":   "header:x-test-suite=smoke"}))
}

// DemoScenario is a simple setup for demo purposes.
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
		ForService(productpage, Call(HTTP(), reviews), Call(HTTP(), authors), ConnectToGateway("ike-demo.io")),
		ForService(reviews, Call(GRPC(), ratings)),
		ForService(authors, Call(GRPC(), locations)),
		GatewayOnHost("ike-demo.io"),
	)
}
