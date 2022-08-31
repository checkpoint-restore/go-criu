package phaul

import (
	"fmt"
	"os"
	"path/filepath"
)

type image struct {
	cursor int
	dir    string
}

//nolint:unparam // suppress "error is always nil" warning
func preparePhaulImages(wdir string) (*image, error) {
	return &image{dir: wdir}, nil
}

func (i *image) getPath(idx int) string {
	return fmt.Sprintf(i.dir+"/%d", idx)
}

func (i *image) openNextDir() (*os.File, error) {
	ipath := i.getPath(i.cursor)
	err := os.Mkdir(ipath, 0o700)
	if err != nil {
		return nil, err
	}

	i.cursor++
	return os.Open(ipath)
}

func (i *image) lastImagesDir() string {
	var ret string
	if i.cursor == 0 {
		ret = ""
	} else {
		ret, _ = filepath.Abs(i.getPath(i.cursor - 1))
	}
	return ret
}
