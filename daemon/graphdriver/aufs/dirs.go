// +build linux

package aufs

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

// Return all the directories
func loadIds(root string) ([]string, error) {
	dirs, err := ioutil.ReadDir(root)
	if err != nil {
		return nil, err
	}
	out := []string{}
	for _, d := range dirs {
		if !d.IsDir() {
			out = append(out, d.Name())
		}
	}
	return out, nil
}

// Read the layers file for the current id and return all the
// layers represented by new lines in the file
//
// If there are no lines in the file then the id has no parent
// and an empty slice is returned.
func getParentIDs(root, id string) ([]string, error) {
	f, err := os.Open(path.Join(root, "layers", id))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	out := []string{}
	s := bufio.NewScanner(f)

	for s.Scan() {
		if t := s.Text(); t != "" {
			out = append(out, s.Text())
		}
	}
	return out, s.Err()
}

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must *not* exist.
// Symlinks are ignored and skipped.
func CopyDir(src, dst, labelDir string) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	_, err = os.Stat(dst)
	if err == nil {
		return
	}
	if !os.IsNotExist(err) {
		return
	}

	_, err = os.Stat(labelDir)
	if err != nil && os.IsNotExist(err) {
		fmt.Println("label dir doesn't exist, creating at: %q", labelDir)
		err = os.Mkdir(labelDir, 755)
	}
	if err != nil {
		return err
	}

	cpCmd := exec.Command("cp", "-rf", src, dst)
	fmt.Println("Copying: ", src, dst, "\n")

	err = cpCmd.Run()
	if (err != nil) {
		return err
	}

	return
}


func (a *Driver) getMountpoint(id string) string {
	return path.Join(a.mntPath(), id)
}

func (a *Driver) mntPath() string {
	return path.Join(a.rootPath(), "mnt")
}

func (a *Driver) getDiffPath(id string) string {
	return path.Join(a.diffPath(), id)
}

func (a *Driver) getLabelDiffPath(id, label string) string {
	return path.Join(a.diffPath(), label, id)
}

func (a *Driver) labelDiffPath(label string) string {
	return path.Join(a.diffPath(), label)
}

func (a *Driver) diffPath() string {
	return path.Join(a.rootPath(), "diff")
}
