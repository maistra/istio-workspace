// Code generated for package assets by go-bindata DO NOT EDIT. (@generated)
// sources:
// deploy/cluster_role.yaml
// deploy/service_account.yaml
// deploy/operator.yaml
// deploy/operator.tpl.yaml
// deploy/role.yaml
// deploy/crds/maistra.io_sessions_crd.yaml
// deploy/crds/maistra.io_sessions_cr.yaml
// deploy/olm-catalog/istio-workspace/manifests/istio-workspace.clusterserviceversion.yaml
// deploy/olm-catalog/istio-workspace/manifests/maistra.io_sessions_crd.yaml
// deploy/role_binding.yaml
// deploy/cluster_role_binding.yaml
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

var _deployCluster_roleYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xac\x53\xb1\x8e\xd4\x40\x0c\xed\xf3\x15\xa3\x6b\x4e\x42\x4a\x4e\x74\x28\x2d\x05\x1d\x05\x42\xf4\xbe\x89\xb3\x67\xed\xcc\x78\x64\x7b\x82\xb8\xaf\x47\x93\xcd\x42\x16\x92\x45\x41\xb7\xd5\x8e\x1d\xbf\xe7\x67\x3f\x43\xa6\x6f\x28\x4a\x9c\x7a\x27\xcf\xe0\x3b\x28\xf6\xc2\x42\xaf\x60\xc4\xa9\x3b\x7f\xd0\x8e\xf8\x69\x7a\xdf\x9c\x29\x0d\xbd\xfb\x18\x8a\x1a\xca\x17\x0e\xd8\x44\x34\x18\xc0\xa0\x6f\x9c\xf3\x82\x73\xc1\x57\x8a\xa8\x06\x31\xf7\x2e\x95\x10\x1a\xe7\x12\x44\xec\x1d\xa9\x11\xb7\xdf\x59\xce\x9a\xc1\x63\x23\x25\xa0\xd6\xc2\xd6\x41\xa6\x4f\xc2\x25\xcf\xcf\xfa\x6b\xdd\xc3\xc3\xfc\x57\x50\xb9\x88\xc7\x55\xa6\xa2\xcd\x08\x3a\x87\x26\x94\xe7\x55\xf6\x84\x76\x1c\x32\xf3\xa0\xbf\x1e\x8a\x32\xd1\x15\xbd\x06\x30\x0d\x99\x29\xd9\xef\x48\xae\xe3\x52\xc3\x64\x13\x87\x12\xd1\x07\xa0\xb8\x2a\x98\x70\xfd\xb5\xe7\x34\xd2\x29\x42\x5e\x73\x78\x41\xdb\x14\xf0\xf8\xee\x71\x4f\x00\xe4\x05\x62\x43\xc2\x80\x39\xf0\x8f\x78\x43\x3c\x00\x46\x4e\x8a\xab\x90\x60\x0e\xe4\xe1\x26\xa6\x06\x86\x63\x09\xfa\x7f\x1d\x75\x9c\x31\xe9\x0b\x8d\xd6\x11\xff\xbb\xbd\xcb\x34\x8e\x12\x45\x4e\x64\x2c\x94\x4e\x9d\x67\x41\xd6\xce\x73\xdc\x23\x5b\x36\xb8\xd4\xdc\xb1\xc9\xb2\x9f\x6a\x5c\xdc\x63\x9e\x6d\xbb\xd2\x78\x87\xf7\xd2\xff\x11\x59\x09\xad\x1e\x44\x95\x75\xe1\xd9\x9f\xe0\x71\xf0\x08\xa4\x26\xf0\x66\x98\x1b\x06\xfc\x5c\x6f\xf1\x9a\xfd\xf3\xc0\x37\x48\x6f\x7c\xfa\x34\x52\x82\x40\xaf\xf8\xf7\x8a\x5a\x57\xf2\x50\x97\xf2\x33\x00\x00\xff\xff\x5b\x05\xbf\xda\x9c\x04\x00\x00")

func deployCluster_roleYamlBytes() ([]byte, error) {
	return bindataRead(
		_deployCluster_roleYaml,
		"deploy/cluster_role.yaml",
	)
}

func deployCluster_roleYaml() (*asset, error) {
	bytes, err := deployCluster_roleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "deploy/cluster_role.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _deployService_accountYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x04\xc0\x31\x0e\x80\x30\x08\x00\xc0\x9d\x57\xf0\x01\x07\x57\x36\xdf\x60\xe2\x4e\x28\x03\x69\x0a\x4d\xc1\xfa\x7d\x8f\xa7\x3d\xba\xd2\xc2\x09\xf7\x09\xdd\xbc\x11\xde\xba\xb6\x89\x5e\x22\xf1\x7a\xc1\xd0\xe2\xc6\xc5\x04\x88\xce\x43\x09\x2d\xcb\xe2\xf8\x62\xf5\x9c\x2c\x0a\x7f\x00\x00\x00\xff\xff\x94\xa0\xb7\x3f\x46\x00\x00\x00")

func deployService_accountYamlBytes() ([]byte, error) {
	return bindataRead(
		_deployService_accountYaml,
		"deploy/service_account.yaml",
	)
}

func deployService_accountYaml() (*asset, error) {
	bytes, err := deployService_accountYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "deploy/service_account.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _deployOperatorYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x7c\x91\xcb\x6e\xe2\x4c\x10\x85\xf7\x7e\x8a\x52\x58\xdb\x5c\xf2\xff\xb3\xe8\x9d\xc5\x45\x41\x0a\xd8\x02\x6b\xa2\x59\x8d\x6a\xda\x85\x69\xd1\xb7\xe9\x6e\x3b\xf2\xdb\x8f\x4c\x08\x31\xc1\x49\xad\xa0\xce\xa9\xf3\xb9\xba\x46\x23\x48\xeb\x60\x2a\xd2\xe4\x30\x50\x09\x07\x67\x14\x18\xdb\xfd\x33\x2e\x09\x56\x26\x2d\x2a\x09\x18\xe0\x85\x4a\x98\xfc\x0f\x69\x5d\xc1\x6c\x32\x9b\xc0\x74\xc6\xa6\x3f\xd8\x7f\x8f\x90\x6f\x60\xbe\xdc\x17\xd1\x68\x04\x8b\x0c\xb6\x59\x01\x9b\x6c\xb1\x5e\xfd\x82\xe2\x69\xbd\x87\xd5\xfa\x79\x99\x40\x2e\x09\x3d\x01\x3f\xa2\xae\x68\x20\x5f\x68\x1f\x08\xcb\x24\x8a\xd0\x8a\x9f\xe4\xbc\x30\x9a\x01\x5a\xeb\xc7\xcd\x34\x3a\x09\x5d\x32\x58\x90\x95\xa6\x55\xa4\x43\xa4\x28\x60\x89\x01\x59\x04\x20\xf1\x0f\x49\xdf\xfd\x82\x6e\x80\x81\xf0\x41\x98\xf8\xd5\xb8\x93\xb7\xc8\xe9\x2c\x34\xef\x91\xcd\x24\x99\x24\x8f\x11\x80\x46\x45\xf7\x5e\x6f\x89\x77\x51\x8e\xac\x14\x1c\x3d\x83\x69\x04\xe0\x49\x12\x0f\xc6\xbd\x41\x14\x06\x7e\x7c\xee\x51\xbf\xe1\x0e\x91\x03\x29\x2b\x31\xd0\x25\xad\xb7\x4a\x57\xf2\x26\xf8\xdb\xe8\xa1\xf0\xae\x86\x57\xeb\x94\xf7\xf5\xba\xe2\x46\x07\x14\x9a\xdc\x15\x16\x03\xba\xaa\x87\x8e\xc1\x93\x6b\x3e\x68\xdc\x28\x85\xba\xec\x1b\xc4\xe9\x43\x26\xdd\xf4\xa5\xb7\xaf\x78\x49\x8b\xf9\xd3\xef\x6d\xba\x59\xee\xf3\x74\xbe\xbc\xea\x00\x0d\xca\x9a\x18\x3c\x3c\xdc\xcd\xe4\xd9\xe2\x3c\xf1\xd9\xbc\x72\x46\xb1\x5e\x13\xe0\x20\x48\x96\x3b\x3a\xdc\x76\x2f\xfd\x1c\xc3\x91\x5d\xdf\x37\xe9\xb2\xef\x50\x59\xbe\xdc\xa5\x45\xb6\x1b\xe4\x7d\xfd\xee\x42\x61\x45\x0c\xfe\xd6\xd8\x26\xc2\x8c\x15\x0a\x1f\x1c\x8e\x3f\xd9\xd9\xcd\x55\x2e\x53\x79\x2d\x65\x6e\xa4\xe0\x2d\x83\x54\xbe\x62\xeb\xaf\xfa\xd7\x77\x83\xf3\x25\x04\xa7\x94\x73\x53\xeb\xb0\x1d\x74\xc6\x71\x1c\xfd\x0b\x00\x00\xff\xff\x00\xae\x8f\x15\xd1\x03\x00\x00")

