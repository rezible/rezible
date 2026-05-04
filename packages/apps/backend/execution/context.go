package execution

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type (
	Context struct {
		Actor      Actor            `json:"actor"`
		Provenance Provenance       `json:"provenance"`
		Auth       *rez.AuthSession `json:"auth,omitempty"`
	}

	ActorKind string
	Actor     struct {
		Kind     ActorKind  `json:"kind"`
		TenantID *int       `json:"tenant_id,omitempty"`
		UserID   *uuid.UUID `json:"user_id,omitempty"`
	}

	SourceKind string

	Provenance struct {
		Source        SourceKind `json:"source"`
		RequestID     string     `json:"request_id,omitempty"`
		CorrelationID string     `json:"correlation_id,omitempty"`
		ParentKind    string     `json:"parent_kind,omitempty"`
		ParentID      string     `json:"parent_id,omitempty"`
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

type ctxKey struct{}

func StoreInContext(ctx context.Context, exec Context) context.Context {
	if err := exec.validate(); err != nil {
		panic(err)
	}
	return context.WithValue(ctx, ctxKey{}, exec)
}

func ContextExists(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKey{}).(Context)
	return ok
}

func FromContext(ctx context.Context) Context {
	if exec, ok := ctx.Value(ctxKey{}).(Context); ok {
		return exec
	}
	return Context{
		Actor: Actor{Kind: KindAnonymous},
		Provenance: Provenance{
			Source: SourceInternal,
		},
	}
}

func GetActor(ctx context.Context) Actor {
	return FromContext(ctx).Actor
}

func TenantID(ctx context.Context) (int, bool) {
	exec := FromContext(ctx)
	if exec.Actor.TenantID == nil {
		return -1, false
	}
	return *exec.Actor.TenantID, true
}

func UserID(ctx context.Context) (uuid.UUID, bool) {
	exec := FromContext(ctx)
	if exec.Actor.UserID == nil {
		return uuid.Nil, false
	}
	return *exec.Actor.UserID, true
}

func AuthSession(ctx context.Context) *rez.AuthSession {
	exec := FromContext(ctx)
	return exec.Auth
}

func NewContext(ctx context.Context, kind ActorKind, source SourceKind) context.Context {
	return StoreInContext(ctx, Context{
		Actor: Actor{Kind: kind},
		Provenance: Provenance{
			Source: source,
		},
	})
}

func AnonymousContext(ctx context.Context) context.Context {
	exec := FromContext(ctx)
	exec.Actor.Kind = KindAnonymous
	return StoreInContext(ctx, exec)
}

func SystemContext(ctx context.Context) context.Context {
	exec := FromContext(ctx)
	exec.Actor.Kind = KindSystem
	return StoreInContext(ctx, exec)
}

func SystemTenantContext(ctx context.Context, tenantID int) context.Context {
	exec := FromContext(ctx)
	exec.Actor = Actor{
		Kind:     KindSystem,
		TenantID: &tenantID,
	}
	return StoreInContext(ctx, exec)
}

func AnonymousTenantContext(ctx context.Context, tenantID int) context.Context {
	exec := FromContext(ctx)
	exec.Actor = Actor{
		Kind:     KindAnonymous,
		TenantID: &tenantID,
	}
	return StoreInContext(ctx, exec)
}

func UserContext(ctx context.Context, u ent.User, sess *rez.AuthSession) context.Context {
	exec := FromContext(ctx)
	exec.Actor = Actor{
		Kind:     KindUser,
		TenantID: &u.TenantID,
		UserID:   &u.ID,
	}
	exec.Auth = sess
	return StoreInContext(ctx, exec)
}

func (c Context) validate() error {
	switch c.Actor.Kind {
	case KindAnonymous:
		if c.Auth != nil {
			return fmt.Errorf("anonymous actor cannot carry auth")
		}
		if c.Actor.UserID != nil {
			return fmt.Errorf("anonymous actor cannot carry user id")
		}
	case KindUser:
		if c.Actor.TenantID == nil {
			return fmt.Errorf("user actor missing tenant id")
		}
		if c.Actor.UserID == nil {
			return fmt.Errorf("user actor missing user id")
		}
	case KindSystem:
	default:
		return fmt.Errorf("invalid actor kind: %q", c.Actor.Kind)
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

func Decode(encoded []byte) (Context, error) {
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
