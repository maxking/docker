// +build linux

package aufs

import (
	"bufio"
	"fmt"
	"io"
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

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
// func copyFileContents(src, dst string) (err error) {
// 	// If the target path exists, do nothing and just return
// 	if _, err := os.Stat(dst); err == nil {
// 		return nil
// 	}

// 	// If the target path does not exist, just copy the file contents
// 	// from the source to the destination.
//     in, err := os.Open(src)
//     if err != nil {
//         return err
//     }
//     defer in.Close()
//     out, err := os.Create(dst)
//     if err != nil {
//         return err
//     }
//     defer func() {
//         cerr := out.Close()
//         if err == nil {
//             err = cerr
//         }
//     }()
//     if _, err = io.Copy(out, in); err != nil {
//         return err
//     }
//     err = out.Sync()
//     return nil
// }


// The code below has been shamelessly copied from:
// https://gist.github.com/m4ng0squ4sh/92462b38df26839a3ca324697c8cba04


// CopyFile copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file. The file mode will be copied from the source and
// the copied data is synced/flushed to stable storage.
func CopyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	return
}

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must *not* exist.
// Symlinks are ignored and skipped.
func CopyDir(src string, dst string) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	_, err = os.Stat(dst)
	if err == nil {
		return
	}
	if !os.IsNotExist(err) {
		return
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

func (a *Driver) diffPath() string {
	return path.Join(a.rootPath(), "diff")
}
