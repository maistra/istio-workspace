//go:build tools
// +build tools

package tools

// nolint
import (
	_ "github.com/go-bindata/go-bindata/v3"
	_ "github.com/golang/protobuf/protoc-gen-go"
	_ "github.com/kisielk/errcheck"
	_ "github.com/mikefarah/yq/v4"
	_ "github.com/onsi/ginkgo/v2/ginkgo"
	_ "github.com/onsi/ginkgo/v2/ginkgo/generators"
	_ "golang.org/x/tools/cmd/goimports"
	_ "k8s.io/code-generator"
	_ "sigs.k8s.io/controller-tools/cmd/controller-gen"
)