func deployOperatorYamlBytes() ([]byte, error) {
	return bindataRead(
		_deployOperatorYaml,
		"deploy/operator.yaml",
	)
}

func deployOperatorYaml() (*asset, error) {
	bytes, err := deployOperatorYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "deploy/operator.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _deployOperatorTplYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x8c\x54\x5d\x6f\xf2\x3a\x0c\xbe\xe7\x57\x58\x68\xb7\x63\xda\x6d\xee\xaa\xc1\xe1\xa0\x7d\x80\x4a\xb5\xa3\x73\x35\x99\xd4\xd0\xbc\xa4\x49\x96\xa4\xa0\x6a\xe2\xbf\xbf\x4a\xd7\x0e\xe8\xda\x75\xbe\xab\x63\xf9\xf9\xa8\xed\xbd\x50\x29\x83\x84\x72\x23\xd1\xd3\x08\x8d\x78\x25\xeb\x84\x56\x0c\x7c\x9d\x9c\x68\x43\xca\x65\x62\xeb\x27\x42\xdf\x1d\xee\x47\x06\x2d\xe6\xe4\xc9\x3a\x36\x02\xb8\x05\x85\x39\x31\x58\x3c\xce\xde\xa6\xcb\x87\xc7\x59\xfc\x16\xcf\xe6\x8b\x75\x12\xff\x3f\x02\x00\x48\xc9\x71\x2b\x8c\xaf\x7a\x8e\xa7\x9a\xef\xc9\x82\xa5\x9d\x70\xde\x96\x70\xcc\xc8\x12\xa4\x64\xa4\x2e\x29\x05\x91\xe3\x8e\x40\x38\xc0\x03\x0a\x89\x1b\x49\xe3\xaa\x89\xa5\xf7\x42\x58\x4a\x19\x78\x5b\x50\x95\x3a\xa0\x2c\x88\xc1\x7b\x81\xe5\x44\xe8\x3e\x22\xab\xe5\x7a\x91\x2c\x3b\xa9\xc4\x64\xb4\x13\x5e\xdb\x12\x84\x82\x63\x26\x78\x06\x3e\xa3\x9a\x03\x47\x05\x1b\x82\xad\x2e\x54\x3a\xc4\x21\xc7\x20\x06\x5b\x1c\x16\xcf\xd1\x7c\xf6\xf6\x12\x3d\xcf\x3a\xc0\x93\x8c\xaa\x52\xd0\xdb\x0b\xd0\xa3\xf0\x19\x88\x3d\xc1\x46\x28\xb4\xe5\x10\xae\x70\x5e\xe8\xdb\xa3\xb6\x7b\x67\x90\x53\x27\x7e\x12\xcd\x7b\xe0\x3d\xee\xae\xd1\xbd\x0e\x8a\x0b\x47\x83\x82\xc3\x58\x38\xdf\xc2\x7b\x9d\xc5\xeb\xc5\xf2\xa5\x07\xed\xf0\x39\x56\x0d\xe2\xef\x04\x7e\xc3\xf9\x2f\x4a\x1e\xfe\xad\x3c\x5d\xaf\xa2\x87\x9f\x8c\xad\x1c\x09\x92\x8e\xe8\x79\x36\x81\x27\xc2\x03\x01\xe5\xc6\x97\xb0\xd5\x16\xb8\x2c\x9c\x27\x0b\x47\x91\x52\x5d\xd3\x66\xb3\x45\xe9\xae\xe8\x8c\xc7\x23\xbd\xf9\x43\xdc\xd7\x83\xff\xb9\x3c\xd3\x6a\x78\x73\x52\xbe\xaa\xbd\x5c\x21\x34\xc6\x85\x8d\x09\xf9\x9c\x3c\xa6\xe8\x91\x55\x5f\x50\x0b\xfa\xfe\x0b\x43\x48\xdc\x90\x74\x4d\x65\xe8\x69\xfa\x4a\xa1\x71\x96\xc1\xcd\xc7\xc5\x6f\x38\x55\xef\xce\x10\x6f\xda\x58\x32\x52\x70\x74\x0c\xee\xeb\x8c\x23\x49\xdc\x6b\x7b\x06\xca\x83\x11\x4f\x2d\xf4\x01\xfc\x9f\x19\xc0\xd7\x1d\xb9\x40\x69\x59\x31\x64\x47\x97\x25\xbf\xa0\x35\x44\xec\xda\x9e\x4f\x43\xec\x41\x70\x8a\x38\xd7\x85\xf2\x2f\x03\x94\xb8\x56\x1e\x85\xaa\xef\xe0\x39\x6e\x07\xc5\x84\xa8\x76\xae\x61\xd6\x3a\x9c\xa7\xbb\x56\xba\x39\x63\xcd\xc3\xf9\xb6\x9c\xd8\x65\x26\x89\xe6\xa7\x16\x0e\xd7\x79\x8e\x2a\x65\xad\x74\xa0\x29\xf6\x6d\x52\x68\x77\xae\xab\x32\x18\xd3\x29\x60\x55\x48\xb9\xd2\x52\xf0\x92\x41\x24\x8f\x58\xba\x56\x15\xa9\x43\x57\xc3\xfe\x75\xbe\x8e\x66\xf3\x6e\x3e\x5a\xb5\xa7\x71\x6f\xd7\xd5\x72\x7a\x3e\xbb\x1d\xed\xfe\xb1\x3a\xff\xce\x29\xc4\x56\x90\x4c\x63\xda\x76\xbf\xd6\xef\x2b\xf4\x19\xfb\x9a\xe1\x49\xc0\xec\xa5\xb2\x5c\xcd\xe2\x28\x59\xc6\x3f\xf2\x61\x30\x6e\x8d\xca\x78\xf4\x37\x00\x00\xff\xff\xd1\xef\x0d\x97\x98\x07\x00\x00")

