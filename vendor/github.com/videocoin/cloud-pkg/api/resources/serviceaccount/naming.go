package serviceaccount

import (
	"regexp"
	"strings"

	cstr "github.com/videocoin/cloud-pkg/strings"

	"github.com/videocoin/cloud-pkg/api/resources"
	"github.com/videocoin/cloud-pkg/api/resources/project"
)

const (
	// CollectionID is the identifier of the resource that contains a list of
	// service accounts.
	CollectionID = "serviceAccounts"

	// Domain represents the service account email domain.
	Domain = "vserviceaccount.com"
)

var (
	// ErrInvalidID indicates that the service account id is invalid.
	ErrInvalidID = resources.PatternError(IDPattern.String())
	// ErrInvalidEmail indicates that the service account email is invalid.
	ErrInvalidEmail = resources.PatternError(EmailPattern.String())
	// ErrInvalidName indicates that the service account name is invalid.
	ErrInvalidName = resources.PatternError(NamePattern.String())
)

var (
	// IDPattern is the service account identifier pattern.
	IDPattern = regexp.MustCompile(`^[a-z][-a-z0-9]{4,28}[a-z0-9]$`)
	// EmailPattern is the service account email pattern.
	EmailPattern = regexp.MustCompile(`^[a-z][-a-z0-9]{4,28}[a-z0-9]@[a-z0-9][-a-z0-9]{3,48}[a-z0-9]\.vserviceaccount\.com$`)
	// NamePattern is the service account name pattern.
	NamePattern = regexp.MustCompile(`^projects/(([a-z0-9][-a-z0-9]{3,48}[a-z0-9])|\-)/serviceAccounts/[a-z][-a-z0-9]{4,28}[a-z0-9]@[a-z0-9][-a-z0-9]{3,48}[a-z0-9]\.vserviceaccount\.com$`)
)

// Name is the service account's name.
type Name string

// Email returns the service account's email.
func (n Name) Email() string {
	return strings.SplitN(string(n), resources.NameSeparator, 4)[3]
}

// ParseName parses a service account name.
func ParseName(name string) (Name, error) {
	if ok := isValidName(name); !ok {
		return "", ErrInvalidName
	}
	return Name(name), nil
}

// NewName returns the service account's name given a project identifier and a
// service account identifier.
func NewName(projID string, accEmail string) Name {
	return Name(cstr.JoinWithSeparator(resources.NameSeparator, string(project.NewName(projID)), CollectionID, accEmail))
}

// NewNameWildcard returns the service account's name with a wildcard for the
// project id. Requests using `-` as a wildcard for the project identifier
// will infer the project identifier from the account email.
func NewNameWildcard(accEmail string) Name {
	return Name(cstr.JoinWithSeparator(resources.NameSeparator, string(project.NewNameWildcard()), CollectionID, accEmail))
}

// NewEmail returns the service account's email given a project identifier and a
// service account identifier.
func NewEmail(projID string, accID string) string {
	return cstr.Join(accID, "@", projID, ".", Domain)
}

// IsValidID reports whether a service account identifier is valid.
func IsValidID(ID string) bool {
	return IDPattern.MatchString(ID)
}

// IsValidEmail reports whether a service account email is valid.
func IsValidEmail(name string) bool {
	return EmailPattern.MatchString(name)
}

// isValidName reports whether a service account name is valid.
func isValidName(name string) bool {
	return NamePattern.MatchString(name)
}
