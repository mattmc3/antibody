package pathstyle

import (
	"net/url"
	"path/filepath"
	"strings"
)

// PathStyle defines the interface for converting between URLs and filesystem paths
type PathStyle interface {
	// FromURL converts a git URL to a filesystem path
	FromURL(u *url.URL) string
	// ToURL converts a filesystem path back to a git URL
	ToURL(path string) string
	// Segments returns the number of path segments used for this style
	Segments() int
}

// EscapedStyle uses character escaping for paths (default antibody behavior)
type EscapedStyle struct{}

// ShortStyle uses owner/repo format
type ShortStyle struct {
	Domain string
}

// FullStyle uses full github.com/owner/repo format
type FullStyle struct{}

// New creates a PathStyle based on the style name and git domain.
func New(style, domain string) PathStyle {
	switch strings.ToLower(style) {
	case "short":
		return &ShortStyle{Domain: domain}
	case "full":
		return &FullStyle{}
	default:
		return &EscapedStyle{}
	}
}

// extractOwnerRepo gets the owner and repo from a git URL
func extractOwnerRepo(u *url.URL) (string, string) {
	path := strings.TrimPrefix(u.Path, "/")
	path = strings.TrimSuffix(path, ".git")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", path
}

// EscapedStyle implementation
func (e *EscapedStyle) FromURL(u *url.URL) string {
	result := u.String()
	result = strings.ReplaceAll(result, ":", "-COLON-")
	result = strings.ReplaceAll(result, "/", "-SLASH-")
	result = strings.ReplaceAll(result, "@", "-AT-")
	return result
}

func (e *EscapedStyle) ToURL(path string) string {
	result := path
	result = strings.ReplaceAll(result, "-AT-", "@")
	result = strings.ReplaceAll(result, "-SLASH-", "/")
	result = strings.ReplaceAll(result, "-COLON-", ":")
	return result
}

// EscapedStyle uses a single segment (the whole string is the dirname)
func (e *EscapedStyle) Segments() int {
	return 1
}

// ShortStyle implementation
func (s *ShortStyle) FromURL(u *url.URL) string {
	owner, repo := extractOwnerRepo(u)
	if owner == "" {
		return repo
	}
	return filepath.Join(owner, repo)
}

func (s *ShortStyle) ToURL(path string) string {
	return "https://" + s.Domain + "/" + filepath.ToSlash(path)
}

// ShortStyle uses two segments: owner/repo
func (s *ShortStyle) Segments() int {
	return 2
}

// FullStyle implementation
func (f *FullStyle) FromURL(u *url.URL) string {
	owner, repo := extractOwnerRepo(u)
	if owner == "" {
		return filepath.Join(u.Host, repo)
	}
	return filepath.Join(u.Host, owner, repo)
}

func (f *FullStyle) ToURL(path string) string {
	return "https://" + filepath.ToSlash(path)
}

// FullStyle uses three segments: domain/owner/repo
func (f *FullStyle) Segments() int {
	return 3
}
