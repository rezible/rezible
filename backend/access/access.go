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

	noTenantId = -1
)

type Context struct {
	tenantId int
	roles    Roles
}

func newContext(tenantId int, roles Roles) *Context {
	return &Context{tenantId: tenantId, roles: roles}
}

func (v Context) HasRole(r Role) bool {
	return v.roles.Has(r)
}

func (v Context) TenantId() (int, bool) {
	if v.tenantId != noTenantId {
		return v.tenantId, true
	}
	return noTenantId, false
}

type ctxKey struct{}

func storeContext(parent context.Context, ac *Context) context.Context {
	return context.WithValue(parent, ctxKey{}, ac)
}

func AnonymousContext(ctx context.Context) context.Context {
	return storeContext(ctx, newContext(noTenantId, MakeRoles(RoleAnonymous)))
}

func SystemContext(ctx context.Context) context.Context {
	return storeContext(ctx, newContext(noTenantId, MakeRoles(RoleSystem)))
}

func TenantSystemContext(ctx context.Context, tenantId int) context.Context {
	return storeContext(ctx, newContext(tenantId, MakeRoles(RoleSystem)))
}

func UserContext(ctx context.Context, user *ent.User) context.Context {
	return storeContext(ctx, newContext(user.TenantID, MakeRoles(RoleUser)))
}

func GetContext(ctx context.Context) *Context {
	c, _ := ctx.Value(ctxKey{}).(*Context)
	return c
}

func GetContextTenantId(ctx context.Context) (int, bool) {
	ac := ctx.Value(ctxKey{}).(*Context)
	if ac == nil {
		return noTenantId, false
	}
	return ac.TenantId()
}
