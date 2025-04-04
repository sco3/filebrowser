package users

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/afero"

	afs3 "github.com/fclairamb/afero-s3"
	"github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/filebrowser/filebrowser/v2/rules"
)

// ViewMode describes a view mode.
type ViewMode string

const (
	ListViewMode   ViewMode = "list"
	MosaicViewMode ViewMode = "mosaic"
)

// User describes a user.
type User struct {
	ID           uint          `storm:"id,increment" json:"id"`
	Username     string        `storm:"unique" json:"username"`
	Password     string        `json:"password"`
	Scope        string        `json:"scope"`
	Locale       string        `json:"locale"`
	LockPassword bool          `json:"lockPassword"`
	ViewMode     ViewMode      `json:"viewMode"`
	SingleClick  bool          `json:"singleClick"`
	Perm         Permissions   `json:"perm"`
	Commands     []string      `json:"commands"`
	Sorting      files.Sorting `json:"sorting"`
	Fs           afero.Fs      `json:"-" yaml:"-"`
	Rules        []rules.Rule  `json:"rules"`
	HideDotfiles bool          `json:"hideDotfiles"`
	DateFormat   bool          `json:"dateFormat"`
}

// GetRules implements rules.Provider.
func (u *User) GetRules() []rules.Rule {
	return u.Rules
}

var checkableFields = []string{
	"Username",
	"Password",
	"Scope",
	"ViewMode",
	"Commands",
	"Sorting",
	"Rules",
}

type s3FsRootDirHack struct {
	afero.Fs
}

type s3FileRootDirHack struct {
	afero.File

	fs *s3FsRootDirHack
}

func (m *s3FsRootDirHack) Stat(name string) (os.FileInfo, error) {

	if path.Clean(name) == "/" {
		return afs3.NewFileInfo(path.Base(name), true, 0, time.Unix(0, 0)), nil
	}

	return m.Fs.Stat(name)

}

func (m *s3FsRootDirHack) Open(name string) (afero.File, error) {

	f, err := m.Fs.Open(name)

	return &s3FileRootDirHack{File: f, fs: m}, err

}

func (m *s3FileRootDirHack) Stat() (os.FileInfo, error) {

	return m.fs.Stat(m.File.Name())

}

// Clean cleans up a user and verifies if all its fields
// are alright to be saved.
//
//nolint:gocyclo
func (u *User) Clean(baseScope string, fields ...string) error {
	if len(fields) == 0 {
		fields = checkableFields
	}

	for _, field := range fields {
		switch field {
		case "Username":
			if u.Username == "" {
				return errors.ErrEmptyUsername
			}
		case "Password":
			if u.Password == "" {
				return errors.ErrEmptyPassword
			}
		case "ViewMode":
			if u.ViewMode == "" {
				u.ViewMode = ListViewMode
			}
		case "Commands":
			if u.Commands == nil {
				u.Commands = []string{}
			}
		case "Sorting":
			if u.Sorting.By == "" {
				u.Sorting.By = "name"
			}
		case "Rules":
			if u.Rules == nil {
				u.Rules = []rules.Rule{}
			}
		}
	}

	if u.Fs == nil {
		scope := u.Scope
		log.Printf("Scope: %v\n", scope)
		if strings.HasPrefix(scope, "s3://") {

			sess, _ := session.NewSession(&aws.Config{
				Region: aws.String("eu-west-1"),
				//Credentials: credentials.NewStaticCredentials( //
				//	"...", "...", "", //
				//),
			})
			u.Fs = &s3FsRootDirHack{afs3.NewFs("s3://dz-bucket-1234", sess)}

		} else {
			scope = filepath.Join(baseScope, filepath.Join("/", scope)) //nolint:gocritic
			u.Fs = afero.NewBasePathFs(afero.NewOsFs(), scope)
		}

	}

	return nil
}

// FullPath gets the full path for a user's relative path.
func (u *User) FullPath(path string) string {
	return afero.FullBaseFsPath(u.Fs.(*afero.BasePathFs), path)
}

// CanExecute checks if an user can execute a specific command.
func (u *User) CanExecute(command string) bool {
	if !u.Perm.Execute {
		return false
	}

	for _, cmd := range u.Commands {
		if regexp.MustCompile(cmd).MatchString(command) {
			return true
		}
	}

	return false
}
