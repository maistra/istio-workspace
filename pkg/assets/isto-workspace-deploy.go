// Code generated by go-bindata. (@generated) DO NOT EDIT.

// Package assets generated by go-bindata.// sources:
// template/strategies/_basic-remove.tpl
// template/strategies/_basic-version.tpl
// template/strategies/prepared-image.tpl
// template/strategies/prepared-image.var
// template/strategies/telepresence.tpl
// template/strategies/telepresence.var
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

var _templateStrategies_basicRemoveTpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xac\x8e\xbd\x4e\xc4\x30\x10\x06\xfb\x7b\x0a\xcb\xf5\xe9\x4c\x4d\x4d\x41\x49\x81\xe8\x17\xfb\x03\x2c\xc5\x3f\xda\xdd\xa4\xb1\xfc\xee\x28\xa1\x89\x40\x28\x8e\x74\xa5\xad\xfd\x66\xa6\x35\x13\x3f\xcc\xed\x89\x94\x6e\xcf\x24\xc6\x3a\xa9\xf0\x4e\x91\xea\x44\x8a\x9f\x97\x2f\x59\x29\x66\xb0\xb8\x07\x37\xc5\x05\x19\x22\x2f\x5c\xde\x61\x4d\xef\x97\x66\x4b\xb5\x8f\xc6\x32\x52\x59\x60\xaf\xc6\x56\xd2\xaf\xf5\xe7\x24\xac\x5f\x2f\xad\x19\xe4\xb0\x51\x4f\x97\x31\x28\xc4\xbb\xa5\xfd\xa2\x1d\xb4\x25\x28\x05\x52\x72\x0c\x29\x33\x7b\xbc\x81\x25\x96\x7c\x54\xf1\xef\x6e\xd4\xf7\x89\x0c\x26\x3d\xa3\xda\x4d\x46\x2d\x73\x0c\xc3\xf8\xf5\x76\x94\xeb\x19\x5b\xc8\x6b\x4c\x10\xa5\x54\x87\x2d\x7f\x97\x7d\xa7\xfc\x0e\x00\x00\xff\xff\x77\xe1\xaa\x7c\xd7\x02\x00\x00")

func templateStrategies_basicRemoveTplBytes() ([]byte, error) {
	return bindataRead(
		_templateStrategies_basicRemoveTpl,
		"template/strategies/_basic-remove.tpl",
	)
}

