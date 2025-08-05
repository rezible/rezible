package access

import (
	"context"
	"github.com/rezible/rezible/ent"
)

type (
	Role  string
	Roles map[Role]struct{}
)

func (r Roles) Has(role Role) bool {
	_, has := r[role]
	return has
}

const (
	RoleSystem Role = "system"
	RoleUser   Role = "user"
)

type AuthContext struct {
	roles  Roles
	tenant *ent.Tenant
}

func (v AuthContext) HasRole(r Role) bool {
	return v.roles.Has(r)
}

func (v AuthContext) TenantId() (int, bool) {
	if v.tenant != nil {
		return v.tenant.ID, true
	}
	return -1, false
}

type ctxKey struct{}

func StoreAuthContext(parent context.Context, ac *AuthContext) context.Context {
	return context.WithValue(parent, ctxKey{}, ac)
}

func GetAuthContext(ctx context.Context) *AuthContext {
	c, _ := ctx.Value(ctxKey{}).(*AuthContext)
	return c
}
