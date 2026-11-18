package knowledgegraph

import (
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/rezible/rezible/ent/schema/schematypes"
)

func TestRankParentClaims(t *testing.T) {
	day := func(day int) time.Time { return time.Date(2026, 10, day, 0, 0, 0, 0, time.UTC) }
	low := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	high := uuid.MustParse("f0000000-0000-0000-0000-000000000000")
	declared, observed := schematypes.KnowledgeAuthorityDeclared, schematypes.KnowledgeAuthorityObserved
	cases := []struct {
		name   string
		claims []ParentClaim
		winner uuid.UUID
	}{
		{
			name: "a declared claim beats an earlier observed one",
			claims: []ParentClaim{{RelationshipID: low, Authority: observed, FirstObservedAt: day(1)},
				{RelationshipID: high, Authority: declared, FirstObservedAt: day(3)}},
			winner: high,
		},
		{
			name: "within an authority the earliest observed claim wins",
			claims: []ParentClaim{{RelationshipID: low, Authority: declared, FirstObservedAt: day(2)},
				{RelationshipID: high, Authority: declared, FirstObservedAt: day(1)}},
			winner: high,
		},
		{
			name: "claims of one authority first observed together go to the lowest ID",
			claims: []ParentClaim{{RelationshipID: high, Authority: observed, FirstObservedAt: day(1)},
				{RelationshipID: low, Authority: observed, FirstObservedAt: day(1)}},
			winner: low,
		},
		{
			name:   "one claim wins alone",
			claims: []ParentClaim{{RelationshipID: high, Authority: observed, FirstObservedAt: day(3)}},
			winner: high,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ranked := slices.SortedFunc(slices.Values(tc.claims), RankParentClaims)
			assert.Equal(t, tc.winner, ranked[0].RelationshipID)
		})
	}
}
