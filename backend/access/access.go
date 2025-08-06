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
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAnonymous Role = "anonymous"
)

type Context struct {
	roles  Roles
	tenant *ent.Tenant
}

func (v Context) HasRole(r Role) bool {
	return v.roles.Has(r)
}

func (v Context) TenantId() (int, bool) {
	if v.tenant != nil {
		return v.tenant.ID, true
	}
	return -1, false
}

type ctxKey struct{}

func storeContext(parent context.Context, ac *Context) context.Context {
	return context.WithValue(parent, ctxKey{}, ac)
}

func SystemContext(ctx context.Context) context.Context {
	return storeContext(ctx, &Context{roles: MakeRoles(RoleSystem)})
}

func TenantContext(ctx context.Context, role Role, tenantId int) context.Context {
	c := &Context{
		roles:  MakeRoles(role),
		tenant: &ent.Tenant{ID: tenantId},
	}
	return storeContext(ctx, c)
}

func GetContext(ctx context.Context) *Context {
	c, _ := ctx.Value(ctxKey{}).(*Context)
	return c
}

func GetContextTenantId(ctx context.Context) (int, bool) {
	ac := ctx.Value(ctxKey{}).(*Context)
	if ac == nil {
		return -1, false
	}
	return ac.TenantId()
}
