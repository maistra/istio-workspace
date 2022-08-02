package scenarios

import (
	"github.com/maistra/istio-workspace/pkg/generator"
)

type TestScenario func(string, string, generator.Printer)

// TestScenarioHTTPThreeServicesInSequence is a basic test setup with a few services
// calling each other in a chain over http. Similar to the original bookinfo example setup.
func TestScenarioHTTPThreeServicesInSequence(ns, image string, printer generator.Printer) {
	productpage := generator.NewServiceEntry("productpage", ns, "Deployment", image)
	reviews := generator.NewServiceEntry("reviews", ns, "Deployment", image)
	ratings := generator.NewServiceEntry("ratings", ns, "Deployment", image)

	generator.Generate(
		printer,
		[]generator.ServiceEntry{productpage, reviews, ratings},
		generator.NsGenerators,
		generator.AllSubGenerators,
		generator.WithVersion("v1"),
		generator.GatewayOnHost(generator.GatewayHost),
		generator.ForService(productpage, generator.Call(generator.HTTP(), reviews), generator.ConnectToGateway(generator.GatewayHost)),
		generator.ForService(reviews, generator.Call(generator.HTTP(), ratings)),
	)
}

// TestScenarioGRPCThreeServicesInSequence is a basic test setup with a few services
// calling each other in a chain over grpc. Similar to the original bookinfo example setup.
func TestScenarioGRPCThreeServicesInSequence(ns, image string, printer generator.Printer) {
	productpage := generator.NewServiceEntry("productpage", ns, "Deployment", image)
	reviews := generator.NewServiceEntry("reviews", ns, "Deployment", image)
	ratings := generator.NewServiceEntry("ratings", ns, "Deployment", image)

	generator.Generate(
		printer,
		[]generator.ServiceEntry{productpage, reviews, ratings},
		generator.NsGenerators,
		generator.AllSubGenerators,
		generator.WithVersion("v1"),
		generator.ForService(productpage, generator.Call(generator.GRPC(), reviews), generator.ConnectToGateway(generator.GatewayHost)),
		generator.ForService(reviews, generator.Call(generator.GRPC(), ratings)),
		generator.GatewayOnHost(generator.GatewayHost),
	)
}

// TestScenarioThreeServicesInSequenceWithDeploymentConfig is a basic test setup with a
// few services calling each other in a chain. Similar to the original bookinfo example setup.
// Using DeploymentConfig.
func TestScenarioThreeServicesInSequenceWithDeploymentConfig(ns, image string, printer generator.Printer) {
	productpage := generator.NewServiceEntry("productpage", ns, "DeploymentConfig", image)
	reviews := generator.NewServiceEntry("reviews", ns, "DeploymentConfig", image)
	ratings := generator.NewServiceEntry("ratings", ns, "DeploymentConfig", image)

	generator.Generate(
		printer,
		[]generator.ServiceEntry{productpage, reviews, ratings},
		generator.NsGenerators,
		generator.AllSubGenerators,
		generator.WithVersion("v1"),
		generator.ForService(productpage, generator.Call(generator.HTTP(), reviews), generator.ConnectToGateway(generator.GatewayHost)),
		generator.ForService(reviews, generator.Call(generator.HTTP(), ratings)),
		generator.GatewayOnHost(generator.GatewayHost),
	)
}

// DemoScenario is a simple setup for demo purposes.
func DemoScenario(ns, image string, printer generator.Printer) {
	productpage := generator.NewServiceEntry("productpage", ns, "Deployment", image)
	reviews := generator.NewServiceEntry("reviews", ns, "Deployment", image)
	ratings := generator.NewServiceEntry("ratings", ns, "Deployment", image)
	authors := generator.NewServiceEntry("authors", ns, "Deployment", image)
	locations := generator.NewServiceEntry("locations", ns, "Deployment", image)

	generator.Generate(
		printer,
		[]generator.ServiceEntry{productpage, reviews, ratings, authors, locations},
		generator.NsGenerators,
		generator.AllSubGenerators,
		generator.WithVersion("v1"),
		generator.ForService(productpage, generator.Call(generator.HTTP(), reviews), generator.Call(generator.HTTP(), authors), generator.ConnectToGateway("ike-demo.io")),
		generator.ForService(reviews, generator.Call(generator.GRPC(), ratings)),
		generator.ForService(authors, generator.Call(generator.GRPC(), locations)),
		generator.GatewayOnHost("ike-demo.io"),
	)
}

// IncompleteMissingDestinationRules generates a scenario where there are no DestinationRules.
func IncompleteMissingDestinationRules(ns, image string, printer generator.Printer) {
	productpage := generator.NewServiceEntry("productpage", ns, "Deployment", image)
	reviews := generator.NewServiceEntry("reviews", ns, "Deployment", image)
	ratings := generator.NewServiceEntry("ratings", ns, "Deployment", image)

	generator.Generate(
		printer,
		[]generator.ServiceEntry{productpage, reviews, ratings},
		generator.NsGenerators,
		[]generator.SubGenerator{generator.Deployment, generator.DeploymentConfig, generator.Service, generator.VirtualService},
		generator.WithVersion("v1"),
		generator.ForService(productpage, generator.Call(generator.HTTP(), reviews), generator.ConnectToGateway(generator.GatewayHost)),
		generator.ForService(reviews, generator.Call(generator.HTTP(), ratings)),
		generator.GatewayOnHost(generator.GatewayHost),
	)
}

// IncompleteMissingVirtualServices generates a scenario where there are no VirtualServices.
func IncompleteMissingVirtualServices(ns, image string, printer generator.Printer) {
	productpage := generator.NewServiceEntry("productpage", ns, "Deployment", image)
	reviews := generator.NewServiceEntry("reviews", ns, "Deployment", image)
	ratings := generator.NewServiceEntry("ratings", ns, "Deployment", image)

	generator.Generate(
		printer,
		[]generator.ServiceEntry{productpage, reviews, ratings},
		generator.NsGenerators,
		[]generator.SubGenerator{generator.Deployment, generator.DeploymentConfig, generator.Service, generator.DestinationRule},
		generator.WithVersion("v1"),
		generator.ForService(productpage, generator.Call(generator.HTTP(), reviews), generator.ConnectToGateway(generator.GatewayHost)),
		generator.ForService(reviews, generator.Call(generator.HTTP(), ratings)),
		generator.GatewayOnHost(generator.GatewayHost),
	)
}