func templateStrategies_basicRemoveTpl() (*asset, error) {
	bytes, err := templateStrategies_basicRemoveTplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "template/strategies/_basic-remove.tpl", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _templateStrategies_basicVersionTpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xb4\x94\x4d\x6f\xf2\x30\x10\x84\xef\xfc\x8a\xd5\x9e\xde\x57\x02\x7c\xe7\x5a\x2a\xf5\x50\xf5\xc8\x7d\xeb\x2c\x25\xaa\xbf\x1a\x1b\x2a\x64\xf9\xbf\x57\x09\x1f\x05\x0a\xc5\x31\xf4\x98\xc4\x3b\xf3\x64\x76\xe4\x18\xa1\x9e\x83\xb1\x01\xfe\x8d\xa7\x14\x68\xfc\x44\x1e\x50\x78\xc7\x52\x04\xd6\x4e\x51\x60\xa1\x39\x50\x45\x81\xf0\x3f\xa4\x34\x88\x68\x1d\x4e\x00\xa9\xaa\x70\x08\xe8\x28\x2c\xda\xc7\x4b\x33\x43\xc0\x15\xa9\x25\xe3\x04\x62\x4a\xc3\x41\x8c\xc0\xa6\xea\x84\xfa\x78\x0b\x45\xaf\xac\x7c\x09\xc2\x6e\xf4\x1a\x49\x36\x84\x58\x71\xe3\x6b\x6b\xf0\x90\x45\x5a\xb7\x6e\x2d\xe6\x8d\xd5\xd7\x61\xf6\x12\xd9\xf8\xbb\x89\x91\xb7\xcb\x46\x32\xb6\x3f\xb0\xb5\x6e\xd8\x29\x92\xdc\x5f\xeb\x30\x12\x8c\x71\xfc\xc2\x9f\xb3\xcd\x97\x94\xf0\xe6\x5d\xed\x5d\xca\x77\x76\x5f\x50\xcf\x8a\x65\xb0\x4d\x06\xd0\xfe\x68\x5e\x69\x1e\x3f\x96\xa4\x00\xc5\x7b\x6d\x2a\x04\x9c\xb2\x53\x76\xad\xd9\x84\xae\x22\x00\x39\x54\x42\x53\x90\x8b\xe7\x83\x9a\x03\x64\x31\x1e\x0d\x9e\xf2\x76\xde\x5b\xe2\x1d\x47\x16\xc2\x51\xc9\xbf\x49\x2e\x76\xed\x57\x8d\x2b\xeb\x3b\x07\x99\x1f\xd6\x71\xcf\x0a\x42\x2b\xc1\xec\xd9\x83\x07\x6b\xe6\xf5\x1b\xe6\x6d\xa1\x34\xf9\xbf\x4b\xbb\x28\xe1\xbb\xa5\xba\x21\xca\xb9\x84\xcf\x65\x74\xfb\x95\x92\x23\x6e\x48\xf3\xa9\x64\x07\x3f\x6b\x5f\xfc\x38\x99\xd2\xe8\x8c\xe7\x57\x00\x00\x00\xff\xff\x82\xd4\x0f\xab\x8e\x07\x00\x00")

func templateStrategies_basicVersionTplBytes() ([]byte, error) {
	return bindataRead(
		_templateStrategies_basicVersionTpl,
		"template/strategies/_basic-version.tpl",
	)
}

func templateStrategies_basicVersionTpl() (*asset, error) {
	bytes, err := templateStrategies_basicVersionTplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "template/strategies/_basic-version.tpl", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _templateStrategiesPreparedImageTpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x9c\x90\x31\x6b\xc3\x30\x10\x85\x77\xff\x8a\xc7\x4d\x2d\xb8\x52\xbb\x7a\xee\xd0\x5f\xd0\xa5\x94\x72\x95\xaf\x8d\xc0\x96\x84\xa4\x78\x11\xfa\xef\x41\xb1\x33\x78\x48\x20\x19\x8f\x7b\xdf\xf7\xe0\x7d\x75\x1d\x50\x0a\xb2\xcc\x61\xe2\x2c\xa0\x9f\x5f\x4e\xd6\xbc\x2c\x12\x93\xf5\x8e\xa0\x50\xeb\x16\xb2\x7f\x70\x3e\xe3\x49\xbd\x73\x66\xf5\xc1\x09\xa4\x53\x10\xa3\x2f\xf4\x7a\x45\x09\x93\x35\x9c\xe8\xb9\xa1\x40\x21\x1f\x68\x00\xf1\x38\x52\x0f\x0a\x9c\x0f\xed\xbc\x89\xf6\xa0\x85\xa7\xa3\xd0\x80\x52\x6b\xbf\xf6\x8b\x1b\xf7\xc6\x16\x67\x23\x8f\x58\xe9\x8d\x56\xed\x7d\x2a\xe3\x5d\x66\xeb\x24\x26\xfd\xaa\xed\xcc\xff\xb2\x93\x96\xa2\x3e\x39\x26\x75\xfe\xd4\xda\x2a\xae\xec\x1b\x65\xf6\x8b\x6c\xf3\x7e\x77\xa7\x00\x00\x00\xff\xff\x0a\x21\x5d\x3a\x88\x01\x00\x00")

func templateStrategiesPreparedImageTplBytes() ([]byte, error) {
	return bindataRead(
		_templateStrategiesPreparedImageTpl,
		"template/strategies/prepared-image.tpl",
	)
}

func templateStrategiesPreparedImageTpl() (*asset, error) {
	bytes, err := templateStrategiesPreparedImageTplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "template/strategies/prepared-image.tpl", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _templateStrategiesPreparedImageVar = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xca\xcc\x4d\x4c\x4f\xb5\xe5\x02\x04\x00\x00\xff\xff\xe0\xf3\xaa\xf1\x07\x00\x00\x00")

func templateStrategiesPreparedImageVarBytes() ([]byte, error) {
	return bindataRead(
		_templateStrategiesPreparedImageVar,
		"template/strategies/prepared-image.var",
	)
}

func templateStrategiesPreparedImageVar() (*asset, error) {
	bytes, err := templateStrategiesPreparedImageVarBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "template/strategies/prepared-image.var", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _templateStrategiesTelepresenceTpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xa4\x94\x5f\x6b\xdb\x30\x14\xc5\xdf\xfd\x29\x2e\xf7\x69\x83\xc4\x5e\xdf\x86\xdf\x42\xaa\xb1\xc2\xe6\x85\xb4\xe4\xa5\x94\x70\x63\x5f\x77\x62\xb2\x64\x24\xcd\x1b\x08\x7f\xf7\xe1\x7f\x69\xdd\xd2\x91\xa4\x0f\x09\x28\x9c\xfb\x3b\x07\x9d\x1b\x85\x00\x25\x49\x75\x53\xee\xc8\x4a\x3a\x28\xbe\x36\xec\x32\xe3\xc5\x5f\xe9\x3c\xc4\x3b\xb2\x0e\xb0\x61\xeb\xa4\xd1\x08\xcb\xb6\x8d\xa2\xfb\x08\x20\x04\xf0\x5c\xd5\x8a\x3c\x03\xee\x0f\xe4\x64\xbe\x3c\xaa\x62\xe8\x64\xbd\x48\x96\xa0\x8d\x87\x0f\xf1\x35\x79\x8a\xbf\x92\x03\x4c\x5c\xcd\x79\x32\x4d\x0f\x27\xcb\xb5\x92\x39\x39\xfc\xd8\x8d\x02\x04\x34\x35\xa6\x80\x54\x14\xb8\x00\xac\xc9\xff\xec\x8e\xff\x1d\x5d\x00\x36\xa4\x7e\x33\xa6\x10\xda\x76\x31\xf8\xb3\x2e\xe6\xc4\x4e\x4e\x39\x5f\x42\xc5\x2b\x1c\xb0\x27\x85\xab\xd8\x53\x41\x9e\x12\x45\x07\x56\x2e\xf1\xac\xb8\xb6\xec\x58\x0f\xee\x47\xaa\x67\xe7\xe7\xe0\x13\x33\xe6\x46\x7b\x92\x9a\xad\x4b\x3e\x25\xb2\xa2\xc7\x39\xb7\x33\xff\x23\x2d\xcf\x9c\x97\xbf\x3e\xbb\x34\x84\xbe\xd7\x78\x2c\xac\x6d\x71\xba\xae\x13\xeb\x9a\x39\xb3\x6e\x2e\xa8\xed\x15\xe2\x59\xf4\xfb\x87\x37\xeb\xbb\x80\x9c\x2c\x67\xab\x11\x01\x00\xa0\xa6\xaa\xbf\xa4\x3b\xf1\x4d\x6c\xb6\xe2\x56\x64\x6b\xb1\x5f\xff\xc8\xee\x56\x37\x99\xd8\xee\xb3\xd5\x77\x71\xbb\x59\xad\x05\x2e\x06\x7d\x3f\xfe\xc5\x9a\xea\x88\x00\xc0\x52\xb2\x2a\xb6\x5c\x3e\xfb\x0d\x00\xa9\x96\xbb\xf1\x9f\x90\x02\x36\x57\x23\xe2\x69\x62\x33\x86\x9f\x36\x24\xee\xc2\xb8\xba\x6b\x7c\x54\xb6\xd1\xf4\xdd\x7f\x9e\xca\x39\xa7\x17\xb2\x8f\x0e\x5f\x2e\x7f\x65\x9a\xf3\xf6\xaa\xa7\xbc\xae\xe3\xec\x34\xb9\xa9\x2a\xd2\xc5\xfb\x03\x4d\xa0\x17\x99\xde\x78\x93\x46\x83\xe1\x49\x7a\x88\xfe\x05\x00\x00\xff\xff\xe0\x41\x6e\xb4\xee\x04\x00\x00")

func templateStrategiesTelepresenceTplBytes() ([]byte, error) {
	return bindataRead(
		_templateStrategiesTelepresenceTpl,
		"template/strategies/telepresence.tpl",
	)
}

func templateStrategiesTelepresenceTpl() (*asset, error) {
	bytes, err := templateStrategiesTelepresenceTplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "template/strategies/telepresence.tpl", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _templateStrategiesTelepresenceVar = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x2a\x4b\x2d\x2a\xce\xcc\xcf\xb3\xe5\x02\x04\x00\x00\xff\xff\xa0\x0b\xe9\x34\x09\x00\x00\x00")

func templateStrategiesTelepresenceVarBytes() ([]byte, error) {
	return bindataRead(
		_templateStrategiesTelepresenceVar,
		"template/strategies/telepresence.var",
	)
}

func templateStrategiesTelepresenceVar() (*asset, error) {
	bytes, err := templateStrategiesTelepresenceVarBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "template/strategies/telepresence.var", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
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
	"template/strategies/_basic-remove.tpl":  templateStrategies_basicRemoveTpl,
	"template/strategies/_basic-version.tpl": templateStrategies_basicVersionTpl,
	"template/strategies/prepared-image.tpl": templateStrategiesPreparedImageTpl,
	"template/strategies/prepared-image.var": templateStrategiesPreparedImageVar,
	"template/strategies/telepresence.tpl":   templateStrategiesTelepresenceTpl,
	"template/strategies/telepresence.var":   templateStrategiesTelepresenceVar,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//
//	data/
//	  foo.txt
//	  img/
//	    a.png
//	    b.png
//
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
	"template": &bintree{nil, map[string]*bintree{
		"strategies": &bintree{nil, map[string]*bintree{
			"_basic-remove.tpl":  &bintree{templateStrategies_basicRemoveTpl, map[string]*bintree{}},
			"_basic-version.tpl": &bintree{templateStrategies_basicVersionTpl, map[string]*bintree{}},
			"prepared-image.tpl": &bintree{templateStrategiesPreparedImageTpl, map[string]*bintree{}},
			"prepared-image.var": &bintree{templateStrategiesPreparedImageVar, map[string]*bintree{}},
			"telepresence.tpl":   &bintree{templateStrategiesTelepresenceTpl, map[string]*bintree{}},
			"telepresence.var":   &bintree{templateStrategiesTelepresenceVar, map[string]*bintree{}},
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
