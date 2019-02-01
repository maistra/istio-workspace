package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/aslakknutsen/istio-workspace/pkg/stub"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/operator-framework/operator-sdk/pkg/util/k8sutil"
	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ike",
	Short: "ike lets you safely develop and test on prod without a sweat",
	Long: `Lorem ipsum dolor sit amet, consectetur adipiscing elit. 
Aliquam vitae dolor neque. Aliquam facilisis posuere nulla sit amet porttitor. 

Duis nec interdum velit, id consectetur erat. In tempor tempor turpis vel rhoncus.`,

	Run: func(cmd *cobra.Command, args []string) {
		printVersion()
		startOperator()
	},
}

func startOperator() {
	sdk.ExposeMetricsPort()
	resource := "istio.openshift.com/v1alpha1"
	kind := "Session"
	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		logrus.Fatalf("Failed to get watch namespace: %v", err)
	}
	resyncPeriod := 0
	logrus.Infof("Watching resource %s, kind %s, namespace %s, resyncPeriod %d", resource, kind, namespace, resyncPeriod)
	sdk.Watch(resource, kind, namespace, resyncPeriod)
	sdk.Handle(stub.NewHandler("Istio Workspace Handler"))
	sdk.Run(context.TODO())
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