func deployOperatorTplYamlBytes() ([]byte, error) {
	return bindataRead(
		_deployOperatorTplYaml,
		"deploy/operator.tpl.yaml",
	)
}

func deployOperatorTplYaml() (*asset, error) {
	bytes, err := deployOperatorTplYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "deploy/operator.tpl.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _deployRoleYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xac\x53\xc1\x8e\xd4\x30\x0c\xbd\xf7\x2b\xa2\xbd\xac\x84\xd4\xae\xb8\xa1\xfe\x00\x37\x0e\x08\x71\xf7\xa6\xee\xac\x35\x49\x1c\xd9\x4e\x11\xf3\xf5\x28\x9d\xc2\x74\xa0\x1d\x54\x69\x7b\x6a\x1c\xdb\xcf\xef\xf9\x05\x32\x7d\x47\x51\xe2\xd4\x3b\x79\x05\xdf\x41\xb1\x37\x16\xba\x80\x11\xa7\xee\xfc\x49\x3b\xe2\x97\xe9\x63\x73\xa6\x34\xf4\xee\x2b\x07\x6c\x22\x1a\x0c\x60\xd0\x37\xce\x79\xc1\x39\xf3\x1b\x45\x54\x83\x98\x7b\x97\x4a\x08\x8d\x73\x09\x22\xf6\x8e\xd4\x88\xdb\x1f\x2c\x67\xcd\xe0\xb1\x91\x12\x50\x6b\x61\xeb\x20\xd3\x67\xe1\x92\xe7\x63\xfd\x5a\xf7\xf4\x34\xff\x0a\x2a\x17\xf1\xb8\xba\xc9\x3c\xe8\x9f\x83\xa2\x4c\xe4\xf1\x16\xc0\x34\x64\xa6\x64\xb7\x48\xae\xa4\xd4\x30\xd9\xc4\xa1\x44\xf4\x01\x28\xae\x0a\x26\x5c\x67\x7b\x4e\x23\x9d\x22\xe4\x35\x86\x17\x5c\x52\x26\x94\xd7\xd5\x2c\xcf\x1f\x9e\x8f\x13\xa8\x72\xcc\x12\x6c\xb6\x3c\xa1\xed\xb5\x84\xbc\x4c\xb5\xd1\x74\xc0\x1c\xf8\x67\xbc\xe3\x32\x00\x46\x4e\x8a\xab\x90\x60\x0e\xe4\xe1\x2e\xa6\x06\x86\x63\x09\x7a\x9c\x64\x9d\xa8\xe3\x8c\x49\xdf\x68\xb4\x8e\xf8\xff\xe3\x5d\x05\x3e\x0a\x14\x39\x91\xb1\x50\x3a\x75\x9e\x05\x59\x3b\xcf\x71\x0f\x6c\x31\xc5\x52\xf3\x40\xe5\x65\xe5\xd5\xb8\xb8\x87\x3c\xdb\x76\xc5\xf1\x01\xee\x75\xfe\x23\xb4\x12\x5a\x7d\x10\x95\xd6\x15\x67\x5f\xc1\xe3\xcd\x23\x90\x9a\xc0\xbb\xf5\xdc\x30\xe0\x97\x6a\xe5\xdf\xb7\x7f\x3f\xf0\x0d\xd0\x3b\x9f\xbe\x8c\x94\x20\xd0\x05\xff\x5d\x51\xeb\x4a\x1e\xea\x52\x7e\x05\x00\x00\xff\xff\x2e\xd1\xd4\xd7\x95\x04\x00\x00")

func deployRoleYamlBytes() ([]byte, error) {
	return bindataRead(
		_deployRoleYaml,
		"deploy/role.yaml",
	)
}

