package source

import (
	"fmt"
	"os"
	"path"
)

// NewImageSource detects the type of the target and creates the appropriate ImageSource
func NewImageSource(target string) (ImageSource, error) {
	fileInfo, err := os.Stat(target)
	if err != nil {
		return nil, err
	}
	if fileInfo.IsDir() {
		return NewFolderImageSource(target)
	}
	if path.Ext(fileInfo.Name()) == ".zip" {
		return NewZipImageSource(target)
	}
	return nil, fmt.Errorf("unrecognized input type %s", target)
}
