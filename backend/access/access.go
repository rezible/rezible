package access

import (
	"context"
)

type Context struct {
	System   bool `json:"sys"`
	TenantId *int `json:"tid"`
}

func (c Context) GetTenantId() int {
	if c.TenantId == nil {
		return -1
	}
	return *c.TenantId
}

func (c Context) HasTenant() bool {
	return c.TenantId != nil
}

func (c Context) IsSystem() bool {
	return c.System
}

func (c Context) IsAnonymous() bool {
	return !c.HasTenant() && !c.IsSystem()
}

func newContext(tenantId *int, system bool) Context {
	return Context{TenantId: tenantId, System: system}
}

type ctxKey struct{}

func SetContext(parent context.Context, ac Context) context.Context {
	return context.WithValue(parent, ctxKey{}, ac)
}

func GetContext(ctx context.Context) Context {
	c, ok := ctx.Value(ctxKey{}).(Context)
	if !ok {
		return newContext(nil, false)
	}
	return c
}

func AnonymousContext(ctx context.Context) context.Context {
	return SetContext(ctx, newContext(nil, false))
}

func TenantContext(ctx context.Context, tenantId int) context.Context {
	return SetContext(ctx, newContext(&tenantId, false))
}

func SystemContext(ctx context.Context) context.Context {
	return SetContext(ctx, newContext(nil, true))
}