func deployRoleYaml() (*asset, error) {
	bytes, err := deployRoleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "deploy/role.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _deployCrdsMaistraIo_sessions_crdYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xe4\x58\xc1\x8e\xe3\x36\x0f\xbe\xe7\x29\x08\xfc\x87\xbd\xfc\x49\x30\xe8\xa5\xf0\x6d\x31\xed\x61\xd0\x6d\x31\x98\x14\x7b\x67\x2c\xda\x61\xd7\x96\x54\x89\xca\x36\x2d\xfa\xee\x85\x24\x3b\x71\x1c\x3b\x99\xcc\x62\x81\x62\xcb\x5b\x28\x8a\xa4\xbe\x8f\xa4\x14\xa3\xe5\x8f\xe4\x3c\x1b\x5d\x00\x5a\xa6\x3f\x84\x74\xfc\xe5\x57\x9f\xbe\xf7\x2b\x36\xeb\xfd\xc3\x96\x04\x1f\x16\x9f\x58\xab\x02\x1e\x83\x17\xd3\xbe\x90\x37\xc1\x95\xf4\x03\x55\xac\x59\xd8\xe8\x45\x4b\x82\x0a\x05\x8b\x05\x80\xc6\x96\x0a\xf0\xe4\xb3\xa3\x16\xd9\x8b\xc3\x15\x9b\x85\xb7\x54\x46\x8b\xda\x99\x60\x0b\x18\xac\xe4\x5d\x3e\x2e\x02\xe4\x58\x9b\xec\x20\x69\x1a\xf6\xf2\xd3\x50\xfb\x81\xbd\xa4\x15\xdb\x04\x87\xcd\x29\x5c\x52\x7a\xd6\x75\x68\xd0\x1d\xd5\x0b\x00\x5f\x1a\x4b\x05\xfc\x12\xc3\x58\x2c\x49\x45\x5d\xd8\xba\xee\x2c\x5d\x68\x2f\x28\xc1\x17\xf0\xd7\xdf\x0b\x80\x3d\x36\xac\x30\x9e\x2f\x2f\x1a\x4b\xfa\xfd\xf3\xd3\xc7\xef\x36\xe5\x8e\x5a\xcc\x4a\x00\x45\xbe\x74\x6c\x93\x5d\x9f\x1f\xb0\x07\xd9\x11\x64\x4b\xa8\x8c\x4b\x3f\xfb\x2c\xe1\xfd\xf3\xd3\xaa\xdb\x6e\x9d\xb1\xe4\x84\xfb\x14\xa2\x0c\x68\x39\xea\x46\x81\xde\xc5\x4c\xb2\x0d\xa8\x48\x04\xe5\x88\xfb\xac\x23\x05\x3e\xc7\x36\x15\xc8\x8e\x3d\x38\xb2\x8e\x3c\x69\x49\x27\x1a\xb8\x85\x68\x82\x1a\xcc\xf6\x37\x2a\x65\x05\x1b\x72\xd1\x09\xf8\x9d\x09\x8d\x82\xd2\xe8\x3d\x39\x01\x47\xa5\xa9\x35\xff\x79\xf4\xec\x41\x4c\x0a\xd9\xa0\x50\x47\x47\x2f\xac\x85\x9c\xc6\x26\x62\x18\xe8\xff\x80\x5a\x41\x8b\x07\x70\x14\x63\x40\xd0\x03\x6f\xc9\xc4\xaf\xe0\x67\xe3\x08\x58\x57\xa6\x80\x9d\x88\xf5\xc5\x7a\x5d\xb3\xf4\x85\x58\x9a\xb6\x0d\x9a\xe5\xb0\x2e\x8d\x16\xc7\xdb\x20\xc6\xf9\xb5\xa2\x3d\x35\x6b\xcf\xf5\x12\x5d\xb9\x63\xa1\x52\x82\xa3\x35\x5a\x5e\xa6\xc4\xb5\xe4\x22\x54\xff\x3b\x32\xfd\x6e\x90\xa9\x1c\x62\x51\x78\x71\xac\xeb\xa3\x3a\xd5\xdf\x2c\xee\xb1\x0e\x23\xbd\xd8\x6d\xcb\xf9\x9f\xe0\x8d\xaa\x88\xca\xcb\x8f\x9b\x5f\xa1\x0f\x9a\x28\x38\xc7\x3c\xa1\x7d\xda\xe6\x4f\xc0\x47\xa0\x58\x57\xe4\x32\x71\x95\x33\x6d\xf2\x48\x5a\x59\xc3\x5a\xd2\x8f\xb2\x61\xd2\xe7\xa0\xfb\xb0\x6d\x59\x22\xd3\xbf\x07\xf2\x12\xf9\x59\xc1\x23\x6a\x6d\x04\xb6\x04\xc1\x2a\x14\x52\x2b\x78\xd2\xf0\x88\x2d\x35\x8f\xe8\xe9\xab\xc3\x1e\x11\xf6\xcb\x08\xe9\x6d\xe0\x87\x53\xe4\xdc\x30\xa3\x75\x54\xf7\x83\x64\x92\xa1\xae\x05\x37\x96\xca\xb3\xce\x50\xe4\xd9\xc5\xea\x15\x14\x8a\x35\xdf\x19\xae\x06\x8e\xa6\x9a\x31\x8a\xa3\xea\x5c\x01\xc0\x42\xad\x1f\x2b\x47\xa9\xbc\x50\x75\x25\x85\x38\x15\x30\x0d\xab\x26\xd6\x4f\x45\x8e\x74\x49\x17\x1e\x01\x3e\xb3\xec\x58\xe7\x81\x72\x99\xf3\xf5\xcc\xb3\xa0\xab\x27\xf5\x00\xa8\x54\x1a\xe1\xd8\x3c\x5f\xf5\x00\x73\xa4\x5d\x1a\x8c\xc8\x3a\x49\xba\x19\xae\x6c\x9c\xf5\x1c\xaf\x09\xa1\xfa\xf0\x86\xcd\x57\x52\xca\x4b\xe8\x1c\x1e\xce\xc9\x36\x41\x2e\xf2\x3c\xe7\x35\x5a\x9c\x31\xdb\x67\x98\x48\xdd\x99\xcf\x49\x29\x0e\xab\x8a\xcb\x38\x2d\x92\x4f\x75\x99\x5d\x9e\xa0\x2f\x54\x8d\x19\xbd\xc6\xe7\x1c\x8e\xb7\x81\xb8\x77\x53\x9a\x6d\x77\xee\x9a\x41\x7c\xba\x91\xf3\x7d\x7b\xab\x95\x93\xd5\x19\xde\x66\xeb\xe3\xb4\x7c\x73\x37\x5f\xa0\xfa\xca\x76\x7e\x55\x2a\x51\xcb\x5a\xf1\x9e\x55\xc0\x66\x02\xbd\x09\xbe\xbf\xf1\x0e\x1e\xbd\xb3\xc6\x32\x83\x7e\x96\x31\x07\xfd\xfb\x73\x9a\x85\x63\x24\x68\x43\x24\x44\xad\x4b\x47\x38\xd5\x7c\xbd\xa0\x07\x8b\x4e\x7a\xe6\x26\xc9\xc9\x72\x9d\xa2\xce\x5b\x29\xa3\x77\xdb\x58\x6e\x92\x00\x13\x8f\x90\x37\x39\x99\xa7\xeb\x0e\x27\xf1\xd0\xd7\x9c\xdc\x53\x7f\x77\x84\xbd\x59\x8b\xaf\x32\x99\x9b\xf1\x59\xbe\xe0\x5e\x01\x10\x74\x35\xc9\x17\x17\xf4\x07\xdc\x52\x43\x6a\x58\xd7\xe9\x75\x39\x54\xc4\x07\x40\x36\xf4\xdf\x4a\x69\x36\xe9\x34\xff\xd6\xba\xfa\x0f\xb4\x4e\xfc\x97\x10\x1f\xa3\x73\xf1\x96\x1d\x45\x5f\xa3\xf3\xde\xf0\x28\x4b\x97\xeb\x38\xd7\x19\x34\x26\xdc\x8f\x54\xfb\xfe\xa3\xc7\xfe\x01\x1b\xbb\xc3\x87\x93\x2e\x11\xb0\xec\xbe\x62\x0c\x96\x01\xf2\x05\x53\x80\xb8\x40\xdd\xd7\x02\xe3\xb0\xa6\x4e\xf3\x4f\x00\x00\x00\xff\xff\xb8\x1e\x54\xf0\x4c\x11\x00\x00")

func deployCrdsMaistraIo_sessions_crdYamlBytes() ([]byte, error) {
	return bindataRead(
		_deployCrdsMaistraIo_sessions_crdYaml,
		"deploy/crds/maistra.io_sessions_crd.yaml",
	)
}

func deployCrdsMaistraIo_sessions_crdYaml() (*asset, error) {
	bytes, err := deployCrdsMaistraIo_sessions_crdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "deploy/crds/maistra.io_sessions_crd.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _deployCrdsMaistraIo_sessions_crYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x2c\x8b\xc1\x0d\xc2\x30\x10\x04\xff\xae\xe2\x1a\x30\x28\x5f\xb7\x81\xc4\x7f\x85\x17\x62\x48\xce\x96\xef\x12\x89\xee\x91\x49\x7e\xab\x9d\x19\xb4\x72\x67\xb7\x52\x35\xc9\x8a\x62\xde\x71\x29\xf5\xba\x4f\x58\xda\x8c\x29\x7c\x8a\xe6\x24\x37\xda\x50\xc2\x4a\x47\x86\x23\x05\x11\xc5\xca\x24\x4e\xf3\x68\x27\xb6\xc6\xc7\x40\xbd\x6e\xce\x31\x44\xfc\xdb\x98\x64\x26\x32\xfb\xff\x38\x32\x6a\x8e\x9b\x9d\xd7\x8e\x65\x63\x92\x37\xac\xea\xa8\xf9\x3c\xda\x78\xca\x1d\x5e\xf4\x65\x71\x9f\xc2\x2f\x00\x00\xff\xff\x83\x49\x12\xb0\xaf\x00\x00\x00")

func deployCrdsMaistraIo_sessions_crYamlBytes() ([]byte, error) {
	return bindataRead(
		_deployCrdsMaistraIo_sessions_crYaml,
		"deploy/crds/maistra.io_sessions_cr.yaml",
	)
}

