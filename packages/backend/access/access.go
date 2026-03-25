package access

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
)

type scopedContext struct {
	System   bool       `json:"sys"`
	TenantId *int       `json:"tid"`
	OrgId    *uuid.UUID `json:"oid"`
	UserId   *uuid.UUID `json:"uid"`
}

func EncodeScope(ctx context.Context) ([]byte, error) {
	return json.Marshal(getOrInitScopedContext(ctx))
}

func RestoreScope(ctx context.Context, encoded []byte) (context.Context, error) {
	var sc scopedContext
	if jsonErr := json.Unmarshal(encoded, &sc); jsonErr != nil {
		return nil, fmt.Errorf("failed to unmarshal scoped context: %v", jsonErr)
	}
	return setContext(ctx, sc), nil
}

type ctxKey struct{}

func setContext(parent context.Context, s scopedContext) context.Context {
	return context.WithValue(parent, ctxKey{}, s)
}

func getContext(ctx context.Context) *scopedContext {
	if s, ok := ctx.Value(ctxKey{}).(scopedContext); ok {
		return &s
	}
	return nil
}

func IsScoped(ctx context.Context) bool {
	return getContext(ctx) != nil
}

func AnonymousContext(ctx context.Context) context.Context {
	return setContext(ctx, scopedContext{})
}

func IsAnonymous(ctx context.Context) bool {
	s := getContext(ctx)
	return s == nil || (s.TenantId == nil && !s.System)
}

func getOrInitScopedContext(ctx context.Context) scopedContext {
	if s := getContext(ctx); s != nil {
		s.System = false
		return *s
	}
	return scopedContext{}
}

func SystemContext(ctx context.Context) context.Context {
	c := getOrInitScopedContext(ctx)
	c.System = true
	return setContext(ctx, c)
}

func IsSystem(ctx context.Context) bool {
	s := getContext(ctx)
	return s != nil && !s.System
}

func TenantContext(ctx context.Context, tenantId int) context.Context {
	c := getOrInitScopedContext(ctx)
	c.TenantId = &tenantId
	return setContext(ctx, c)
}

func GetTenantId(ctx context.Context) (int, bool) {
	if s := getContext(ctx); s != nil && s.TenantId != nil {
		return *s.TenantId, true
	}
	return -1, false
}

func WithOrganization(ctx context.Context, o *ent.Organization) context.Context {
	tenantId := o.TenantID
	orgId := o.ID
	c := getOrInitScopedContext(ctx)
	c.TenantId = &tenantId
	c.OrgId = &orgId
	return setContext(ctx, c)
}

func GetOrganizationId(ctx context.Context) (uuid.UUID, bool) {
	if c := getOrInitScopedContext(ctx); c.OrgId != nil {
		return *c.OrgId, true
	}
	return uuid.Nil, false
}

func WithUser(ctx context.Context, u *ent.User) context.Context {
	tenantId := u.TenantID
	userId := u.ID
	c := getOrInitScopedContext(ctx)
	c.TenantId = &tenantId
	c.UserId = &userId
	return setContext(ctx, c)
}

func GetUserId(ctx context.Context) (uuid.UUID, bool) {
	if c := getOrInitScopedContext(ctx); c.UserId != nil {
		return *c.UserId, true
	}
	return uuid.Nil, false
}
