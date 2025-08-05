package access

import (
	"context"

	"github.com/rezible/rezible/ent"
)

type Role string

const (
	RoleSystem Role = "system"
	RoleUser   Role = "user"
)

type Context interface {
	IsSystem() bool
	TenantId() (int, bool)
}

type ViewContext struct {
	tenant *ent.Tenant
	roles  map[Role]struct{}
}

func (v ViewContext) IsSystem() bool {
	_, isSystem := v.roles[RoleSystem]
	return isSystem
}

func (v ViewContext) TenantId() (int, bool) {
	if v.tenant != nil {
		return v.tenant.ID, true
	}
	return 0, false
}

type ctxKey struct{}

func NewContext(parent context.Context, v Context) context.Context {
	return context.WithValue(parent, ctxKey{}, v)
}

func FromContext(ctx context.Context) Context {
	v, _ := ctx.Value(ctxKey{}).(Context)
	return v
}
