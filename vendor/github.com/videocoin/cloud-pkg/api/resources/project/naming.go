package project

import (
	"regexp"
	"strings"

	cstr "github.com/videocoin/cloud-pkg/strings"

	"github.com/videocoin/cloud-pkg/api/resources"
)

const (
	// CollectionID is the identifier of the resource that contains a list of
	// projects.
	CollectionID = "projects"
)

var (
	// ErrInvalidName indicates that the project name is invalid.
	ErrInvalidName = resources.PatternError(NamePattern.String())
	// ErrInvalidID indicates that the project identifier is invalid.
	ErrInvalidID = resources.PatternError(IDPattern.String())
)

var (
	// IDPattern is the project identifier pattern.
	IDPattern = regexp.MustCompile(`^[a-z0-9][-a-z0-9]{3,48}[a-z0-9]$`)
	// NamePattern is the project name pattern.
	NamePattern = regexp.MustCompile(`^projects/[a-z0-9][-a-z0-9]{3,48}[a-z0-9]$`)
)

// Name is the project's name.
type Name string

// ID returns the project's email.
func (n Name) ID() string {
	return strings.SplitN(string(n), resources.NameSeparator, 2)[1]
}

// ParseName parses a project name.
func ParseName(name string) (Name, error) {
	if ok := isValidName(name); !ok {
		return "", ErrInvalidName
	}
	return Name(name), nil
}

// NewName returns a project name given a project identifier.
func NewName(ID string) Name {
	return Name(cstr.JoinWithSeparator(resources.NameSeparator, CollectionID, ID))
}

// NewNameWildcard a project name with a wildcard for the project identifier.
func NewNameWildcard() Name {
	return Name(cstr.JoinWithSeparator(resources.NameSeparator, CollectionID, resources.Wildcard))
}

// IsValidID reports whether a project identifier is valid.
func IsValidID(ID string) bool {
	return IDPattern.MatchString(ID)
}

// isValidName reports whether a project name is valid.
func isValidName(name string) bool {
	return NamePattern.MatchString(name)
}
