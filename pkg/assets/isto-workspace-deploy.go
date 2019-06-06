// Package assets Code generated by go-bindata. (@generated) DO NOT EDIT.
// sources:
// deploy/istio-workspace/crds/istio_v1alpha1_session_cr.yaml
// deploy/istio-workspace/crds/istio_v1alpha1_session_crd.yaml
// deploy/istio-workspace/operator.yaml
// deploy/istio-workspace/role.yaml
// deploy/istio-workspace/role_binding.yaml
// deploy/istio-workspace/service_account.yaml
package assets

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name return file name
func (fi bindataFileInfo) Name() string {
	return fi.name
}

// Size return file size
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}

// Mode return file mode
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}

// Mode return file modify time
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir return file whether a directory
func (fi bindataFileInfo) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys return file is sys mode
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _deployIstioWorkspaceCrdsIstio_v1alpha1_session_crYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x2c\x8b\x41\xae\x83\x30\x0c\x44\xf7\x39\x85\x2f\x10\xbe\xd8\xe6\x1a\x5f\xea\xde\x22\x83\x70\x0b\x76\x84\x4d\xa4\xde\xbe\x4a\xcb\x6e\xf4\xde\x3c\x6e\xf2\xc0\xe9\x62\x5a\x48\x3c\xc4\x26\x6b\x50\xdf\x64\x8d\x69\xb1\xe3\xaf\xcf\xbc\xb7\x8d\xe7\xf4\x12\xad\x85\xfe\xe1\xe3\x9b\x0e\x04\x57\x0e\x2e\x89\x48\xf9\x40\xa1\x80\x47\xf6\x5b\x7b\xc3\x32\xd4\x69\x57\x60\x0c\xa2\x78\x37\x14\xda\xc0\x15\xe7\x17\xfc\x32\x68\xcd\x97\xdf\xa8\xf3\x7e\xa1\xd0\x93\xdd\x74\xd4\x58\x47\x9b\xa9\x22\x58\x76\xcf\x7d\x4e\x9f\x00\x00\x00\xff\xff\xd4\x2d\x6a\x39\xb0\x00\x00\x00")

func deployIstioWorkspaceCrdsIstio_v1alpha1_session_crYamlBytes() ([]byte, error) {
	return bindataRead(
		_deployIstioWorkspaceCrdsIstio_v1alpha1_session_crYaml,
		"deploy/istio-workspace/crds/istio_v1alpha1_session_cr.yaml",
	)
}

