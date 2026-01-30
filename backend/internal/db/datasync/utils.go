package datasync

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
	
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/user"
)

func makeSyncContext(ctx context.Context, ignoreHistory bool, createDefaults bool) context.Context {
	if ignoreHistory {
		ctx = context.WithValue(ctx, ignoreHistoryKey{}, true)
	}
	if createDefaults {
		ctx = context.WithValue(ctx, createDefaultsKey{}, true)
	}
	return access.SystemContext(ctx)
}

type ignoreHistoryKey struct{}

func isHardSync(ctx context.Context) bool {
	_, ok := ctx.Value(ignoreHistoryKey{}).(bool)
	return ok
}

type createDefaultsKey struct{}

func shouldCreateDefaults(ctx context.Context) bool {
	_, ok := ctx.Value(createDefaultsKey{}).(bool)
	return ok
}

type initialSlugCountFn = func(context.Context, string) (int, error)

// TODO: just do this in postgres
type slugTracker struct {
	prefixCount    map[string]int
	initialCountFn initialSlugCountFn
}

func newSlugTracker(initialCountFn initialSlugCountFn) *slugTracker {
	return &slugTracker{
		prefixCount:    make(map[string]int),
		initialCountFn: initialCountFn,
	}
}

func (s *slugTracker) generateUnique(ctx context.Context, title string) (string, error) {
	base := slug.MakeLang(title, "en")

	numExisting, ok := s.prefixCount[base]
	if !ok || numExisting == 0 {
		var countErr error
		numExisting, countErr = s.initialCountFn(ctx, base)
		if countErr != nil {
			return "", countErr
		}
	}

	count := numExisting + 1
	s.prefixCount[base] = count

	newSlug := base
	if count > 1 {
		newSlug = fmt.Sprintf("%s-%d", base, count)
	}
	return newSlug, nil
}

type providerUserTracker struct {
	users   *ent.UserClient
	created map[string]uuid.UUID
}

func newProviderUserTracker(users *ent.UserClient) *providerUserTracker {
	return &providerUserTracker{
		users:   users,
		created: make(map[string]uuid.UUID),
	}
}

func (ut *providerUserTracker) lookupOrCreate(ctx context.Context, provUser *ent.User, provMapping *ent.User) (uuid.UUID, *ent.UserMutation, error) {
	// TODO: this should use UserService.LookupProviderUser

	email := provUser.Email
	if email != "" {
		if createdId, ok := ut.created[email]; ok {
			return createdId, nil, nil
		}
		dbId, lookupErr := ut.users.Query().Where(user.Email(email)).OnlyID(ctx)
		if lookupErr == nil {
			ut.created[email] = dbId
			return dbId, nil, nil
		} else if ent.IsNotFound(lookupErr) {
			newId := uuid.New()
			mut := ut.users.Create().SetID(newId).SetEmail(email).Mutation()
			ut.created[email] = newId
			return newId, mut, nil
		}
	}

	return uuid.Nil, nil, fmt.Errorf("failed to match or create user")
}
