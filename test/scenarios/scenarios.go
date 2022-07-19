package scenarios

import (
	"io"

	"github.com/maistra/istio-workspace/pkg/generator"
)

func BasicNewService(out io.Writer, name, ns string) {
	newService := generator.NewEntry(name, ns, "Deployment")

	generator.Generate(
		out,
		[]generator.Entry{newService},
		generator.AllSubGenerators,
		generator.WithVersion("v1"),
		generator.UsingImage("quay.io/maistra-dev/istio-workspace-test-prepared-prepared-image"),
		generator.ForService(newService, generator.ConnectToGateway(generator.GatewayHost)),
	)
}

// TestScenario1HTTPThreeServicesInSequence is a basic test setup with a few services
// calling each other in a chain over http. Similar to the original bookinfo example setup.
func TestScenario1HTTPThreeServicesInSequence(out io.Writer, ns string) {
	productpage := generator.NewEntry("productpage", ns, "Deployment")
	reviews := generator.NewEntry("reviews", ns, "Deployment")
	ratings := generator.NewEntry("ratings", ns, "Deployment")

	generator.Generate(
		out,
		[]generator.Entry{productpage, reviews, ratings},
		generator.AllSubGenerators,
		generator.WithVersion("v1"),
		generator.ForService(productpage, generator.Call(generator.HTTP(), reviews), generator.ConnectToGateway(generator.GatewayHost)),
		generator.ForService(reviews, generator.Call(generator.HTTP(), ratings)),
		generator.GatewayOnHost(generator.GatewayHost),
	)
}

// TestScenario1GRPCThreeServicesInSequence is a basic test setup with a few services
// calling each other in a chain over grpc. Similar to the original bookinfo example setup.
func TestScenario1GRPCThreeServicesInSequence(out io.Writer, ns string) {
	productpage := generator.NewEntry("productpage", ns, "Deployment")
	reviews := generator.NewEntry("reviews", ns, "Deployment")
	ratings := generator.NewEntry("ratings", ns, "Deployment")

	generator.Generate(
		out,
		[]generator.Entry{productpage, reviews, ratings},
		generator.AllSubGenerators,
		generator.WithVersion("v1"),
		generator.ForService(productpage, generator.Call(generator.GRPC(), reviews), generator.ConnectToGateway(generator.GatewayHost)),
		generator.ForService(reviews, generator.Call(generator.GRPC(), ratings)),
		generator.GatewayOnHost(generator.GatewayHost),
	)
}

// TestScenario2ThreeServicesInSequenceDeploymentConfig is a basic test setup with a
// few services calling each other in a chain. Similar to the original bookinfo example setup.
// Using DeploymentConfig.
func TestScenario2ThreeServicesInSequenceDeploymentConfig(out io.Writer, ns string) {
	productpage := generator.NewEntry("productpage", ns, "DeploymentConfig")
	reviews := generator.NewEntry("reviews", ns, "DeploymentConfig")
	ratings := generator.NewEntry("ratings", ns, "DeploymentConfig")

	generator.Generate(
		out,
		[]generator.Entry{productpage, reviews, ratings},
		generator.AllSubGenerators,
		generator.WithVersion("v1"),
		generator.ForService(productpage, generator.Call(generator.HTTP(), reviews), generator.ConnectToGateway(generator.GatewayHost)),
		generator.ForService(reviews, generator.Call(generator.HTTP(), ratings)),
		generator.GatewayOnHost(generator.GatewayHost),
	)
}

// DemoScenario is a simple setup for demo purposes.
func DemoScenario(out io.Writer, ns string) {
	productpage := generator.NewEntry("productpage", ns, "Deployment")
	reviews := generator.NewEntry("reviews", ns, "Deployment")
	ratings := generator.NewEntry("ratings", ns, "Deployment")
	authors := generator.NewEntry("authors", ns, "Deployment")
	locations := generator.NewEntry("locations", ns, "Deployment")

	generator.Generate(
		out,
		[]generator.Entry{productpage, reviews, ratings, authors, locations},
		generator.AllSubGenerators,
		generator.WithVersion("v1"),
		generator.ForService(productpage, generator.Call(generator.HTTP(), reviews), generator.Call(generator.HTTP(), authors), generator.ConnectToGateway("ike-demo.io")),
		generator.ForService(reviews, generator.Call(generator.GRPC(), ratings)),
		generator.ForService(authors, generator.Call(generator.GRPC(), locations)),
		generator.GatewayOnHost("ike-demo.io"),
	)
}

// IncompleteMissingDestinationRules generates a scenario where there are no DestinationRules.
func IncompleteMissingDestinationRules(out io.Writer, ns string) {
	productpage := generator.NewEntry("productpage", ns, "Deployment")
	reviews := generator.NewEntry("reviews", ns, "Deployment")
	ratings := generator.NewEntry("ratings", ns, "Deployment")

	generator.Generate(
		out,
		[]generator.Entry{productpage, reviews, ratings},
		[]generator.SubGenerator{generator.Deployment, generator.DeploymentConfig, generator.Service, generator.VirtualService},
		generator.WithVersion("v1"),
		generator.ForService(productpage, generator.Call(generator.HTTP(), reviews), generator.ConnectToGateway(generator.GatewayHost)),
		generator.ForService(reviews, generator.Call(generator.HTTP(), ratings)),
		generator.GatewayOnHost(generator.GatewayHost),
	)
}

// IncompleteMissingVirtualServices generates a scenario where there are no VirtualServices.
func IncompleteMissingVirtualServices(out io.Writer, ns string) {
	productpage := generator.NewEntry("productpage", ns, "Deployment")
	reviews := generator.NewEntry("reviews", ns, "Deployment")
	ratings := generator.NewEntry("ratings", ns, "Deployment")

	generator.Generate(
		out,
		[]generator.Entry{productpage, reviews, ratings},
		[]generator.SubGenerator{generator.Deployment, generator.DeploymentConfig, generator.Service, generator.DestinationRule},
		generator.WithVersion("v1"),
		generator.ForService(productpage, generator.Call(generator.HTTP(), reviews), generator.ConnectToGateway(generator.GatewayHost)),
		generator.ForService(reviews, generator.Call(generator.HTTP(), ratings)),
		generator.GatewayOnHost(generator.GatewayHost),
	)
}
