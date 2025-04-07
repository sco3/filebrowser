package s3

import (
	"github.com/fclairamb/afero-s3"
	"github.com/spf13/afero"
	"io/fs"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

type S3FsRootDirHack struct {
	afero.Fs
}

type S3FileRootDirHack struct {
	afero.File

	fs *S3FsRootDirHack
}

func (m *S3FsRootDirHack) Stat(name string) (os.FileInfo, error) {

	if path.Clean(name) == "/" {
		return s3.NewFileInfo(path.Base(name), true, 0, time.Unix(0, 0)), nil
	}

	return m.Fs.Stat(name)

}

func (m *S3FsRootDirHack) Open(name string) (afero.File, error) {

	f, err := m.Fs.Open(name)

	return &S3FileRootDirHack{File: f, fs: m}, err

}

func (m *S3FileRootDirHack) Stat() (os.FileInfo, error) {

	return m.fs.Stat(m.File.Name())

}

func (m *S3FsRootDirHack) LstatIfPossible(name string) (os.FileInfo, bool, error) {
	var info fs.FileInfo
	var err error
	if path.Clean(name) == "/" {
		info = s3.NewFileInfo(path.Base(name), true, 0, time.Unix(0, 0))
	} else {
		if strings.HasSuffix(name, "/") {
			name = strings.TrimSuffix(name, "/")
		}
		info, err = m.Fs.Stat(name)
	}

	log.Printf("name: %v %v dir: %v", name, info.Name(), info.IsDir())

	return info, false, err // false = not a true Lstat
}

func IsS3(fs afero.Fs) bool {
	return fs.(*S3FsRootDirHack) != nil
}

func (m *S3FsRootDirHack) RemoveAll(name string) error {
	// workaround for remove all in s3
	if m.Fs.Remove(name) == nil {
		return nil
	}
	// simple remove did not work try original implementation

	err := m.Fs.RemoveAll(name)
	if err != nil {
		return err
	}
	return nil
}
