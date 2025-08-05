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

func MakeRoles(roles ...Role) Roles {
	r := Roles{}
	for _, role := range roles {
		r[role] = struct{}{}
	}
	return r
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

func storeAuthContext(parent context.Context, ac *AuthContext) context.Context {
	return context.WithValue(parent, ctxKey{}, ac)
}

func GetAuthContext(ctx context.Context) *AuthContext {
	c, _ := ctx.Value(ctxKey{}).(*AuthContext)
	return c
}

func GetContextTenantId(ctx context.Context) (int, bool) {
	ac := ctx.Value(ctxKey{}).(*AuthContext)
	if ac == nil {
		return -1, false
	}
	return ac.TenantId()
}

func SystemContext(ctx context.Context) context.Context {
	return storeAuthContext(ctx, &AuthContext{roles: MakeRoles(RoleSystem)})
}

func TenantContext(ctx context.Context, role Role, tenantId int) context.Context {
	c := &AuthContext{
		roles:  MakeRoles(role),
		tenant: &ent.Tenant{ID: tenantId},
	}
	return storeAuthContext(ctx, c)
}
