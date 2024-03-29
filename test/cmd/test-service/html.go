// Code generated for package main by go-bindata DO NOT EDIT. (@generated)
// sources:
// test/cmd/test-service/assets/index.html
package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("read %q: %w", name, err)
	}
	if clErr != nil {
		return nil, fmt.Errorf("read %q: %w", name, clErr)
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

var _indexHtml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xec\x58\x6d\x6f\xe3\xb8\x11\xfe\xee\x5f\x31\xa7\x02\x67\xe7\x60\x49\x7e\xc9\xab\xd7\x32\xae\x97\xdb\x16\x57\xb4\xd8\x97\x24\x8b\x6d\x16\x01\x96\x96\xc6\x16\x1d\x89\x54\x49\x4a\x76\x10\xf8\xbf\x17\x24\x65\x59\x96\x9d\xb7\x7e\x6a\x17\xfd\x12\x4b\xc3\x99\xe1\x3c\x0f\x87\xcf\xd8\x19\xff\xf4\xfb\x87\xcb\xeb\x7f\x7e\x7c\x0f\xb1\x4a\x93\x49\x6b\xac\x3f\x20\x21\x6c\x1e\x38\xc8\x9c\x49\x0b\x60\x1c\x23\x89\xcc\x43\x8a\x8a\x40\x18\x13\x21\x51\x05\x4e\xae\x66\xee\xb9\xb3\x5d\x88\x95\xca\x5c\xfc\x57\x4e\x8b\xc0\xf9\xea\xde\xfc\xd9\xbd\xe4\x69\x46\x14\x9d\x26\xe8\x40\xc8\x99\x42\xa6\x02\xe7\x8f\xf7\x01\x46\x73\xac\xc5\x31\x92\x62\xe0\x14\x14\x97\x19\x17\xaa\xe6\xba\xa4\x91\x8a\x83\x08\x0b\x1a\xa2\x6b\x5e\xba\x40\x19\x55\x94\x24\xae\x0c\x49\x82\x41\x7f\x2f\x4d\x84\x32\x14\x34\x53\x94\xb3\x5a\x26\x2a\x15\xe5\xee\x92\x8b\x7b\x99\x91\x10\xe1\x1a\xa5\x82\x2b\x14\x3a\x33\x48\x9a\x66\x09\xc2\xcd\x1f\xce\xa4\xa5\xb3\x29\xaa\x12\x9c\x34\x42\x46\x3b\x31\x63\xdf\x3a\x19\xff\x84\xb2\x7b\x10\x98\x04\x8e\x54\x0f\x09\xca\x18\x51\x39\x10\x0b\x9c\x05\x8e\xe6\x44\x8e\x7c\x3f\x8c\xd8\x42\x7a\x61\xc2\xf3\x68\x96\x10\x81\x5e\xc8\x53\x9f\x2c\xc8\xca\x4f\xe8\x54\xfa\xd3\x3c\x49\x89\xdf\xf3\xce\xbc\x13\x3f\x94\xe5\xbb\x17\x4a\xe9\x00\x65\x0a\xe7\x82\xaa\x87\xc0\x91\x31\x19\x9c\x9c\xba\xf9\xe2\xbd\x7f\x7b\xf3\xdb\xe9\xe5\x3f\x6e\xd3\x87\xbf\x5d\xcd\x3f\x2d\xbe\xfe\xf5\xf2\x2f\xc7\xf2\xf3\x67\x86\x1f\x68\xfa\x29\x4b\x7e\xfb\xf2\xf7\xfc\xfc\xc3\xcd\xc9\x32\x70\x20\x14\x5c\x4a\x2e\xe8\x9c\xb2\xc0\x21\x8c\xb3\x87\x94\xe7\xd2\x01\xdf\x30\x67\xc9\x02\x29\xc2\x6d\xad\x45\x8e\x0b\xe9\x71\x31\xf7\x17\x52\xbf\x78\x0b\xe9\x4c\xc6\xbe\x75\x7d\x32\x2a\x67\xd9\xfd\xdc\xc0\x2a\x72\x5c\xfd\x3a\xf0\x7a\x5e\xef\x4d\x61\x64\x45\xb9\xfc\xb5\xe7\xf5\x2f\xbc\x9e\x1f\x51\xa9\xac\xc5\x4b\x29\x6b\x54\xa0\x73\xf9\x55\x53\x4e\x79\xf4\xa0\x1f\x00\xc6\x11\x2d\x80\x46\x81\x43\xb2\xcc\xb1\x26\x80\x31\x23\x05\x84\x09\x91\x32\x70\x18\x29\xa6\x44\x54\x4b\x65\xc4\xce\xa2\x3b\x15\x84\x45\x35\x17\x80\x31\x69\xb8\x50\x85\x69\xf3\x80\xe7\x54\xc5\xf9\xd4\x00\x49\x09\x95\x4a\x10\xbf\xd1\x41\x3b\x39\x01\x9a\x2d\xe9\x82\xd2\x0d\x26\x6d\x83\xd5\xb7\xf7\x49\xad\x60\x3f\xa2\x45\x05\xcd\x67\xa4\x30\x74\x98\x37\xa9\x48\x78\x0f\x85\x3b\xa5\x2c\x1a\xcd\x84\xb9\x0e\xc6\x66\xc9\xd3\x4f\x25\x4f\x9b\x24\x63\xdf\x92\x37\xf6\xed\xe5\x6f\x6d\x8e\xc8\xde\x7f\x85\x2b\xe5\x2f\x48\x41\xac\x55\x5f\x8f\x90\x33\x5d\xa4\xe2\x02\x21\x00\x86\x4b\xf8\x92\xe3\xca\xbb\xd2\x86\xce\x63\x0b\x40\x2a\xa2\x70\x04\x8f\x66\x23\x89\x52\x52\xce\xe4\x08\xbe\xdd\x75\xad\x45\x57\x31\x02\x96\x27\x89\x36\xac\xf5\x9f\x34\x57\x44\x59\x37\x1b\x96\x67\x11\x51\x78\xa5\x5d\x3b\x26\x5f\x17\x32\xf2\x90\x70\x12\x1d\x95\x1e\xe5\x3e\x9e\x85\x1c\x6c\x96\xcd\xda\xba\x5b\x4f\x52\x56\xf0\x52\x9e\xd2\xad\x99\x6a\x53\x22\x09\x77\x0a\x14\x38\x13\x28\x63\x5b\xa1\x11\x99\x95\xda\xa6\x14\xa8\x72\xc1\x0c\x39\x1f\x05\x4f\xa9\xc4\x4e\x47\xa0\xe4\x49\x81\x47\x10\x4c\x2a\x3f\x00\xdb\xe0\x73\x54\x9d\x25\x65\x11\x5f\x7a\x09\x0f\x0d\x15\x5d\x78\x5c\x1f\x79\x2a\x46\x66\x42\x33\xce\x64\x33\x16\xa0\xdc\x58\xf7\x5c\x4a\x55\xa7\x5d\xa3\xad\xdd\x85\x4d\x98\x17\x11\x45\x8e\xde\xd5\xe2\xca\x5a\x3a\x35\xe3\xba\x7a\xde\x3c\x19\xd8\xeb\xd6\xfa\xa8\xd5\xfa\x92\x1b\xb9\xca\x38\x43\xa6\x3a\x6d\xdb\x52\xed\xae\xa9\x45\x61\x9a\x25\xe6\xc4\xbf\xb7\x9a\x17\x2a\xe4\x49\x9e\x32\xb9\xbd\x89\x9b\x06\x9d\x71\x11\x38\xa6\x41\x81\x32\x30\x0f\xd2\x69\x34\xae\xf9\xa8\x8c\xf7\xf8\x50\x9a\xbc\x90\x24\x09\x8a\xa7\x3a\xfa\xbb\x2e\x3b\x13\x3c\xd3\x3d\xd7\xb6\xa9\xdb\x77\xdd\x27\x71\xbc\x1a\x86\x2e\x85\xce\x36\x75\x55\x90\xea\x7e\x44\x44\x55\xc1\x66\x12\x04\xce\xe3\x94\x84\xf7\x73\xc1\x73\x16\x5d\xf2\x84\x8b\x11\x94\x20\xf4\xcb\xba\x2e\x44\x5a\xce\x50\xd4\x73\xb9\xd6\xb4\x2b\x45\xd9\x01\x0f\xd7\xcc\x22\x67\xf2\xf8\x58\x67\x68\xbd\x1e\xfb\xd9\x1b\x62\x71\x85\x61\xae\x7b\xef\x9a\xa6\xb8\x5e\x43\x2a\x77\xe2\xad\xe0\xa2\x38\xac\x9d\x26\x63\x39\x6d\x37\x4c\xc5\x44\x5e\xea\x4a\x1a\x62\x6a\xdb\x67\xe7\xb4\xe5\xce\xd9\x46\xd5\xd9\xca\x27\x74\x0f\x60\x3c\xe3\x5c\x35\xe8\xb2\xa6\xe7\xe8\xb2\x1e\x56\xbe\x2b\xb6\x32\xa2\xe2\x06\x57\x63\xdf\x7a\x6e\x75\xb6\xda\xfc\xb9\x46\x6b\x1b\x8d\xd3\x1d\x96\x2b\x8c\x36\x5a\xb1\xc3\xeb\x08\x66\x39\x33\x5a\xd2\x39\xa8\x17\xbf\x13\x85\x1d\x15\x53\xe9\xd9\xe2\x90\x45\x3a\xec\x08\xdc\x83\xcb\x52\x11\xa1\x8c\xc3\xbb\xba\xf4\x55\xcc\x3f\xb7\x5d\x2d\x8d\xa5\x1d\x7e\x0a\x8c\x2e\xc3\xcf\x3f\xef\x2f\x7a\x09\xb2\xb9\x8a\x61\x02\xbd\xba\x3e\xd8\x8b\x55\x10\x01\x24\xcb\xb6\x03\xc1\x4c\x02\x4c\x46\xe0\xfc\x49\x8f\x62\xed\xaa\x55\xa8\x31\x16\x46\xd0\x54\xbd\x8d\xdc\x2e\x89\x0a\xe3\x3d\xef\x0a\x4b\x41\x92\x2d\x9c\x46\x0e\x08\xa0\x20\xc9\x6e\x8d\x00\xa1\x40\xa2\x0e\xf2\x61\x46\x99\x17\x51\x99\xe9\x3d\x3b\xed\xba\xb4\xb7\x8f\xaa\xf8\xc6\xa1\xda\xf6\x7c\x8e\x5e\x9b\xb7\x36\xa5\x76\x4e\x68\x3b\x18\x5f\x99\xa1\xf4\x7f\xb7\x1d\x4b\x86\x79\xff\x17\x43\x2c\xe0\x8a\xe8\xef\xb0\x2d\x9b\xc2\x29\x35\x72\x04\x4e\x26\x78\x94\x87\x2a\x23\x73\x74\x8b\xbe\x63\x77\x77\x74\xcf\xeb\x55\x7f\x63\xa8\x1a\x49\x5b\x07\xbd\xfe\x85\xdb\x3b\x73\xfb\xbd\xeb\xfe\xe9\x68\x38\x1c\xf5\xce\xbd\xe3\xc1\xd9\x71\xff\x64\xd0\x3b\xbd\xdd\x84\x94\xad\xf9\x64\xc0\xf0\x78\x70\x31\x3c\xee\x0f\xaa\x80\xf2\x72\x8f\xe0\x5b\x75\xd1\x1e\x77\xbe\x11\xd5\xca\x16\xa8\x7f\x0e\xc8\x6d\xc9\x95\x4f\xb3\xf4\x6a\xe1\x15\x10\xce\x2f\xfa\xa7\x17\x67\x67\xb7\xcd\xd0\x17\xa1\x0c\x8e\x87\xbd\x93\x7e\x7f\x2f\x70\x0b\xe9\x6e\xbb\xb2\xee\xbe\x01\xdf\xe0\x7f\x04\xdf\x8e\x7d\x1f\xda\x21\x88\x44\x51\x36\x3f\x70\x84\x2f\x42\x7d\x03\xe4\x61\xef\xf4\xa4\xdf\x3b\xd9\x83\xfc\x7a\xe8\x3a\xc1\x70\xb0\x07\xfd\x00\x05\x77\xfb\x1e\xeb\x7d\xd3\x8b\xbc\x44\xa8\x08\x4d\xfe\xcf\x4b\x93\x17\xfd\xeb\xe7\xc7\x26\x65\xc7\xf2\x9f\x0a\xc6\xf0\xc7\x17\x8c\xa6\x26\xbe\x08\xf5\x0d\x90\xff\xcb\x7b\xe0\xce\x4e\xf6\xd6\x2f\x7e\x6b\xfb\xef\x8e\x7f\x07\x00\x00\xff\xff\x6c\x58\x4e\xc8\xa3\x13\x00\x00")

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

	info := bindataFileInfo{name: "index.html", size: 5027, mode: os.FileMode(436), modTime: time.Unix(1585740871, 0)}
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
		return fmt.Errorf("read %q: %w", name, err)
	}
	err = os.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return fmt.Errorf("read %q: %w", name, err)
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return fmt.Errorf("read %q: %w", name, err)
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