func deployCrdsMaistraIo_sessions_crYaml() (*asset, error) {
	bytes, err := deployCrdsMaistraIo_sessions_crYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "deploy/crds/maistra.io_sessions_cr.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _deployOlmCatalogIstioWorkspaceManifestsIstioWorkspaceClusterserviceversionYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xec\x58\x51\x6f\xdb\x36\x10\x7e\xf7\xaf\x38\xf8\x25\xc0\x10\xb9\x0d\x36\xec\x41\x4f\x73\xbb\x6e\x0d\xb6\x24\x46\x52\xac\x0f\x45\x31\x5c\xc8\xb3\xcd\x9a\x22\x39\xf2\xe4\xcc\xed\xfa\xdf\x07\x8a\xb4\x2d\xc9\x76\x9d\xac\x59\x1f\x8a\xea\xc5\xf2\xf1\x8e\xf7\x1d\xf9\xf1\xf4\x49\xe8\xd4\x1f\xe4\x83\xb2\xa6\x04\xeb\xc8\x23\x5b\x1f\x46\xc2\x7a\xb2\xf1\xa7\x7a\xb2\x3c\x43\xed\xe6\x78\x36\x58\x28\x23\x4b\x78\xae\xeb\xc0\xe4\x6f\xc8\x2f\x95\xa0\x1c\x3a\xa8\x88\x51\x22\x63\x39\x00\x40\x63\x2c\x23\x2b\x6b\x42\xfc\x0b\x80\xba\x2a\xe8\x6f\xac\x9c\xa6\x50\xc2\x3f\x45\x63\x04\x78\x93\x7f\x01\x3e\x6c\xee\x00\x86\x5b\x40\xc3\x12\x86\x15\xaa\xc0\x1e\x47\xca\x6e\x80\x0c\x4f\xdb\xee\x11\x55\x74\xbc\xa1\xd0\xc4\x74\x06\xd7\xb0\x86\x65\x27\x07\xc0\xd0\x60\x45\x31\x8c\x29\x70\x11\x72\x6c\xcb\xe5\x63\x67\x9e\xe0\x48\xec\xce\xe1\x69\x3a\x2c\x5b\x65\xec\x16\xd3\x4b\xe6\x91\x95\x99\x85\x62\x79\x36\xec\x39\x7d\xec\xfc\x7f\x7b\xda\x4b\x64\x6b\xa6\x9d\xfc\xad\x89\xc9\xc8\xa2\x0e\xe4\x87\xa7\x7d\x0f\x5e\xb9\xc6\x63\x4e\x28\xf7\x8d\x2f\x51\xd7\x8d\xc3\x3b\x0c\xdd\x15\xe8\x82\xda\xde\xaf\xef\xde\x36\xbf\x02\x1d\xde\x2a\xad\x58\xc5\xbd\x7d\x86\x41\x09\x38\x37\x81\x51\xeb\x01\x40\xc4\x57\x82\x0a\xac\x6c\x71\x67\xfd\x22\x38\x14\x34\x5a\x3e\x1d\x3d\x1d\x7d\x9f\x87\x1b\x53\x09\x4e\xa3\xa0\xb9\xd5\x92\xfc\x20\xae\x76\x43\x24\xa7\x42\xa2\x99\xa4\xa9\x32\x2a\x51\x0a\x3e\x44\x04\xa2\x0e\x6c\x2b\x4f\xc1\xd6\xbe\x3b\xde\xc0\xb2\x77\x86\x64\xba\x2d\x40\x52\x10\x5e\x39\x6e\x38\x9e\x79\x02\x2a\x00\xcf\x09\x6e\xc4\x9c\x2a\x84\xa9\xf5\xcd\xdf\xcc\x84\x00\xe3\xc9\xf9\x28\x17\x9a\x78\x9f\xe3\xb2\x2d\x15\xb6\xf6\x1e\x6d\x59\x9a\xc7\x97\xeb\x23\xb5\x39\x3d\xd0\x83\x81\x53\xd2\x2b\x90\xb4\x24\x6d\x1d\xa0\x91\x10\xa9\x08\xd6\x00\x9a\x15\x2c\xea\x5b\xf2\x86\x98\x02\x88\x74\xe0\xe0\x4e\xf1\xdc\xd6\x0c\xb2\x49\x25\x22\x93\x52\xa9\x3c\x27\x1f\x22\x58\xa9\x82\xd3\xb8\xba\x6c\xb0\x9d\xc7\x45\x87\xd7\xeb\x45\x1f\x00\x28\x61\x4d\x5c\x92\x02\x6e\x31\xd0\x8f\x3f\x34\xe7\x15\x86\x69\xcf\x2b\x92\x0a\x23\x59\xb2\x45\xa5\x3d\x4c\x4b\xb8\xde\x90\x66\xc3\x13\x9c\x09\xf9\x4a\xa5\xea\xd7\x23\x05\xf8\x3a\x9e\xf0\x0d\x53\x8a\xb8\x83\xbf\x7a\x5b\xbb\x96\x31\x9a\x87\x6d\x9e\xad\xf7\xb0\xe7\xb3\xe1\x46\x68\x99\x97\xe4\x6f\x7b\x7e\x33\xe2\xc7\x49\xe8\xac\x0c\x1d\x43\xe6\x5e\xd7\x48\x46\x3a\xab\x0c\x77\xad\x2e\x6e\x77\x60\x32\xbc\xb4\xba\xae\x48\x68\x54\x55\x2f\x70\x49\xfd\x28\x61\xcd\x54\xcd\x2a\x74\xfd\xbc\xc2\x13\x1f\x29\xfb\xe4\xbb\x93\xe3\x65\xa3\xeb\x4c\x7d\xa0\x70\x49\x4e\xdb\x55\xb5\x03\x4f\x22\x55\xd6\x04\xea\x99\x3d\x39\xad\x04\xee\xd8\x03\x23\xd3\xb4\xd6\xe1\xf1\xb0\x8f\xac\x23\x13\xe6\x6a\xca\xdb\x93\x75\xaf\x42\xd2\xca\x3e\x06\x8c\xca\x1a\xc5\xd6\x2b\x33\x6b\x3d\x13\x8f\x43\xc9\xdc\xc9\xd1\x0f\xa0\x70\xc3\x0b\x4f\xc8\x74\x1c\x5b\xd3\x57\x5b\x6b\x74\x2f\x64\xed\xaa\xff\xfb\xb2\x18\xe2\xd8\xcf\xe3\xb2\x24\x14\xf7\xd9\x9f\xc7\x49\xbd\xd3\x6b\xff\xf7\x8c\x07\x8e\x51\xec\xb3\xfb\x76\x64\xfb\xa4\x7b\xd0\xd9\x7b\x32\x55\x06\xb5\x7a\x4f\xc7\xe8\x52\x3b\xd9\xa6\x47\xa6\xda\x58\x08\x5b\x1b\xbe\xdc\xf7\xc4\xcd\xbe\xad\x64\xdb\x9e\xbd\xf7\x11\xbd\x9d\xbc\xd5\xf9\x53\x19\xe9\xf0\x97\x70\xd6\xb2\x06\xd2\x24\xd8\xfa\xb2\x23\x1f\x2a\x64\x31\xff\x1d\x6f\x49\x87\xb2\x27\x3b\xd0\xb9\xc3\x39\x37\x85\xe7\x27\xe8\x5a\x2d\x6c\xb2\xb1\x47\xa6\xd9\x2a\x2b\x81\xf5\xc5\x54\x39\x8d\x4c\x3d\x10\x2d\x61\xda\xbe\xf4\x5e\x5c\xf7\x42\xf6\x29\x6c\x00\x07\x44\x4f\xc7\xa7\xbf\xaa\xf1\x12\xd6\x30\x2a\x43\x7e\x07\x54\x01\xe8\x67\x7b\xa0\xa6\x2e\xb3\x8b\x4e\xd8\xaa\x42\x23\xf7\x05\xa8\xc5\xae\x3b\x99\xe5\x3e\xd7\x54\xc5\xeb\xf1\xab\xe7\x2f\xff\xbc\x1c\x5f\xbc\xb8\x99\x8c\x9f\xbf\xd8\xf1\x03\x68\x04\xe4\x2f\xde\x56\xbb\x93\xc4\x6b\xaa\x48\xcb\x6b\x9a\xee\x1f\xcd\xe3\x13\xe4\x79\xb9\xd9\xaa\x51\xeb\x05\xe2\xcd\x89\xd5\xd5\x88\xd1\xcf\xa8\xa1\x76\x52\x04\x27\x6f\x0f\x02\x9e\x5c\xfd\xdc\xc0\xfd\x32\x48\x63\xce\x83\x50\xae\x26\x2f\xae\xc7\xaf\xae\xae\x3f\x89\xe7\x38\xd9\x54\x85\x33\x2a\xe1\xaf\x1a\x57\xf1\x2d\x28\xb7\xbf\x27\xbd\xb0\x72\x2f\x15\x73\xf4\xa4\xd6\x7a\x62\xb5\x12\xab\x12\xc6\xfa\x0e\x57\x61\xc7\xef\x38\x69\xa1\xd3\xc5\xba\x67\x0f\x1e\xd4\x87\xdc\xe7\x68\xc7\x87\xb4\xff\xb5\x3e\xbf\xff\x33\x60\x5f\x11\xbd\x59\x1e\x57\xe6\x7e\x85\xaa\xf3\x8b\xab\xfb\x6f\x32\xf7\xb3\x61\x7c\x93\xb9\xdf\x64\xee\xd7\x21\x73\xb7\xea\x70\x9b\x74\xfb\x3d\xe3\xc2\xca\x84\xad\x80\x50\x3b\x67\x3d\x93\x2c\x81\x7d\x9d\x82\xd3\x07\x90\xab\x3b\xb3\xd1\x1a\x47\x5c\x6f\x94\x99\x69\x3a\xe8\x3d\x45\x1d\xda\xee\x17\xb5\x66\x75\xdf\xb9\xc7\x5a\x5f\xb6\xdb\xe4\x82\x56\x77\xd6\xcb\x0c\xbf\xa9\xbd\xb9\x63\xd2\xe4\x3c\x05\x32\x79\x4a\x6d\x05\xea\xf5\xa7\xa5\x5c\x7f\xd1\x7c\x5e\x4a\xdf\x8c\x2a\x54\x6d\xad\x59\x00\x55\xa8\x74\x09\x18\x34\x2e\x7e\xf2\x24\xe7\xb8\x3d\x65\x49\x18\x8c\xe3\x10\xfc\x66\x6a\x0e\x64\xda\x31\xb7\xe8\xd9\x86\xf7\xfb\xa3\x9e\xa5\x41\xb8\xc0\x77\x01\x17\x4d\x66\xae\xbd\xe2\x55\x09\xcd\x57\xb1\x01\x80\xf3\x76\xa9\x24\xe5\x57\x87\x14\x76\x4d\x12\x5e\x22\x9f\xc2\xb9\x11\xa3\x41\x4b\x6e\x27\x89\xf3\x6f\x00\x00\x00\xff\xff\x87\xeb\x75\xe9\xb2\x16\x00\x00")

