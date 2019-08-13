// Package main Code generated by go-bindata. (@generated) DO NOT EDIT.
// sources:
// test/cmd/test-service/assets/index.html
package main

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

var _indexHtml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xec\x58\x5f\x73\xe3\xb6\x11\x7f\xf7\xa7\xd8\x30\x33\xb1\x9c\x8a\xa4\x24\xff\x3b\xab\xa4\x26\x8d\x73\xed\xa4\xd3\xce\x25\xf1\xdd\x4d\x7a\x19\x4f\x03\x91\x2b\x09\x36\x08\xb0\x00\x48\xcb\xe3\xd1\x77\xef\x00\x20\x29\x92\xd2\x59\xbe\xeb\x4b\x7b\xd3\x97\x33\xb8\xd8\x5d\xec\xef\x87\xc5\xee\xea\xa2\xaf\x7e\x78\x73\xfd\xf6\x1f\x3f\xbd\x86\x95\xce\xd8\xec\x28\x32\x7f\x80\x11\xbe\x8c\x3d\xe4\xde\xec\x08\x20\x5a\x21\x49\xed\x22\x43\x4d\x20\x59\x11\xa9\x50\xc7\x5e\xa1\x17\xfe\x2b\x6f\xbb\xb1\xd2\x3a\xf7\xf1\x5f\x05\x2d\x63\xef\x57\xff\xdd\x9f\xfc\x6b\x91\xe5\x44\xd3\x39\x43\x0f\x12\xc1\x35\x72\x1d\x7b\x3f\xbe\x8e\x31\x5d\x62\xcb\x8e\x93\x0c\x63\xaf\xa4\xf8\x90\x0b\xa9\x5b\xaa\x0f\x34\xd5\xab\x38\xc5\x92\x26\xe8\xdb\x8f\x21\x50\x4e\x35\x25\xcc\x57\x09\x61\x18\x8f\x77\xdc\xa4\xa8\x12\x49\x73\x4d\x05\x6f\x79\xa2\x4a\x53\xe1\x3f\x08\x79\xaf\x72\x92\x20\xbc\x45\xa5\xe1\x06\xa5\xf1\x0c\x8a\x66\x39\x43\x78\xf7\xa3\x37\x3b\x32\xde\x34\xd5\x0c\x67\x3d\x93\x69\xc7\x26\x0a\x9d\x92\xd5\x67\x94\xdf\x83\x44\x16\x7b\x4a\x3f\x32\x54\x2b\x44\xed\xc1\x4a\xe2\x22\xf6\x0c\x27\x6a\x1a\x86\x49\xca\xef\x54\x90\x30\x51\xa4\x0b\x46\x24\x06\x89\xc8\x42\x72\x47\xd6\x21\xa3\x73\x15\xce\x0b\x96\x91\x70\x14\x5c\x06\xe7\x61\xa2\xaa\xef\x20\x51\xca\x03\xca\x35\x2e\x25\xd5\x8f\xb1\xa7\x56\x64\x72\x7e\xe1\x17\x77\xaf\xc3\x0f\xef\xbe\xbf\xb8\xfe\xfb\x87\xec\xf1\xaf\x37\xcb\x9f\xef\x7e\xfd\xcb\xf5\x9f\xcf\xd4\x2f\xbf\x70\x7c\x43\xb3\x9f\x73\xf6\xfd\xfb\xbf\x15\xaf\xde\xbc\x3b\x7f\x88\x3d\x48\xa4\x50\x4a\x48\xba\xa4\x3c\xf6\x08\x17\xfc\x31\x13\x85\xf2\x20\xb4\xcc\x39\xb2\x40\xc9\x64\x1b\x6b\x59\xe0\x9d\x0a\x84\x5c\x86\x77\xca\x7c\x04\x77\xca\x9b\x45\xa1\x53\xfd\xa8\x55\xc1\xf3\xfb\xa5\x85\x55\x16\xb8\xfe\x6e\x12\x8c\x82\xd1\x27\x99\x91\x35\x15\xea\xbb\x51\x30\xbe\x0a\x46\x61\x4a\x95\x76\x92\x20\xa3\xbc\x17\x81\xf5\x65\xa8\x36\x5e\x01\x52\x5a\xc2\x93\x5d\x01\xac\xe7\x42\xa6\x28\xa7\x30\xce\xd7\xa0\x04\xa3\x29\x7c\x3d\x1a\x8d\xec\xee\xc6\xd8\x85\x8d\x61\x14\x36\x79\x3d\x17\xe9\xa3\xf3\x15\x19\x67\x34\x8d\x3d\x92\xe7\xde\xac\x72\x0a\x10\x71\x52\x42\xc2\x88\x52\xb1\xc7\x49\x39\x27\xb2\xb5\xd9\xd8\x75\x14\xfc\xb9\x24\x3c\xed\xa9\x01\x44\xa4\xa7\x46\x35\x66\xfd\x6c\x59\x52\xbd\x2a\xe6\x96\x95\x8c\x50\xa5\x25\x09\x7b\xe9\xb8\xe3\x17\xa0\x9f\xe3\x3e\x68\x93\xb1\xca\x65\x6c\x3f\x8c\x90\xf4\x00\x84\x29\x2d\x2d\xb5\xcf\x82\xca\x90\x17\xbb\x98\x76\xf5\x70\x0f\xf4\xbd\x8a\x16\xfe\xae\x66\x57\x57\x21\xc3\x44\xef\x55\x33\xa9\x60\x77\xa1\xf4\x33\x91\xda\x57\x88\x4a\x99\xe7\xbf\x5f\x1d\x20\x12\xb6\x3c\x40\x4a\x15\x99\x33\x4c\x67\x37\xce\x41\x65\x17\x85\x6e\xff\x90\x79\xe9\x2f\x84\x6c\x4e\x03\xca\x6b\x07\xca\x83\xd2\x9f\x53\x9e\x4e\x4b\xc2\x0a\x6c\x54\x82\x42\x32\x6f\xf6\xf4\x54\x7f\x9a\x82\xb5\xd9\x3c\x7f\x5c\x14\x3a\x74\x7b\x19\x72\x57\xf6\x22\xf1\x1e\xe1\x8e\x28\x0a\x39\x69\xa7\x40\xa4\x34\x49\xee\x6b\x2c\x0b\x69\xeb\xab\x95\xb9\xd7\x68\x56\xd5\xab\xa9\x5d\x45\xa1\x7b\x4a\x51\xe8\xba\xc9\x51\xfd\xe6\x5d\x43\xd1\xb8\xd6\xe1\x1d\x29\x89\x93\x9a\x7a\x9b\x08\x6e\x92\x54\x0b\x89\x10\x03\xc7\x07\x78\x5f\xe0\x3a\xb8\x31\x82\x81\x79\xd5\x4a\x13\x8d\xd3\xea\x81\xd7\x0c\x4f\xe1\xb7\xdb\xa1\x93\x98\x28\xa6\xc0\x0b\xc6\x8c\x60\x63\xfe\xc9\x0a\x4d\xb4\x53\x73\x66\x45\x9e\x12\x8d\x37\x46\x75\x60\xfd\x0d\x21\x27\x8f\x4c\x90\xf4\xa4\xa9\x1c\x56\x1e\x38\xc8\x71\xbd\xed\xea\xc6\xb0\xed\xa4\x8a\xe0\x90\x9f\x4a\xad\xef\xaa\x0e\x91\x24\x9d\x00\x25\x2e\x24\xaa\x95\x8b\xd0\x76\xad\xb5\xde\xba\x94\xa8\x0b\xc9\x2d\x39\x3f\x49\x91\x51\x85\x83\x81\x44\x25\x58\x89\x27\x10\xcf\x1a\x3d\x00\x57\x31\x97\xa8\x07\x0f\x94\xa7\xe2\x21\x60\x22\xb1\x54\x0c\xe1\x69\x73\x12\xe8\x15\x72\x6b\x9a\x0b\xae\xfa\xb6\x00\xd5\xc1\xa6\xee\x64\x54\x0f\x8e\x5b\xb4\x1d\x0f\xa1\x36\x0b\x52\xa2\xc9\xc9\x1f\x5b\x76\x55\x2c\x83\x96\x70\xd3\xac\xeb\x55\xc5\x62\x8d\xb4\xa6\xf1\x3f\x00\x1b\x7e\x7b\x18\x37\xfc\x01\xbc\xf0\x9f\xcd\xc3\xfc\x5c\x0a\x2a\xfb\xcf\x67\x01\xe0\xdb\xf0\xe8\xc5\x87\xfc\xd6\x72\xfb\xd4\x79\xb4\xa6\x66\x4c\xc1\xcb\xa5\x48\xbd\x61\x67\xa7\x90\x6c\x0a\xb6\x83\x4c\xc3\x90\x28\x46\xee\x7d\x53\xff\x6d\x13\xe9\xa8\x6e\x86\x87\xbc\x2f\x90\xe8\x42\xa2\xbf\x7e\xee\x88\x46\x29\xe8\x1f\xd6\x3e\xab\x59\xdf\xb6\xa8\xd8\x61\xaa\xc9\x11\xf3\x3c\x8e\x36\x27\x47\x47\xef\x0b\x3b\x22\xe5\x82\x23\xd7\x83\x63\x57\x75\x8e\x87\x36\x5e\x8d\x59\xce\x6c\x51\xf8\x7d\xdb\xb3\xab\x4e\x91\x08\x56\x64\x5c\x35\xb5\xbf\xa9\x61\xb6\x56\xdb\x1a\x66\x2a\xb5\x5d\x6c\xeb\x74\x55\xdb\xec\x9f\x46\x78\x8f\x8f\x95\x28\x48\x08\x63\x28\x3f\x56\xf4\x7e\x37\x2c\xe5\x52\xe4\xa6\x2c\x1d\x3b\xd7\xc7\xb7\xc3\x8f\xe2\x78\x31\x0c\x13\x0a\x5d\xd4\x71\x35\x90\xda\x7a\x44\xa6\x4d\xc0\x76\xb2\x89\xbd\xa7\x39\x49\xee\x97\x52\x14\x3c\xbd\x16\x4c\xc8\x29\x54\x20\xcc\xc7\xa6\x3d\xd4\x98\xf9\x07\x65\xdb\x97\xef\x44\x9d\xd6\x19\xe5\x7b\x34\x7c\x3b\xff\x9a\x6e\xd6\x66\xc8\x74\xb3\xfc\x13\x6c\x71\x8d\x49\x61\x9e\xe9\x5b\x6a\x5a\x21\x64\xaa\x63\xef\x26\x34\x94\x2d\x49\x0f\xbb\x5f\x4d\xf8\x35\x53\x2b\xa2\xae\x4d\x24\xdd\xe1\xc3\x25\x81\xea\xdc\xb6\xea\xdc\x6d\xda\xdc\xad\x6a\x1f\xdf\x69\x90\xd1\x42\x08\xdd\xa3\xcb\x89\x9e\xa3\xcb\x69\x54\x63\x4e\xcd\x56\x4e\xf4\xaa\xc7\x55\x14\x3a\xcd\xe6\x96\xb7\x87\x3f\x97\x68\xc7\xb6\x0d\x9a\x0c\x2b\x34\xa6\x75\x3b\xe9\xf0\x3a\x85\x45\xc1\x6d\xbb\x19\xec\xad\xb2\x3f\x10\x8d\x03\xbd\xa2\x2a\x70\xc1\x21\x4f\x8d\xd9\x09\xf8\x7b\xb7\x95\x26\x52\x5b\x85\x4e\x5d\x6f\x98\x7f\xee\xb8\x96\x1b\x47\x3b\x7c\x15\xdb\xd6\x0d\xdf\x7c\xb3\xbb\x19\x30\xe4\x4b\xbd\x82\x19\x8c\x5a\x47\x6d\xdc\xc3\x2a\x89\x04\x92\xe7\xdb\x99\xc1\x0e\x0b\x68\x0a\xd4\xd7\x66\x76\x37\xaa\xa6\x44\xf7\x26\x87\x29\xf4\x1b\x63\xdd\x91\x1f\x88\x4e\x56\x3b\xda\x0d\x96\x92\xb0\x2d\x9c\x7e\x93\x89\xa1\x24\xac\x1b\x23\x40\x22\x91\xe8\xbd\x7c\xd8\x69\x27\x48\xa9\xca\xcd\x99\x83\xe3\x76\xf7\x3f\x3e\x79\x56\xa5\xee\x11\x27\xcd\x29\xbd\xab\x77\x49\xfc\xdc\x25\x38\xd7\xad\x71\xa7\x73\x8f\xdb\x09\xeb\x85\x1e\x2a\xfd\x7e\x01\xaf\xad\x61\x89\xda\x8d\x35\xce\x4d\xe5\xc2\x79\xf4\xaa\xc2\x5a\x75\xb4\x22\xd1\x39\x59\xa2\x5f\x8e\xab\xc6\xe3\x99\x87\x62\x76\xeb\x0e\xe6\x35\xd9\x67\xa4\x93\xd1\xf8\xca\x1f\x5d\xfa\xe3\xd1\xdb\xf1\xc5\xf4\xf4\x74\x3a\x7a\x15\x9c\x4d\x2e\xcf\xc6\xe7\x93\xd1\xc5\x87\xda\xa4\xca\xe7\x8f\x1a\x9c\x9e\x4d\xae\x4e\xcf\xc6\x93\xc6\xa0\xaa\x08\xd3\x56\x1b\xee\xb6\xc9\x56\xd8\x12\x4b\x8a\x0f\x6a\x1b\x72\xa3\xd3\x0f\xbd\xd9\x78\x01\x84\x57\x57\xe3\x8b\xab\xcb\xcb\x0f\x7d\xd3\x83\x50\x26\x67\xa7\xa3\xf3\xf1\x78\xc7\x70\x0b\xe9\x76\xbb\xd3\x1a\x04\x0e\xe3\x9b\xfc\x8f\xe0\xdb\xf9\x39\xf4\xb4\xf7\x77\x55\x1b\x22\xd1\x94\x2f\xf7\x5c\xe1\x41\xa8\x9f\x00\xf9\x74\x74\x71\x3e\x1e\x9d\xef\x40\x7e\x39\x74\xe3\xe0\x74\xb2\x03\x7d\x0f\x05\xb7\xbb\x1a\x9b\x5d\xd1\x41\x5e\x52\xd4\x84\xb2\xff\xf3\xd2\xe7\xc5\x0c\xba\x5f\x36\x29\x1d\xc9\xe7\x16\x8c\xd3\x2f\xbf\x60\xf4\x6b\xe2\x41\xa8\x9f\x00\xf9\xbf\x3c\x07\x6e\x5d\xa3\x3f\xda\xfe\xa7\xec\xbf\x03\x00\x00\xff\xff\xc7\x06\xac\xfe\x49\x18\x00\x00")

func indexHtmlBytes() ([]byte, error) {
	return bindataRead(
		_indexHtml,
		"index.html",
	)
}

func indexHtml() (*asset, error) {
	bytes, err := indexHtmlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "index.html", size: 6217, mode: os.FileMode(436), modTime: time.Unix(1565694963, 0)}
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
	"index.html": indexHtml,
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
	"index.html": &bintree{indexHtml, map[string]*bintree{}},
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
