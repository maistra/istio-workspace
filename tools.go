// +build tools

package tools

// nolint
import (
	_ "k8s.io/code-generator"
	_ "golang.org/x/tools/cmd/goimports"
	_ "github.com/onsi/ginkgo/ginkgo"
    _ "github.com/go-bindata/go-bindata"
	_ "github.com/golang/protobuf/protoc-gen-go"
	_ "github.com/mikefarah/yq/v3"
)
