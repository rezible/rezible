package datasync

import (
	"context"
	"fmt"
	"github.com/gosimple/slug"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/user"
)

// TODO: just do this in postgres
type slugTracker struct {
	prefixCount map[string]int
}

func newSlugTracker() *slugTracker {
	return &slugTracker{
		prefixCount: make(map[string]int),
	}
}

func (s *slugTracker) reset() {
	s.prefixCount = make(map[string]int)
}

func (s *slugTracker) generateUnique(title string, initialCountFn func(string) (int, error)) (string, error) {
	base := slug.MakeLang(title, "en")

	numExisting, ok := s.prefixCount[base]
	if !ok || numExisting == 0 {
		var countErr error
		numExisting, countErr = initialCountFn(base)
		if countErr != nil {
			return "", countErr
		}
	}

	count := numExisting + 1
	s.prefixCount[base] = count

	if count > 1 {
		return fmt.Sprintf("%s-%d", base, count), nil
	}
	return base, nil
}

// TODO: userTracker ?
func lookupProviderUser(ctx context.Context, dbc *ent.Client, u *ent.User) (*ent.User, error) {
	// TODO: cache?
	return dbc.User.Query().Where(user.Email(u.Email)).First(ctx)
}