func deployOlmCatalogIstioWorkspaceManifestsIstioWorkspaceClusterserviceversionYamlBytes() ([]byte, error) {
	return bindataRead(
		_deployOlmCatalogIstioWorkspaceManifestsIstioWorkspaceClusterserviceversionYaml,
		"deploy/olm-catalog/istio-workspace/manifests/istio-workspace.clusterserviceversion.yaml",
	)
}

func deployOlmCatalogIstioWorkspaceManifestsIstioWorkspaceClusterserviceversionYaml() (*asset, error) {
	bytes, err := deployOlmCatalogIstioWorkspaceManifestsIstioWorkspaceClusterserviceversionYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "deploy/olm-catalog/istio-workspace/manifests/istio-workspace.clusterserviceversion.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _deployOlmCatalogIstioWorkspaceManifestsMaistraIo_sessions_crdYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xe4\x58\xc1\x8e\xe3\x36\x0f\xbe\xe7\x29\x08\xfc\x87\xbd\xfc\x49\x30\xe8\xa5\xf0\x6d\x31\xed\x61\xd0\x6d\x31\x98\x14\x7b\x67\x2c\xda\x61\xd7\x96\x54\x89\xca\x36\x2d\xfa\xee\x85\x24\x3b\x71\x1c\x3b\x99\xcc\x62\x81\x62\xcb\x5b\x28\x8a\xa4\xbe\x8f\xa4\x14\xa3\xe5\x8f\xe4\x3c\x1b\x5d\x00\x5a\xa6\x3f\x84\x74\xfc\xe5\x57\x9f\xbe\xf7\x2b\x36\xeb\xfd\xc3\x96\x04\x1f\x16\x9f\x58\xab\x02\x1e\x83\x17\xd3\xbe\x90\x37\xc1\x95\xf4\x03\x55\xac\x59\xd8\xe8\x45\x4b\x82\x0a\x05\x8b\x05\x80\xc6\x96\x0a\xf0\xe4\xb3\xa3\x16\xd9\x8b\xc3\x15\x9b\x85\xb7\x54\x46\x8b\xda\x99\x60\x0b\x18\xac\xe4\x5d\x3e\x2e\x02\xe4\x58\x9b\xec\x20\x69\x1a\xf6\xf2\xd3\x50\xfb\x81\xbd\xa4\x15\xdb\x04\x87\xcd\x29\x5c\x52\x7a\xd6\x75\x68\xd0\x1d\xd5\x0b\x00\x5f\x1a\x4b\x05\xfc\x12\xc3\x58\x2c\x49\x45\x5d\xd8\xba\xee\x2c\x5d\x68\x2f\x28\xc1\x17\xf0\xd7\xdf\x0b\x80\x3d\x36\xac\x30\x9e\x2f\x2f\x1a\x4b\xfa\xfd\xf3\xd3\xc7\xef\x36\xe5\x8e\x5a\xcc\x4a\x00\x45\xbe\x74\x6c\x93\x5d\x9f\x1f\xb0\x07\xd9\x11\x64\x4b\xa8\x8c\x4b\x3f\xfb\x2c\xe1\xfd\xf3\xd3\xaa\xdb\x6e\x9d\xb1\xe4\x84\xfb\x14\xa2\x0c\x68\x39\xea\x46\x81\xde\xc5\x4c\xb2\x0d\xa8\x48\x04\xe5\x88\xfb\xac\x23\x05\x3e\xc7\x36\x15\xc8\x8e\x3d\x38\xb2\x8e\x3c\x69\x49\x27\x1a\xb8\x85\x68\x82\x1a\xcc\xf6\x37\x2a\x65\x05\x1b\x72\xd1\x09\xf8\x9d\x09\x8d\x82\xd2\xe8\x3d\x39\x01\x47\xa5\xa9\x35\xff\x79\xf4\xec\x41\x4c\x0a\xd9\xa0\x50\x47\x47\x2f\xac\x85\x9c\xc6\x26\x62\x18\xe8\xff\x80\x5a\x41\x8b\x07\x70\x14\x63\x40\xd0\x03\x6f\xc9\xc4\xaf\xe0\x67\xe3\x08\x58\x57\xa6\x80\x9d\x88\xf5\xc5\x7a\x5d\xb3\xf4\x85\x58\x9a\xb6\x0d\x9a\xe5\xb0\x2e\x8d\x16\xc7\xdb\x20\xc6\xf9\xb5\xa2\x3d\x35\x6b\xcf\xf5\x12\x5d\xb9\x63\xa1\x52\x82\xa3\x35\x5a\x5e\xa6\xc4\xb5\xe4\x22\x54\xff\x3b\x32\xfd\x6e\x90\xa9\x1c\x62\x51\x78\x71\xac\xeb\xa3\x3a\xd5\xdf\x2c\xee\xb1\x0e\x23\xbd\xd8\x6d\xcb\xf9\x9f\xe0\x8d\xaa\x88\xca\xcb\x8f\x9b\x5f\xa1\x0f\x9a\x28\x38\xc7\x3c\xa1\x7d\xda\xe6\x4f\xc0\x47\xa0\x58\x57\xe4\x32\x71\x95\x33\x6d\xf2\x48\x5a\x59\xc3\x5a\xd2\x8f\xb2\x61\xd2\xe7\xa0\xfb\xb0\x6d\x59\x22\xd3\xbf\x07\xf2\x12\xf9\x59\xc1\x23\x6a\x6d\x04\xb6\x04\xc1\x2a\x14\x52\x2b\x78\xd2\xf0\x88\x2d\x35\x8f\xe8\xe9\xab\xc3\x1e\x11\xf6\xcb\x08\xe9\x6d\xe0\x87\x53\xe4\xdc\x30\xa3\x75\x54\xf7\x83\x64\x92\xa1\xae\x05\x37\x96\xca\xb3\xce\x50\xe4\xd9\xc5\xea\x15\x14\x8a\x35\xdf\x19\xae\x06\x8e\xa6\x9a\x31\x8a\xa3\xea\x5c\x01\xc0\x42\xad\x1f\x2b\x47\xa9\xbc\x50\x75\x25\x85\x38\x15\x30\x0d\xab\x26\xd6\x4f\x45\x8e\x74\x49\x17\x1e\x01\x3e\xb3\xec\x58\xe7\x81\x72\x99\xf3\xf5\xcc\xb3\xa0\xab\x27\xf5\x00\xa8\x54\x1a\xe1\xd8\x3c\x5f\xf5\x00\x73\xa4\x5d\x1a\x8c\xc8\x3a\x49\xba\x19\xae\x6c\x9c\xf5\x1c\xaf\x09\xa1\xfa\xf0\x86\xcd\x57\x52\xca\x4b\xe8\x1c\x1e\xce\xc9\x36\x41\x2e\xf2\x3c\xe7\x35\x5a\x9c\x31\xdb\x67\x98\x48\xdd\x99\xcf\x49\x29\x0e\xab\x8a\xcb\x38\x2d\x92\x4f\x75\x99\x5d\x9e\xa0\x2f\x54\x8d\x19\xbd\xc6\xe7\x1c\x8e\xb7\x81\xb8\x77\x53\x9a\x6d\x77\xee\x9a\x41\x7c\xba\x91\xf3\x7d\x7b\xab\x95\x93\xd5\x19\xde\x66\xeb\xe3\xb4\x7c\x73\x37\x5f\xa0\xfa\xca\x76\x7e\x55\x2a\x51\xcb\x5a\xf1\x9e\x55\xc0\x66\x02\xbd\x09\xbe\xbf\xf1\x0e\x1e\xbd\xb3\xc6\x32\x83\x7e\x96\x31\x07\xfd\xfb\x73\x9a\x85\x63\x24\x68\x43\x24\x44\xad\x4b\x47\x38\xd5\x7c\xbd\xa0\x07\x8b\x4e\x7a\xe6\x26\xc9\xc9\x72\x9d\xa2\xce\x5b\x29\xa3\x77\xdb\x58\x6e\x92\x00\x13\x8f\x90\x37\x39\x99\xa7\xeb\x0e\x27\xf1\xd0\xd7\x9c\xdc\x53\x7f\x77\x84\xbd\x59\x8b\xaf\x32\x99\x9b\xf1\x59\xbe\xe0\x5e\x01\x10\x74\x35\xc9\x17\x17\xf4\x07\xdc\x52\x43\x6a\x58\xd7\xe9\x75\x39\x54\xc4\x07\x40\x36\xf4\xdf\x4a\x69\x36\xe9\x34\xff\xd6\xba\xfa\x0f\xb4\x4e\xfc\x97\x10\x1f\xa3\x73\xf1\x96\x1d\x45\x5f\xa3\xf3\xde\xf0\x28\x4b\x97\xeb\x38\xd7\x19\x34\x26\xdc\x8f\x54\xfb\xfe\xa3\xc7\xfe\x01\x1b\xbb\xc3\x87\x93\x2e\x11\xb0\xec\xbe\x62\x0c\x96\x01\xf2\x05\x53\x80\xb8\x40\xdd\xd7\x02\xe3\xb0\xa6\x4e\xf3\x4f\x00\x00\x00\xff\xff\xb8\x1e\x54\xf0\x4c\x11\x00\x00")

