package main

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"github.com/ivaxer/go-xattr"
)

func OptimizeLayer(i SirenImage, statusCb func(string)) {
	if err := removeUnchanged(i, statusCb); err != nil {
		statusCb(err.Error())
	}

	if err := removeEmpty(i, statusCb); err != nil {
		statusCb(err.Error())
	}
}

func removeUnchanged(i SirenImage, statusCb func(string)) error {
	if i.base == nil {
		return nil
	}

	return filepath.Walk(i.LayerFSRoot(), func(thisPath string, this os.FileInfo, err error) error{
		path := strings.TrimPrefix(thisPath, i.LayerFSRoot())
		basePath := i.base.RealPath(path)

		base, _ := os.Stat(basePath)

		if this.IsDir() {
			if base == nil {
				// New directory.
				return filepath.SkipDir
			}

			iO, err := isOpaqueDirectory(thisPath)
			if err != nil {
				statusCb("Cannot check opaque-ness: " + thisPath)
				return filepath.SkipDir
			}
			if iO {
				statusCb("Found opaque subtree: " + path)
				// TODO Optimize opaque subtrees too.
				return filepath.SkipDir
			}

			return nil // Don't remove directories.
		}

		if base == nil {
			// New file.
			return nil
		}

		if this.Size() != base.Size() || this.Mode() != base.Mode() {
			return nil // different!
		}

		baseContent, err := ioutil.ReadFile(basePath)
		if err != nil {
			return nil // whatever!
		}
		thisContent, err := ioutil.ReadFile(thisPath)
		if err != nil {
			return nil // whatever!
		}

		if !reflect.DeepEqual(baseContent, thisContent) {
			return nil // different!
		}

		// identical!
		statusCb("Removing unchanged file: " + path)
		os.Remove(thisPath)
		return nil
	})
}

func removeEmpty(i SirenImage, statusCb func(string)) error {
	if i.base == nil {
		return nil
	}

	// TODO Find a Walk that calls one func at the beggining of the directory,
	// and a second one at the end.
	// This way we will have to traverse the tree only once.

	removedDirs := map[string]bool{}

	removedAnything := true
	for removedAnything {
		removedAnything = false
		filepath.Walk(i.LayerFSRoot(), func(thisPath string, this os.FileInfo, err error) error{
			if !this.IsDir() {
				// We don't care about files.
				return nil
			}

			if removedDirs[thisPath] {
				// os.Walk is insane, after we remove anything - it calls us back with its name.
				return nil
			}

			path := strings.TrimPrefix(thisPath, i.LayerFSRoot())
			basePath := i.base.RealPath(path)

			base, _ := os.Stat(basePath)

			if base == nil {
				// New directory.
				return filepath.SkipDir
			}

			iED, err := isEmptyDirectory(thisPath)
			if err != nil {
				statusCb("Cannot check emptiness: " + thisPath)
				return nil
			}
			if !iED {
				// Not empty.
				return nil
			}

			iO, err := isOpaqueDirectory(thisPath)
			if err != nil {
				statusCb("Cannot check opaque-ness: " + thisPath)
				return nil
			}
			if iO {
				// Opaque.
				return nil
			}

			if path == "" {
				// We don't want to remove the root directory.
				return nil
			}

			statusCb("Removing empty non-opaque directory: " + path)
			os.Remove(thisPath)
			removedDirs[thisPath] = true
			removedAnything = true
			return nil
		})
	}

	return nil
}

func isEmptyDirectory(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

func isOpaqueDirectory(name string) (bool, error) {
	opaque, err := xattr.Get(name, "trusted.overlay.opaque")
	if err == nil && len(opaque) == 1 && opaque[0] == 'y' {
		return true, nil
	}
	if xattr.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
