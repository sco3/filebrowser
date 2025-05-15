package users

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	afs3 "github.com/fclairamb/afero-s3"
	"github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/filebrowser/filebrowser/v2/rules"
	"github.com/filebrowser/filebrowser/v2/s3"
	"github.com/spf13/afero"
	"log"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// ViewMode describes a view mode.
type ViewMode string

const (
	ListViewMode   ViewMode = "list"
	MosaicViewMode ViewMode = "mosaic"
)

var (
	s3Region string = "us-east-1"
	once     sync.Once
)

func SetS3Region(region string) {
	once.Do(func() {
		s3Region = region
	})
}

func getS3Region() string {
	return s3Region
}

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
		log.Printf("base scope: %v\n", baseScope)
		if strings.HasPrefix(baseScope, "s3://") {

			sess, _ := session.NewSession(&aws.Config{
				Region: aws.String(getS3Region()),
				//Credentials: credentials.NewStaticCredentials( //
				//	"...", "...", "", //
				//),
			})
			bucket := baseScope[4:]
			log.Printf("Bucket: %v\n", bucket)
			u.Fs = &s3.S3FsRootDirHack{afs3.NewFs(bucket, sess)}
			log.Printf("Users Fs: %v", u.Fs)

		} else {
			scope = filepath.Join(baseScope, filepath.Join("/", scope)) //nolint:gocritic
			u.Fs = afero.NewBasePathFs(afero.NewOsFs(), scope)
		}

	}

	return nil
}

// FullPath gets the full path for a user's relative path.
func (u *User) FullPath(path string) string {

	fs, ok := u.Fs.(*afero.BasePathFs)
	if !ok {
		return path
	}
	return afero.FullBaseFsPath(fs, path)
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