func deployOlmCatalogIstioWorkspaceManifestsMaistraIo_sessions_crdYamlBytes() ([]byte, error) {
	return bindataRead(
		_deployOlmCatalogIstioWorkspaceManifestsMaistraIo_sessions_crdYaml,
		"deploy/olm-catalog/istio-workspace/manifests/maistra.io_sessions_crd.yaml",
	)
}

func deployOlmCatalogIstioWorkspaceManifestsMaistraIo_sessions_crdYaml() (*asset, error) {
	bytes, err := deployOlmCatalogIstioWorkspaceManifestsMaistraIo_sessions_crdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "deploy/olm-catalog/istio-workspace/manifests/maistra.io_sessions_crd.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _deployRole_bindingYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x84\x90\x31\x6f\x1b\x31\x0c\x85\xf7\xfb\x15\x84\xd1\xd5\x57\x74\x2b\xb4\xb9\x85\xd1\xa9\x45\x61\x1b\xd9\x79\x12\xcf\xc7\xf8\x4e\x54\x28\xca\x06\x12\xe4\xbf\x07\x96\xcf\x71\x90\xc1\xd1\x26\x3e\xe2\xbd\xf7\xf1\xc0\x31\x38\xd8\xd1\x94\x46\x34\x6a\x30\xf1\x03\x69\x66\x89\x0e\x6c\x1e\xb6\x92\x28\xe6\x81\x7b\x6b\x59\xbe\x1f\x7f\x34\x09\x15\x27\x32\xd2\xec\x1a\x80\x25\x44\x9c\xc8\xc1\xbf\xd5\xdf\xf5\xf6\xff\xea\xf7\xba\x01\x00\x08\x94\xbd\x72\xb2\xea\xb4\xd8\xa1\xee\xc9\xea\x62\x4e\xe8\x09\x7a\x51\x38\x0d\xec\x07\x50\x19\x09\x3a\x8e\x81\xe3\x1e\xf2\x20\x65\x0c\xd0\x11\x04\xea\x39\x52\x58\x54\x33\xa5\xa7\xc2\x4a\xc1\x81\x69\xa1\x3a\x3a\xe2\x58\xc8\x01\x67\x63\x59\x9e\x44\x0f\xd5\x77\x29\x89\x14\x4d\xb4\x91\xee\x91\xbc\xcd\x05\x2f\x90\x1b\x19\xe9\xd7\x25\xa8\x5a\x7c\x64\xd5\x0e\x7d\x8b\xc5\x06\x51\x7e\xc6\x73\xeb\xf6\xf0\x33\xcf\xbc\xe7\xe5\x89\x0c\x03\x1a\xba\xfa\x83\x99\xf9\x53\x7c\xd5\x72\xb9\x45\x9f\xdf\x35\x7e\x4b\x7a\x64\x4f\x2b\xef\xa5\x44\x9b\xc5\x7b\x46\x57\xb5\x4e\x1c\x7c\x7b\x79\xbf\xf0\xeb\xe5\x2a\x32\xd2\x86\xfa\x6b\xce\x0d\xf2\xcb\x8a\x95\xfd\x8f\x4a\x49\x77\xc8\x9b\xb7\x00\x00\x00\xff\xff\x0b\xb5\x61\x6a\x1b\x02\x00\x00")

