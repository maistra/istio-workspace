package assets

import (
	"io/ioutil"
	"os"
)

// Load loads file asset into byte array
// Assets from given directory are added to the final binary through go-bindata code generation.
//
// If filePath exists locally it will be loaded instead of looked up through go-bindata assets.
func Load(filePath string) (data []byte, err error) {
	if fileExists(filePath) {
		data, err = ioutil.ReadFile(filePath)
	} else {
		data, err = Asset(filePath)
	}
	return data, err
}

func ListDir(path string) ([]string, error) {
	if dirExists(path) {
		dirInfo, e := ioutil.ReadDir(path)
		if e != nil {
			return nil, e
		}

		dirs := []string{}
		for _, info := range dirInfo {
			dirs = append(dirs, info.Name())
		}
		return dirs, nil
	}
	return AssetDir(path)
}

// fileExists checks if a file exists and is not a directory.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// dirExists checks if a file exists and is not a directory.
func dirExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}
