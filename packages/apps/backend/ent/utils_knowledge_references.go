package ent

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/ent/knowledgerelationship"
)

// ResolveKnowledgeEntityID returns the entity id names, following the redirect a merge left, if any. A missing
// entity, or one in another tenant, is a not found error.
func ResolveKnowledgeEntityID(ctx context.Context, client *Client, id uuid.UUID) (uuid.UUID, error) {
	return resolveKnowledgeID(id, func(id uuid.UUID) (*uuid.UUID, error) {
		lockEntity := client.KnowledgeEntity.Query().
			Where(knowledgeentity.ID(id)).
			Select(knowledgeentity.FieldMergedIntoID).
			ForShare(KeyShareLock)
		entity, lockErr := lockEntity.Only(IncludeMerged(ctx))
		if lockErr != nil {
			return nil, lockErr
		}
		return entity.MergedIntoID, nil
	})
}

// CanonicalKnowledgeEntityID returns the entity id names, following the redirect a merge left, if any, without
// locking it. A missing entity is returned as it is.
func CanonicalKnowledgeEntityID(ctx context.Context, client *Client, id uuid.UUID) (uuid.UUID, error) {
	return resolveKnowledgeID(id, func(id uuid.UUID) (*uuid.UUID, error) {
		queryEntity := client.KnowledgeEntity.Query().
			Where(knowledgeentity.ID(id)).
			Select(knowledgeentity.FieldMergedIntoID)
		entity, queryErr := queryEntity.Only(IncludeMerged(ctx))
		if IsNotFound(queryErr) {
			return nil, nil
		} else if queryErr != nil {
			return nil, queryErr
		}
		return entity.MergedIntoID, nil
	})
}

// ResolveKnowledgeRelationshipID returns the relationship id names, following the redirect a merge left when it
// folded it into another, if any. A missing relationship, or one in another tenant, is a not found error.
func ResolveKnowledgeRelationshipID(ctx context.Context, client *Client, id uuid.UUID) (uuid.UUID, error) {
	return resolveKnowledgeID(id, func(id uuid.UUID) (*uuid.UUID, error) {
		lockRelationship := client.KnowledgeRelationship.Query().
			Where(knowledgerelationship.ID(id)).
			Select(knowledgerelationship.FieldMergedIntoID).
			ForShare(KeyShareLock)
		relationship, lockErr := lockRelationship.Only(IncludeMerged(ctx))
		if lockErr != nil {
			return nil, lockErr
		}
		return relationship.MergedIntoID, nil
	})
}

// maxKnowledgeRedirectHops bounds resolveKnowledgeID. Merges keep redirects flat, so a longer chain is corrupt
// data, a cycle included.
const maxKnowledgeRedirectHops = 8

// resolveKnowledgeID follows redirects from id, reading each subject FOR KEY SHARE with lockMergedInto. A merge
// locks the subject it merges away FOR UPDATE, so inside a transaction the subject returned cannot be merged away
// before the caller commits.
func resolveKnowledgeID(id uuid.UUID, lockMergedInto func(uuid.UUID) (*uuid.UUID, error)) (uuid.UUID, error) {
	start := id
	for range maxKnowledgeRedirectHops {
		mergedInto, lockErr := lockMergedInto(id)
		if lockErr != nil {
			return uuid.Nil, lockErr
		}
		if mergedInto == nil {
			return id, nil
		}
		id = *mergedInto
	}
	return uuid.Nil, fmt.Errorf("redirects from %s exceed %d hops", start, maxKnowledgeRedirectHops)
}