func deployRole_bindingYamlBytes() ([]byte, error) {
	return bindataRead(
		_deployRole_bindingYaml,
		"deploy/role_binding.yaml",
	)
}

func deployRole_bindingYaml() (*asset, error) {
	bytes, err := deployRole_bindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "deploy/role_binding.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _deployCluster_role_bindingYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x84\x90\x41\x6f\xdb\x30\x0c\x85\xef\xfe\x15\x44\xb0\x6b\x3c\xec\x36\xe8\x96\x05\xc1\x4e\x1b\x86\x24\xd8\x9d\x96\xe8\x98\x8b\x2d\x6a\x14\x95\x00\x2d\xfa\xdf\x8b\x28\x4e\x5b\xb4\x40\xaa\x9b\xf8\x88\xc7\xef\xbd\x23\xc7\xe0\x60\x4f\x53\x1a\xd1\xa8\xc1\xc4\x7f\x49\x33\x4b\x74\x60\xf3\xb0\x95\x44\x31\x0f\xdc\x5b\xcb\xf2\xf5\xf4\xad\x49\xa8\x38\x91\x91\x66\xd7\x00\x2c\x21\xe2\x44\x0e\x7e\xaf\x7e\x6d\x76\x7f\x56\xeb\x4d\x03\x00\x10\x28\x7b\xe5\x64\xd5\x69\xb1\x47\x3d\x90\xd5\xc5\x9c\xd0\x13\xf4\xa2\x70\x1e\xd8\x0f\xa0\x32\x12\x74\x1c\x03\xc7\x03\xe4\x41\xca\x18\xa0\x23\x08\xd4\x73\xa4\xb0\xa8\x66\x4a\xff\x0b\x2b\x05\x07\xa6\x85\xea\xe8\x84\x63\x21\x07\x9c\x8d\x65\x79\x16\x3d\x56\xdf\xa5\x24\x52\x34\xd1\x46\xba\x7f\xe4\x6d\x06\xbc\x86\x5c\x8f\x25\x1b\xe9\x56\x46\xfa\x71\xbd\x57\x9d\xde\x46\xd6\x0e\x7d\x8b\xc5\x06\x51\x7e\xc0\x0b\x7c\x7b\xfc\x9e\xe7\xd8\x97\xe5\x89\x0c\x03\x1a\xba\xfa\x83\x39\xfa\x3b\x8a\xaa\xe5\xf2\x4a\x70\x79\x37\x8a\x1d\xe9\x89\x3d\xad\xbc\x97\x12\x6d\x16\xef\x19\xdd\xd4\x3a\x71\xf0\xe5\xf1\xa5\xe8\xa7\x6b\x39\x32\xd2\x96\xfa\xdb\x9d\x0f\x59\x3f\x25\xad\x15\xfc\x54\x29\xe9\x4e\x01\xcd\x73\x00\x00\x00\xff\xff\x5a\x3b\x0c\xba\x29\x02\x00\x00")

func deployCluster_role_bindingYamlBytes() ([]byte, error) {
	return bindataRead(
		_deployCluster_role_bindingYaml,
		"deploy/cluster_role_binding.yaml",
	)
}

func deployCluster_role_bindingYaml() (*asset, error) {
	bytes, err := deployCluster_role_bindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "deploy/cluster_role_binding.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
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
	"deploy/cluster_role.yaml":                 deployCluster_roleYaml,
	"deploy/service_account.yaml":              deployService_accountYaml,
	"deploy/operator.yaml":                     deployOperatorYaml,
	"deploy/operator.tpl.yaml":                 deployOperatorTplYaml,
	"deploy/role.yaml":                         deployRoleYaml,
	"deploy/crds/maistra.io_sessions_crd.yaml": deployCrdsMaistraIo_sessions_crdYaml,
	"deploy/crds/maistra.io_sessions_cr.yaml":  deployCrdsMaistraIo_sessions_crYaml,
	"deploy/olm-catalog/istio-workspace/manifests/istio-workspace.clusterserviceversion.yaml": deployOlmCatalogIstioWorkspaceManifestsIstioWorkspaceClusterserviceversionYaml,
	"deploy/olm-catalog/istio-workspace/manifests/maistra.io_sessions_crd.yaml":               deployOlmCatalogIstioWorkspaceManifestsMaistraIo_sessions_crdYaml,
	"deploy/role_binding.yaml":         deployRole_bindingYaml,
	"deploy/cluster_role_binding.yaml": deployCluster_role_bindingYaml,
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
		"cluster_role.yaml":         &bintree{deployCluster_roleYaml, map[string]*bintree{}},
		"cluster_role_binding.yaml": &bintree{deployCluster_role_bindingYaml, map[string]*bintree{}},
		"crds": &bintree{nil, map[string]*bintree{
			"maistra.io_sessions_cr.yaml":  &bintree{deployCrdsMaistraIo_sessions_crYaml, map[string]*bintree{}},
			"maistra.io_sessions_crd.yaml": &bintree{deployCrdsMaistraIo_sessions_crdYaml, map[string]*bintree{}},
		}},
		"olm-catalog": &bintree{nil, map[string]*bintree{
			"istio-workspace": &bintree{nil, map[string]*bintree{
				"manifests": &bintree{nil, map[string]*bintree{
					"istio-workspace.clusterserviceversion.yaml": &bintree{deployOlmCatalogIstioWorkspaceManifestsIstioWorkspaceClusterserviceversionYaml, map[string]*bintree{}},
					"maistra.io_sessions_crd.yaml":               &bintree{deployOlmCatalogIstioWorkspaceManifestsMaistraIo_sessions_crdYaml, map[string]*bintree{}},
				}},
			}},
		}},
		"operator.tpl.yaml":    &bintree{deployOperatorTplYaml, map[string]*bintree{}},
		"operator.yaml":        &bintree{deployOperatorYaml, map[string]*bintree{}},
		"role.yaml":            &bintree{deployRoleYaml, map[string]*bintree{}},
		"role_binding.yaml":    &bintree{deployRole_bindingYaml, map[string]*bintree{}},
		"service_account.yaml": &bintree{deployService_accountYaml, map[string]*bintree{}},
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
