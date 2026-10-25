//go:build ignore

package main

import (
	"log"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"

	"github.com/rezible/rezible/ent/schema/extensions"
)

func main() {
	cfg := &gen.Config{
		Features: []gen.Feature{
			gen.FeatureSnapshot,
			gen.FeatureVersionedMigration,
			gen.FeatureEntQL,
			gen.FeatureIntercept,
			gen.FeatureUpsert,
			gen.FeaturePrivacy,
			gen.FeatureSchemaConfig,
			gen.FeatureLock,
		},
	}
	withExtensions := entc.Extensions(
		extensions.EnumValues(),
		extensions.TenantReferences(),
		extensions.SplitMutations(),
		extensions.KnowledgeReferences(),
	)
	if genErr := entc.Generate("./schema", cfg, withExtensions); genErr != nil {
		log.Fatalln("running ent codegen:", genErr)
	}
}
