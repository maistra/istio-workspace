// Package assets Code generated by go-bindata. (@generated) DO NOT EDIT.
// sources:
// deploy/istio-workspace/operator.yaml
// deploy/istio-workspace/role_local.yaml
// deploy/istio-workspace/service_account.yaml
// deploy/istio-workspace/role_binding.yaml
// deploy/istio-workspace/role_binding_local.yaml
// deploy/istio-workspace/crds/istio_v1alpha1_session_crd.yaml
// deploy/istio-workspace/role.yaml
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
		return nil, fmt.Errorf("read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("read %q: %v", name, err)
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

// ModTime return file modify time
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

var _deployIstioWorkspaceOperatorYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x8c\x94\xc1\x8e\xe2\x38\x10\x86\xef\x3c\x45\x29\x9a\x6b\x33\xd3\xd7\xdc\xa2\xee\x2c\x8b\xa6\x07\xa2\x90\x9d\xd5\x9e\x50\xe1\x14\xc4\x8b\x63\x7b\x6c\x87\x28\x1a\xf1\xee\x2b\x87\xd0\x40\x3a\xe9\xac\x6f\x94\x2b\xf5\x7f\xf5\x53\xae\x23\x97\x79\x08\x19\x95\x5a\xa0\xa3\x19\x6a\xfe\x93\x8c\xe5\x4a\x86\xe0\xba\xe0\x5c\x69\x92\xb6\xe0\x7b\x37\xe7\xea\xeb\xe9\x79\xa6\xd1\x60\x49\x8e\x8c\x0d\x67\x00\x4f\x20\xb1\xa4\x10\x96\xdf\xe3\xed\xeb\xfa\xe5\x7b\x9c\x6e\xd3\x78\xb1\xdc\x64\xe9\x3f\x33\x00\x80\x9c\x2c\x33\x5c\xbb\xb6\x66\xf0\xaa\xd8\x91\x0c\x18\x3a\x70\xeb\x4c\x03\x75\x41\x86\x20\x27\x2d\x54\x43\x39\xf0\x12\x0f\x04\xdc\x02\x9e\x90\x0b\xdc\x09\x0a\xda\x22\x86\x7e\x55\xdc\x50\x1e\x82\x33\x15\xb5\xa1\x13\x8a\x8a\x42\xf8\x55\x61\x33\xe7\x6a\x0c\x24\x59\x6f\x96\xd9\x7a\x10\x25\x25\xad\x2c\x77\xca\x34\xc0\x25\xd4\x05\x67\x05\xb8\x82\x3a\x06\x86\x12\x76\x04\x7b\x55\xc9\x7c\x8a\xa1\x44\xdf\x0c\xf6\x18\x96\x3f\xa2\x45\xbc\x5d\x45\x3f\xe2\x01\xf1\xac\xa0\x36\x15\xd4\xfe\x4e\xb4\xe6\xae\x00\x7e\x24\xd8\x71\x89\xa6\x99\xd2\xe5\xd6\x71\xf5\x54\x2b\x73\xb4\x1a\x19\x0d\xea\x67\xd1\x62\x44\xde\xe1\xe1\x51\xdd\x29\xdf\x71\x65\x69\xb2\x61\x3f\x16\xd6\xdd\xe9\xfd\x1d\x65\x2f\x7f\xb6\xbd\x6e\x92\xe8\xe5\xb3\x86\x5b\x52\x2f\x55\xa3\x63\xc5\x1c\xde\x08\x4f\x04\x54\x6a\xd7\xc0\x5e\x19\x60\xa2\xb2\x8e\x0c\xd4\x3c\xa7\x2e\xa7\x4f\xb3\x47\x61\x1f\x70\x82\xe0\x0e\x25\x8b\xdf\xe2\x24\x8d\x37\xf1\xea\x25\xde\xfe\x8c\xd3\xcd\x72\xbd\x1a\xe0\xf9\x4b\xe6\x64\x44\xc3\xe5\x01\x32\x12\xa4\x0d\x59\x92\x8c\xe0\x74\x19\xff\xd6\x06\x8f\x79\x19\x4d\xe0\xce\x5e\x5c\xb2\x7e\x64\x25\xd8\x1a\xb5\xf6\x1f\x5f\xee\x4b\x92\x3e\x41\xb6\x6e\x76\x2d\x4c\x99\x18\x7c\x9b\x3f\x7f\x7b\x0e\x66\x6a\xf7\x2f\x31\xd7\xbd\xa5\xcb\x7b\x7c\x7d\x2f\xda\x7e\x70\xff\x2a\x51\x6b\xeb\x1f\xa1\x8f\x97\xe4\x30\x47\x87\x61\xfb\x0b\x3a\x03\x3e\x4e\x85\x3f\x02\x77\x24\xec\x35\xd3\xd7\xd4\x63\xa9\x70\x35\x21\x84\x2f\xbf\xfd\x24\x75\x2e\x9e\xdb\x7b\xab\x89\x5d\xcb\x18\xd2\x82\x33\xb4\x21\x3c\x77\x11\x4b\x82\x98\x53\xe6\x26\x54\xfa\xff\xf0\xad\xa7\x3e\xa1\xff\x39\x01\xbc\xaf\xa6\x3b\x95\x9e\x15\x53\x76\x0c\x59\xf2\x3f\xb0\xa6\xc0\x1e\xed\xb9\x18\x62\x4e\x9c\x51\xc4\x98\xaa\xa4\x5b\x4d\x20\x31\x25\x1d\x72\xd9\xad\xd6\xdb\x79\x9a\x6c\xc6\x9f\x76\x40\xaf\x64\xbd\x5d\x7c\xfe\xda\x0b\x5f\x37\xe3\xf5\xe2\xb6\xae\xce\xe1\x7d\x24\x8b\x16\xe7\x9e\x0e\x53\x65\x89\x32\x0f\x7b\x61\x8f\xc9\x8f\x7d\x28\x34\x07\x3b\x94\xe9\x8d\x19\x6c\x20\xa9\x84\x48\x94\xe0\xac\x09\x21\x12\x35\x36\xb6\x97\x45\xf2\x34\x54\x70\x7c\x13\x3d\x9e\xeb\xf3\xfb\xf2\xbb\x97\x7b\x0e\x46\xab\x26\xeb\xd7\xdb\x26\x1f\x28\xf7\x87\x51\xe5\x47\x26\x7f\xf6\x9c\x44\x9e\xd2\x7e\xf8\xb6\xbb\x4f\xd0\x15\xe1\xfb\x0c\xcf\xbd\xe6\x28\xca\x3a\x89\xd3\x28\x5b\xa7\x9f\xf2\x84\x10\xf4\x46\x65\xbc\xb7\xd1\x85\x39\x66\xdb\xd0\x07\xe7\x60\xf6\x5f\x00\x00\x00\xff\xff\xfb\x50\x5f\x03\x4b\x08\x00\x00")

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

	info := bindataFileInfo{name: "deploy/istio-workspace/operator.yaml", size: 2123, mode: os.FileMode(436), modTime: time.Unix(1565641417, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _deployIstioWorkspaceRole_localYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x9c\x53\xcf\x6a\xf3\x30\x0c\xbf\xe7\x29\x4c\x2f\x85\x0f\x9a\x8f\xdd\x46\x5e\x60\xf7\x31\x76\x57\x1d\xb5\x15\xb5\x2d\x23\x29\x19\xdb\xd3\x0f\xa7\x61\x75\xa1\xe9\xc8\x72\x8a\x15\xe9\xf7\x2f\x32\x64\x7a\x47\x51\xe2\xd4\x39\xd9\x83\x6f\x61\xb0\x13\x0b\x7d\x81\x11\xa7\xf6\xfc\xac\x2d\xf1\xff\xf1\xa9\x39\x53\xea\x3b\xf7\xca\x01\x9b\x88\x06\x3d\x18\x74\x8d\x73\x5e\x70\xea\x7c\xa3\x88\x6a\x10\x73\xe7\xd2\x10\x42\xe3\x5c\x82\x88\x9d\x23\x35\xe2\xdd\x07\xcb\x59\x33\x78\x6c\x64\x08\xa8\x65\x70\xe7\x20\xd3\x8b\xf0\x90\xa7\x63\x79\x76\x6e\xb3\x99\x5e\x05\x95\x07\xf1\x58\x7d\xc9\xdc\xeb\xcf\x41\x51\x46\xf2\x78\x2d\x60\xea\x33\x53\xb2\x6b\x25\x17\x53\x6a\x98\x6c\xe4\x30\x44\xf4\x01\x28\x56\x03\x23\xd6\xdd\x9e\xd3\x81\x8e\x11\x72\xcd\xe1\x05\xe7\x96\x11\x65\x5f\x69\xd9\xfe\xdb\xae\x37\x50\xe2\x98\x22\xb8\x0b\x79\x44\x5b\x82\x84\x3c\xab\xba\x03\xda\x63\x0e\xfc\x19\x6f\xbc\xf4\x80\x91\x93\x62\x55\x12\xcc\x81\x3c\xdc\xd4\xd4\xc0\xf0\x30\x04\x5d\x6f\xb2\x28\x6a\x39\x63\xd2\x13\x1d\xac\x25\xfe\x5d\xde\x25\xe0\xb5\x44\x91\x13\x19\x0b\xa5\x63\xeb\x59\x90\xb5\xf5\x1c\x97\xc8\xe6\xa5\x98\x67\x1e\xa4\x3c\xff\xf2\xb2\xb8\xb8\xc4\x3c\xad\x6d\xe5\xf1\x01\xef\x45\xff\x1a\x5b\x09\xad\x5c\x88\x62\xeb\xc2\xb3\x9c\xe0\x7a\xf0\x08\xa4\x26\xf0\x37\xcc\xef\x00\x00\x00\xff\xff\x7a\xce\x78\xc1\x0e\x04\x00\x00")

func deployIstioWorkspaceRole_localYamlBytes() ([]byte, error) {
	return bindataRead(
		_deployIstioWorkspaceRole_localYaml,
		"deploy/istio-workspace/role_local.yaml",
	)
}

func deployIstioWorkspaceRole_localYaml() (*asset, error) {
	bytes, err := deployIstioWorkspaceRole_localYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "deploy/istio-workspace/role_local.yaml", size: 1038, mode: os.FileMode(436), modTime: time.Unix(1565641417, 0)}
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

	info := bindataFileInfo{name: "deploy/istio-workspace/service_account.yaml", size: 70, mode: os.FileMode(436), modTime: time.Unix(1561716969, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _deployIstioWorkspaceRole_bindingYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x84\x90\x41\x6f\xdb\x30\x0c\x85\xef\xfe\x15\x44\xb0\x6b\x3c\xec\x36\xe8\x96\x05\xc1\x4e\x1b\x86\x24\xd8\x9d\x96\xe8\x98\x8b\x2d\x6a\x14\x95\x00\x2d\xfa\xdf\x8b\x28\x4e\x5b\xb4\x40\xaa\x9b\xf8\x88\xc7\xef\xbd\x23\xc7\xe0\x60\x4f\x53\x1a\xd1\xa8\xc1\xc4\x7f\x49\x33\x4b\x74\x60\xf3\xb0\x95\x44\x31\x0f\xdc\x5b\xcb\xf2\xf5\xf4\xad\x49\xa8\x38\x91\x91\x66\xd7\x00\x2c\x21\xe2\x44\x0e\x7e\xaf\x7e\x6d\x76\x7f\x56\xeb\x4d\x03\x00\x10\x28\x7b\xe5\x64\xd5\x69\xb1\x47\x3d\x90\xd5\xc5\x9c\xd0\x13\xf4\xa2\x70\x1e\xd8\x0f\xa0\x32\x12\x74\x1c\x03\xc7\x03\xe4\x41\xca\x18\xa0\x23\x08\xd4\x73\xa4\xb0\xa8\x66\x4a\xff\x0b\x2b\x05\x07\xa6\x85\xea\xe8\x84\x63\x21\x07\x9c\x8d\x65\x79\x16\x3d\x56\xdf\xa5\x24\x52\x34\xd1\x46\xba\x7f\xe4\x6d\x06\xbc\x86\x5c\x8f\x25\x1b\xe9\x56\x46\xfa\x71\xbd\x57\x9d\xde\x46\xd6\x0e\x7d\x8b\xc5\x06\x51\x7e\xc0\x0b\x7c\x7b\xfc\x9e\xe7\xd8\x97\xe5\x89\x0c\x03\x1a\xba\xfa\x83\x39\xfa\x3b\x8a\xaa\xe5\xf2\x4a\x70\x79\x37\x8a\x1d\xe9\x89\x3d\xad\xbc\x97\x12\x6d\x16\xef\x19\xdd\xd4\x3a\x71\xf0\xe5\xf1\xa5\xe8\xa7\x6b\x39\x32\xd2\x96\xfa\xdb\x9d\x0f\x59\x3f\x25\xad\x15\xfc\x54\x29\xe9\x4e\x01\xcd\x73\x00\x00\x00\xff\xff\x5a\x3b\x0c\xba\x29\x02\x00\x00")

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

	info := bindataFileInfo{name: "deploy/istio-workspace/role_binding.yaml", size: 553, mode: os.FileMode(436), modTime: time.Unix(1563889278, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _deployIstioWorkspaceRole_binding_localYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x84\x90\x31\x6f\x1b\x31\x0c\x85\xf7\xfb\x15\x84\xd1\xd5\x57\x74\x2b\xb4\xb9\x85\xd1\xa9\x45\x61\x1b\xd9\x79\x12\xcf\xc7\xf8\x4e\x54\x28\xca\x06\x12\xe4\xbf\x07\x96\xcf\x71\x90\xc1\xd1\x26\x3e\xe2\xbd\xf7\xf1\xc0\x31\x38\xd8\xd1\x94\x46\x34\x6a\x30\xf1\x03\x69\x66\x89\x0e\x6c\x1e\xb6\x92\x28\xe6\x81\x7b\x6b\x59\xbe\x1f\x7f\x34\x09\x15\x27\x32\xd2\xec\x1a\x80\x25\x44\x9c\xc8\xc1\xbf\xd5\xdf\xf5\xf6\xff\xea\xf7\xba\x01\x00\x08\x94\xbd\x72\xb2\xea\xb4\xd8\xa1\xee\xc9\xea\x62\x4e\xe8\x09\x7a\x51\x38\x0d\xec\x07\x50\x19\x09\x3a\x8e\x81\xe3\x1e\xf2\x20\x65\x0c\xd0\x11\x04\xea\x39\x52\x58\x54\x33\xa5\xa7\xc2\x4a\xc1\x81\x69\xa1\x3a\x3a\xe2\x58\xc8\x01\x67\x63\x59\x9e\x44\x0f\xd5\x77\x29\x89\x14\x4d\xb4\x91\xee\x91\xbc\xcd\x05\x2f\x90\x1b\x19\xe9\xd7\x25\xa8\x5a\x7c\x64\xd5\x0e\x7d\x8b\xc5\x06\x51\x7e\xc6\x73\xeb\xf6\xf0\x33\xcf\xbc\xe7\xe5\x89\x0c\x03\x1a\xba\xfa\x83\x99\xf9\x53\x7c\xd5\x72\xb9\x45\x9f\xdf\x35\x7e\x4b\x7a\x64\x4f\x2b\xef\xa5\x44\x9b\xc5\x7b\x46\x57\xb5\x4e\x1c\x7c\x7b\x79\xbf\xf0\xeb\xe5\x2a\x32\xd2\x86\xfa\x6b\xce\x0d\xf2\xcb\x8a\x95\xfd\x8f\x4a\x49\x77\xc8\x9b\xb7\x00\x00\x00\xff\xff\x0b\xb5\x61\x6a\x1b\x02\x00\x00")

func deployIstioWorkspaceRole_binding_localYamlBytes() ([]byte, error) {
	return bindataRead(
		_deployIstioWorkspaceRole_binding_localYaml,
		"deploy/istio-workspace/role_binding_local.yaml",
	)
}

func deployIstioWorkspaceRole_binding_localYaml() (*asset, error) {
	bytes, err := deployIstioWorkspaceRole_binding_localYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "deploy/istio-workspace/role_binding_local.yaml", size: 539, mode: os.FileMode(436), modTime: time.Unix(1565641417, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _deployIstioWorkspaceCrdsIstio_v1alpha1_session_crdYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x4c\x8f\xb1\x4e\xc4\x30\x10\x44\x7b\x7f\xc5\x7e\x41\x50\x3a\xe4\x16\x3a\x10\x05\x48\xf4\x7b\xc9\x72\xac\xce\xb1\x57\x9e\x75\x84\x84\xf8\x77\xe4\xe4\xe0\x52\xfa\xcd\x78\x66\x87\x4d\xdf\xa5\x42\x4b\x8e\xc4\xa6\xf2\xe5\x92\xfb\x0b\xc3\xe5\x1e\x83\x96\xbb\x75\x3c\x89\xf3\x18\x2e\x9a\xe7\x48\x0f\x0d\x5e\x96\x57\x41\x69\x75\x92\x47\xf9\xd0\xac\xae\x25\x87\x45\x9c\x67\x76\x8e\x81\x28\xf3\x22\x91\x20\xd8\x83\x16\x56\x78\xe5\x41\x4b\x80\xc9\xd4\x1d\xe7\x5a\x9a\x45\x3a\x28\xfb\x2f\x74\x91\x68\xef\x7a\xdb\x03\x36\x92\x14\xfe\x74\xa4\xcf\x0a\xdf\x14\x4b\xad\x72\xba\xd5\x6d\x10\x9a\xcf\x2d\x71\xfd\xc7\x81\x08\x53\x31\x89\xf4\xd2\x6b\x8c\x27\x99\x03\xd1\xfa\x37\x7d\x1d\x39\xd9\x27\x8f\xdd\xd7\x4e\xf5\xba\xef\x7a\x0e\x9c\xbd\x21\xd2\xf7\x4f\xf8\x0d\x00\x00\xff\xff\x22\x41\x9e\x9e\x2f\x01\x00\x00")

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

	info := bindataFileInfo{name: "deploy/istio-workspace/crds/istio_v1alpha1_session_crd.yaml", size: 303, mode: os.FileMode(436), modTime: time.Unix(1565641417, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _deployIstioWorkspaceRoleYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x9c\x53\x4f\x6b\xeb\x30\x0c\xbf\xe7\x53\x98\x5e\x0a\x0f\x9a\xc7\xbb\x3d\x72\xdd\x61\xf7\x31\x76\x57\x1d\xb5\x15\xb5\x2d\x23\xc9\x19\xdb\xa7\x1f\x4e\xc3\x9a\x42\xd3\x91\xe5\x14\x2b\xd2\xef\x5f\x64\xc8\xf4\x86\xa2\xc4\xa9\x73\xb2\x07\xdf\x42\xb1\x13\x0b\x7d\x82\x11\xa7\xf6\xfc\x5f\x5b\xe2\xbf\xc3\xbf\xe6\x4c\xa9\xef\xdc\x53\x28\x6a\x28\x2f\x1c\xb0\x89\x68\xd0\x83\x41\xd7\x38\xe7\x05\xc7\x81\x57\x8a\xa8\x06\x31\x77\x2e\x95\x10\x1a\xe7\x12\x44\xec\x1c\xa9\x11\xef\xde\x59\xce\x9a\xc1\x63\x23\x25\xa0\xd6\xc1\x9d\x83\x4c\xcf\xc2\x25\x8f\xc7\xfa\xec\xdc\x66\x33\xbe\x0a\x2a\x17\xf1\x38\xfb\x92\xb9\xd7\xef\x83\xa2\x0c\xe4\xf1\x5a\xc0\xd4\x67\xa6\x64\xd7\x4a\xae\xde\xd4\x30\xd9\xc0\xa1\x44\xf4\x01\x28\xce\x06\x06\x9c\x77\x7b\x4e\x07\x3a\x46\xc8\x73\x0e\x2f\x38\xb5\x0c\x28\xfb\x99\x96\xed\x9f\xed\x7a\x03\x35\x8e\x31\x82\xbb\x90\x47\xb4\x25\x48\xc8\x93\xaa\x3b\xa0\x3d\xe6\xc0\x1f\xf1\xc6\x4b\x0f\x18\x39\x29\xce\x4a\x82\x39\x90\x87\x9b\x9a\x1a\x18\x1e\x4a\xd0\xf5\x26\xab\xa2\x96\x33\x26\x3d\xd1\xc1\x5a\xe2\x9f\xe5\x5d\x02\x5e\x4b\x14\x39\x91\xb1\x50\x3a\xb6\x9e\x05\x59\x5b\xcf\x71\x89\x6c\x5a\x8a\x69\xe6\x41\xca\xd3\x2f\xaf\x8b\x8b\x4b\xcc\xe3\xda\xce\x3c\x3e\xe0\xbd\xe8\x5f\x63\x2b\xa1\xd5\x0b\x51\x6d\x5d\x78\x96\x13\x5c\x0f\x1e\x81\xd4\x04\x7e\x87\xf9\x15\x00\x00\xff\xff\xb9\xb9\xb4\xf9\x15\x04\x00\x00")

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

	info := bindataFileInfo{name: "deploy/istio-workspace/role.yaml", size: 1045, mode: os.FileMode(436), modTime: time.Unix(1565641417, 0)}
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
	"deploy/istio-workspace/operator.yaml":                        deployIstioWorkspaceOperatorYaml,
	"deploy/istio-workspace/role_local.yaml":                      deployIstioWorkspaceRole_localYaml,
	"deploy/istio-workspace/service_account.yaml":                 deployIstioWorkspaceService_accountYaml,
	"deploy/istio-workspace/role_binding.yaml":                    deployIstioWorkspaceRole_bindingYaml,
	"deploy/istio-workspace/role_binding_local.yaml":              deployIstioWorkspaceRole_binding_localYaml,
	"deploy/istio-workspace/crds/istio_v1alpha1_session_crd.yaml": deployIstioWorkspaceCrdsIstio_v1alpha1_session_crdYaml,
	"deploy/istio-workspace/role.yaml":                            deployIstioWorkspaceRoleYaml,
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
				"istio_v1alpha1_session_crd.yaml": &bintree{deployIstioWorkspaceCrdsIstio_v1alpha1_session_crdYaml, map[string]*bintree{}},
			}},
			"operator.yaml":           &bintree{deployIstioWorkspaceOperatorYaml, map[string]*bintree{}},
			"role.yaml":               &bintree{deployIstioWorkspaceRoleYaml, map[string]*bintree{}},
			"role_binding.yaml":       &bintree{deployIstioWorkspaceRole_bindingYaml, map[string]*bintree{}},
			"role_binding_local.yaml": &bintree{deployIstioWorkspaceRole_binding_localYaml, map[string]*bintree{}},
			"role_local.yaml":         &bintree{deployIstioWorkspaceRole_localYaml, map[string]*bintree{}},
			"service_account.yaml":    &bintree{deployIstioWorkspaceService_accountYaml, map[string]*bintree{}},
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
