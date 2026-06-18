package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
)

type (
	ActorKind string
	Context   struct {
		ActorKind  ActorKind  `json:"actor_kind"`
		Auth       Auth       `json:"auth"`
		Provenance Provenance `json:"provenance"`
	}

	Auth struct {
		TenantID            *int       `json:"tenant_id,omitempty"`
		TokenID             *uuid.UUID `json:"token_id,omitempty"`
		UserID              *uuid.UUID `json:"user_id,omitempty"`
		ImpersonatingUserID *uuid.UUID `json:"impersonating_user_id,omitempty"`
		ExpiresAt           time.Time  `json:"exp"`
	}

	SourceKind string
	Provenance struct {
		ID       string     `json:"id,omitempty"`
		Source   SourceKind `json:"source"`
		ParentID *string    `json:"parent_id,omitempty"`
	}
)

const (
	KindAnonymous ActorKind = "anonymous"
	KindUser      ActorKind = "user"
	KindSystem    ActorKind = "system"

	SourceHTTP     SourceKind = "http"
	SourceJob      SourceKind = "job"
	SourceCLI      SourceKind = "cli"
	SourceInternal SourceKind = "internal"
)

func (c Context) IsAnonymous() bool {
	return c.ActorKind == KindAnonymous
}

func (c Context) IsSystem() bool {
	return c.ActorKind == KindSystem
}

func (c Context) IsUser() bool {
	return c.ActorKind == KindUser
}

func (c Context) TenantID() (int, bool) {
	if c.Auth.TenantID == nil {
		return -1, false
	}
	return *c.Auth.TenantID, true
}

func (c Context) UserID() (uuid.UUID, bool) {
	if c.Auth.UserID == nil {
		return uuid.Nil, false
	}
	return *c.Auth.UserID, true
}

type ctxKey struct{}

func ContextExists(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKey{}).(Context)
	return ok
}

func SetContext(ctx context.Context, exec Context) context.Context {
	if err := exec.validate(); err != nil {
		panic(err)
	}
	return context.WithValue(ctx, ctxKey{}, exec)
}

func GetContext(ctx context.Context) Context {
	if exec, ok := getContext(ctx); ok {
		return exec
	}
	return newRootContext(KindAnonymous, SourceInternal)
}

func getContext(ctx context.Context) (Context, bool) {
	exec, ok := ctx.Value(ctxKey{}).(Context)
	return exec, ok
}

func newRootContext(kind ActorKind, source SourceKind) Context {
	return Context{
		ActorKind: kind,
		Provenance: Provenance{
			Source: source,
		},
	}
}

func NewRootContext(ctx context.Context, kind ActorKind, source SourceKind) context.Context {
	c := newInternalContext(ctx, kind)
	c.Provenance.Source = source
	return SetContext(ctx, c)
}

func newInternalContext(ctx context.Context, kind ActorKind) Context {
	c := Context{
		ActorKind: kind,
		Provenance: Provenance{
			ID:     uuid.New().String(),
			Source: SourceInternal,
		},
	}
	if parent, ok := getContext(ctx); ok {
		c.Provenance.Source = parent.Provenance.Source
		c.Provenance.ParentID = new(parent.Provenance.ID)
	}
	return c
}

func NewSystemContext(ctx context.Context) context.Context {
	c := newInternalContext(ctx, KindSystem)
	return SetContext(ctx, c)
}

func NewTenantContext(ctx context.Context, tenantID int) context.Context {
	c := GetContext(ctx)
	c.ActorKind = KindSystem
	c.Auth = Auth{
		TenantID: &tenantID,
	}
	return SetContext(ctx, c)
}

func NewUserContext(ctx context.Context, sess *ent.UserAuthSession) context.Context {
	c := GetContext(ctx)
	c.ActorKind = KindUser
	c.Auth = Auth{
		TenantID:  &sess.TenantID,
		UserID:    &sess.UserID,
		ExpiresAt: sess.ExpiresAt,
	}
	return SetContext(ctx, c)
}

func (c Context) validate() error {
	switch c.ActorKind {
	case KindAnonymous:
		if c.Auth.UserID != nil {
			return fmt.Errorf("anonymous actor cannot carry user id")
		}
	case KindUser:
		if c.Auth.TenantID == nil {
			return fmt.Errorf("user actor missing tenant id")
		}
		if c.Auth.UserID == nil {
			return fmt.Errorf("user actor missing user id")
		}
	case KindSystem:
	default:
		return fmt.Errorf("invalid actor kind: %q", c.ActorKind)
	}
	return nil
}

type encodedContext struct {
	Version int     `json:"v"`
	Context Context `json:"ctx"`
}

const encodingVersion = 1

func (c Context) Encode() ([]byte, error) {
	return json.Marshal(encodedContext{
		Version: encodingVersion,
		Context: c,
	})
}

func DecodeContext(encoded []byte) (Context, error) {
	var payload encodedContext
	if err := json.Unmarshal(encoded, &payload); err != nil {
		return Context{}, fmt.Errorf("unmarshal execution context: %w", err)
	}
	if payload.Version != encodingVersion {
		return Context{}, fmt.Errorf("unsupported execution context version: %d", payload.Version)
	}
	if err := payload.Context.validate(); err != nil {
		return Context{}, err
	}
	return payload.Context, nil
}