func deployIstioWorkspaceCrdsIstio_v1alpha1_session_crYaml() (*asset, error) {
	bytes, err := deployIstioWorkspaceCrdsIstio_v1alpha1_session_crYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "deploy/istio-workspace/crds/istio_v1alpha1_session_cr.yaml", size: 176, mode: os.FileMode(436), modTime: time.Unix(1559841267, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _deployIstioWorkspaceCrdsIstio_v1alpha1_session_crdYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x6c\x90\x3d\x4e\x04\x31\x0c\x85\xfb\x9c\xc2\x27\x18\x34\x1d\x4a\x0b\x1d\x88\x02\x24\x7a\x6f\xc6\xbb\x6b\x6d\x26\xb6\x62\x67\x84\x84\xb8\x3b\xca\x64\xf9\x29\xb6\xcc\x97\x4f\xef\xbd\x04\x95\xdf\xa9\x1a\x4b\x89\x80\xca\xf4\xe1\x54\xfa\xc9\xa6\xcb\xbd\x4d\x2c\x77\xdb\x7c\x20\xc7\x39\x5c\xb8\x2c\x11\x1e\x9a\xb9\xac\xaf\x64\xd2\x6a\xa2\x47\x3a\x72\x61\x67\x29\x61\x25\xc7\x05\x1d\x63\x00\x28\xb8\x52\x04\x23\x1b\x41\x6c\xce\x32\x89\x52\xb1\x33\x1f\x7d\x4a\xb2\x06\x53\x4a\x5d\x3d\x55\x69\x1a\xe1\x96\x32\x72\xac\x5b\x00\xa3\xfd\x6d\x44\xee\x24\xb3\xf9\xd3\x7f\xfa\xcc\xe6\xfb\x8d\xe6\x56\x31\xff\x0d\xd8\xa1\x71\x39\xb5\x8c\xf5\x17\x07\x00\x4b\xa2\x14\xe1\xa5\xd7\x28\x26\x5a\x02\xc0\xf6\xf3\x19\xdb\x8c\x59\xcf\x38\x77\xaf\x1d\xea\xf5\xc5\xd7\x39\xe6\xe8\xcd\x22\x7c\x7e\x85\xef\x00\x00\x00\xff\xff\x66\x1b\x53\x70\x41\x01\x00\x00")

func deployIstioWorkspaceCrdsIstio_v1alpha1_session_crdYamlBytes() ([]byte, error) {
	return bindataRead(
		_deployIstioWorkspaceCrdsIstio_v1alpha1_session_crdYaml,
		"deploy/istio-workspace/crds/istio_v1alpha1_session_crd.yaml",
	)
}

func deployIstioWorkspaceCrdsIstio_v1alpha1_session_crdYaml() (*asset, error) {
	bytes, err := deployIstioWorkspaceCrdsIstio_v1alpha1_session_crdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "deploy/istio-workspace/crds/istio_v1alpha1_session_crd.yaml", size: 321, mode: os.FileMode(436), modTime: time.Unix(1554454542, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _deployIstioWorkspaceOperatorYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x84\x53\xc1\x6e\xe2\x30\x10\xbd\xe7\x2b\x46\xa8\x57\xe8\xf6\x58\xdf\x22\xf0\xb2\xa8\x2d\x89\x42\xd4\xd5\x9e\xd0\xd4\x19\x8a\x37\x8e\x9d\xb5\x1d\x2a\x54\xf1\xef\xab\x40\x28\x34\x04\x32\xb7\x3c\x3f\xbf\x79\x33\x7e\xc9\xa5\xce\x18\xa4\x54\x94\x0a\x3d\x05\x58\xca\x57\xb2\x4e\x1a\xcd\xc0\x37\xe0\xc8\x94\xa4\xdd\x5a\xae\xfc\x48\x9a\xfb\xcd\x43\x50\xa2\xc5\x82\x3c\x59\xc7\x02\x80\x21\x68\x2c\x88\xc1\xec\x89\x2f\x27\xd1\xf8\x89\x27\xcb\x84\x4f\x67\x8b\x34\xf9\x13\x00\x00\x58\xfa\x57\x49\x4b\x19\x03\x6f\x2b\xda\x43\x1b\x54\x15\x31\xc8\x8c\xc8\xc9\x8e\xa4\xb9\xa6\x12\x47\x8b\x59\x1a\xf5\xea\xa0\x53\x98\xe7\xba\x72\x9e\x74\x4b\x6a\xf6\x12\x4e\xf9\x72\x1e\xbe\xf0\x1e\x0d\xe9\xbc\x34\xc3\x0f\x63\x73\x57\xa2\xa0\x4e\x99\x34\x9c\xf6\xa8\xd4\xeb\x72\xfe\xec\x72\xca\x9f\x79\x9c\xf0\x05\x9f\x8f\xf9\xf2\x95\x27\x8b\x59\x34\xef\xd1\x18\xfc\x18\x3d\x3e\x0e\x02\xf3\xf6\x97\x84\x6f\x36\x7c\x78\xa5\x09\x95\xca\x6c\x0b\xd2\x7e\xcf\x3f\x7f\x2b\x2c\x4b\x57\x3f\x4d\x8d\x17\xe4\x31\x43\x8f\x6c\xff\x05\x8d\x93\xcb\x01\x01\x5c\x49\xe2\xc8\xb2\x54\x2a\x29\xd0\x31\x78\x68\x10\x47\x8a\x84\x37\xf6\xc8\x00\x28\xd0\x8b\xf5\x33\xbe\x91\x72\x27\xf0\x56\x03\xf8\x0a\xd1\x99\x48\xcb\x5e\x5d\xea\x42\xf3\xb6\xea\x77\xeb\x07\xb3\x76\x23\x05\x85\x42\x98\x4a\xfb\xf9\xcd\xbb\x00\xc2\x68\x8f\x52\x37\x09\x3e\xd5\xb0\xa7\xeb\xa1\x64\x81\xef\xc4\xe0\xee\xb3\x23\xf2\xbb\xfb\x16\x7c\xcc\xf0\xf1\xe0\x94\xc8\x1d\x3b\x47\xd2\x70\xba\x6b\xf5\x11\xa6\x28\x50\x67\xac\x05\xd7\x36\x65\xde\x36\x85\xf6\xdd\x75\x31\xeb\xc5\x74\x0e\x10\x57\x4a\xc5\x46\x49\xb1\x65\x10\xaa\x0f\xdc\xba\x16\x8b\xf4\xa6\x4b\xf0\xb0\xa1\xdf\x61\x3a\xfe\xb5\x1f\x63\x11\x87\x63\x7e\xc1\x3b\xc5\x79\x70\x55\x23\x8e\x26\xa7\x5f\xb3\xe3\xf2\x4f\x6b\x8a\x4b\x07\x75\xad\x24\xa9\x2c\xa1\x55\xf7\x69\x73\x1e\xa3\x5f\xb3\xaf\xb8\x8d\xea\x9e\x57\xad\x44\x31\x4f\xc2\x34\x4a\x6e\xfa\x61\x30\x68\x05\xe3\xfa\x6c\x57\x7f\xfc\x4e\xdd\xbb\xcf\x2e\xfe\x2e\xf8\x1f\x00\x00\xff\xff\xd5\x8c\xe7\x81\x9e\x05\x00\x00")

func deployIstioWorkspaceOperatorYamlBytes() ([]byte, error) {
	return bindataRead(
		_deployIstioWorkspaceOperatorYaml,
		"deploy/istio-workspace/operator.yaml",
	)
}

func deployIstioWorkspaceOperatorYaml() (*asset, error) {
	bytes, err := deployIstioWorkspaceOperatorYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "deploy/istio-workspace/operator.yaml", size: 1438, mode: os.FileMode(436), modTime: time.Unix(1559822028, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _deployIstioWorkspaceRoleYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x8c\x51\x3d\xaf\xdb\x30\x0c\xdc\xfd\x2b\x84\x2c\x01\x0a\xc4\x45\xb7\xc2\x6b\x87\xee\x45\xd1\x9d\x91\x2f\x09\x11\x49\x14\x48\xca\xc5\x7b\xbf\xfe\xc1\x8e\x93\x21\xce\x90\x4d\x27\x9d\xee\x83\xa4\xca\xff\xa0\xc6\x52\x86\xa0\x47\x8a\x3d\x35\xbf\x88\xf2\x27\x39\x4b\xe9\xaf\x3f\xad\x67\xf9\x3e\xfd\xe8\xae\x5c\xc6\x21\xfc\x4a\xcd\x1c\xfa\x47\x12\xba\x0c\xa7\x91\x9c\x86\x2e\x84\xa8\x58\x3e\xfc\xe5\x0c\x73\xca\x75\x08\xa5\xa5\xd4\x85\x50\x28\x63\x08\x6c\xce\x72\xf8\x2f\x7a\xb5\x4a\x11\x9d\xb6\x04\x1b\xba\x43\xa0\xca\xbf\x55\x5a\xb5\x59\xe5\x10\x76\xbb\x2e\x04\x85\x49\xd3\x88\xf5\xae\xca\x68\xcb\xc1\xa0\x13\x47\xdc\x00\xca\x58\x85\x8b\xdf\x50\x9d\x3b\x98\xa3\xf8\x24\xa9\x65\xc4\x44\x9c\x57\xe2\x84\x3b\x2b\x4a\x39\xf1\x39\x53\xbd\xeb\x45\xc5\xf2\x34\x41\x8f\xab\xdb\xfe\xdb\xfe\xbd\x58\x73\xb1\xa5\xcc\x93\xc0\x19\xbe\x15\xa0\xba\x78\x3e\x49\x8c\xa8\x49\x3e\xf2\x23\xdf\x48\xc8\x52\x0c\x2b\x54\xd4\xc4\x91\x1e\xd8\x9c\x1c\xa7\x96\xec\xbd\xd0\x59\x0a\xbb\x28\x97\x73\x1f\x45\x21\xd6\x47\xc9\xdb\x10\xeb\x54\x57\xf6\x8b\x32\xcb\xe4\xe6\xfd\x62\xeb\xb1\xec\xb5\x97\x8a\x62\x17\x3e\xf9\x6b\x87\x39\xdd\x26\xee\x57\x00\x00\x00\xff\xff\x31\x33\xc8\x3a\x79\x02\x00\x00")

func deployIstioWorkspaceRoleYamlBytes() ([]byte, error) {
	return bindataRead(
		_deployIstioWorkspaceRoleYaml,
		"deploy/istio-workspace/role.yaml",
	)
}

func deployIstioWorkspaceRoleYaml() (*asset, error) {
	bytes, err := deployIstioWorkspaceRoleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "deploy/istio-workspace/role.yaml", size: 633, mode: os.FileMode(436), modTime: time.Unix(1559841267, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _deployIstioWorkspaceRole_bindingYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x8c\xd0\xb1\x4e\x03\x31\x0c\x06\xe0\x3d\x4f\xe1\x81\xb5\x87\xd8\x50\xb6\x52\x55\x4c\x20\xd4\x22\x76\x37\xe7\x52\x73\x77\x71\x70\x9c\x43\x80\x78\x77\xc4\xe5\xa0\x48\x48\xc0\x68\x27\xf9\xf3\xd9\x1d\xc7\xd6\xc3\x2d\x0d\xa9\x47\x23\x87\x89\xef\x48\x33\x4b\xf4\x60\x73\xb3\x91\x44\x31\x1f\x78\x6f\x0d\xcb\xe9\x78\xe6\x12\x2a\x0e\x64\xa4\xd9\x3b\x80\x05\x44\x1c\xc8\xc3\xf5\xf2\x6a\xbd\xbd\x59\xae\xd6\x0e\x00\x40\xe9\xb1\xb0\x52\xeb\xc1\xb4\xd0\xd4\x1a\xb1\x2f\xe4\x81\xb3\xb1\x2c\xf2\x73\x36\x1a\x9c\xec\x1e\x28\xd8\x9c\x53\x2d\xab\xbe\x64\x23\xdd\x48\x4f\x17\x1c\x5b\x8e\xf7\xd3\xeb\xef\x32\xdd\x61\x68\xb0\xd8\x41\x94\x5f\xd0\x58\x62\xd3\x9d\xe7\x59\xf7\x71\x79\x20\xc3\x16\x0d\xfd\x54\xc1\x2c\xac\x3f\x3f\x89\x76\x39\x61\xa8\xa8\x5c\x8e\x82\xa3\x61\x4b\x3a\x72\xa0\x65\x08\x52\xa2\xfd\x19\x52\xcf\xa6\xda\xc3\xc9\xeb\xd7\x26\xde\xea\x2a\xa4\xa7\x0d\xed\x3f\x2d\x3f\xa6\xfc\x47\x3c\x26\xbe\x54\x29\xe9\x97\xd1\xdd\x7b\x00\x00\x00\xff\xff\x41\x8a\x0b\xd7\xca\x01\x00\x00")

func deployIstioWorkspaceRole_bindingYamlBytes() ([]byte, error) {
	return bindataRead(
		_deployIstioWorkspaceRole_bindingYaml,
		"deploy/istio-workspace/role_binding.yaml",
	)
}

func deployIstioWorkspaceRole_bindingYaml() (*asset, error) {
	bytes, err := deployIstioWorkspaceRole_bindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "deploy/istio-workspace/role_binding.yaml", size: 458, mode: os.FileMode(436), modTime: time.Unix(1559841267, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _deployIstioWorkspaceService_accountYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x04\xc0\x31\x0e\x80\x30\x08\x00\xc0\x9d\x57\xf0\x01\x07\x57\x36\xdf\x60\xe2\x4e\x28\x03\x69\x0a\x4d\xc1\xfa\x7d\x8f\xa7\x3d\xba\xd2\xc2\x09\xf7\x09\xdd\xbc\x11\xde\xba\xb6\x89\x5e\x22\xf1\x7a\xc1\xd0\xe2\xc6\xc5\x04\x88\xce\x43\x09\x2d\xcb\xe2\xf8\x62\xf5\x9c\x2c\x0a\x7f\x00\x00\x00\xff\xff\x94\xa0\xb7\x3f\x46\x00\x00\x00")

func deployIstioWorkspaceService_accountYamlBytes() ([]byte, error) {
	return bindataRead(
		_deployIstioWorkspaceService_accountYaml,
		"deploy/istio-workspace/service_account.yaml",
	)
}

func deployIstioWorkspaceService_accountYaml() (*asset, error) {
	bytes, err := deployIstioWorkspaceService_accountYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "deploy/istio-workspace/service_account.yaml", size: 70, mode: os.FileMode(436), modTime: time.Unix(1554454542, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"deploy/istio-workspace/crds/istio_v1alpha1_session_cr.yaml":  deployIstioWorkspaceCrdsIstio_v1alpha1_session_crYaml,
	"deploy/istio-workspace/crds/istio_v1alpha1_session_crd.yaml": deployIstioWorkspaceCrdsIstio_v1alpha1_session_crdYaml,
	"deploy/istio-workspace/operator.yaml":                        deployIstioWorkspaceOperatorYaml,
	"deploy/istio-workspace/role.yaml":                            deployIstioWorkspaceRoleYaml,
	"deploy/istio-workspace/role_binding.yaml":                    deployIstioWorkspaceRole_bindingYaml,
	"deploy/istio-workspace/service_account.yaml":                 deployIstioWorkspaceService_accountYaml,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"deploy": &bintree{nil, map[string]*bintree{
		"istio-workspace": &bintree{nil, map[string]*bintree{
			"crds": &bintree{nil, map[string]*bintree{
				"istio_v1alpha1_session_cr.yaml":  &bintree{deployIstioWorkspaceCrdsIstio_v1alpha1_session_crYaml, map[string]*bintree{}},
				"istio_v1alpha1_session_crd.yaml": &bintree{deployIstioWorkspaceCrdsIstio_v1alpha1_session_crdYaml, map[string]*bintree{}},
			}},
			"operator.yaml":        &bintree{deployIstioWorkspaceOperatorYaml, map[string]*bintree{}},
			"role.yaml":            &bintree{deployIstioWorkspaceRoleYaml, map[string]*bintree{}},
			"role_binding.yaml":    &bintree{deployIstioWorkspaceRole_bindingYaml, map[string]*bintree{}},
			"service_account.yaml": &bintree{deployIstioWorkspaceService_accountYaml, map[string]*bintree{}},
		}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
