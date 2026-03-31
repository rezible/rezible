//go:build ignore

//go:generate go run -mod=mod entc.go
// old go:generate go run -mod=mod entgo.io/ent/cmd/ent generate --feature schema/snapshot,sql/versioned-migration,entql,intercept,sql/upsert,sql/modifier,privacy --template ./debug.go.tmpl ./schema

package main

import (
	"log"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	
	"github.com/rezible/rezible/ent/schema"
)

func main() {
	cfg := &gen.Config{
		Features: []gen.Feature{
			gen.FeatureSnapshot,
			gen.FeatureVersionedMigration,
			gen.FeatureEntQL,
			gen.FeatureIntercept,
			gen.FeatureUpsert,
			gen.FeatureModifier,
			gen.FeaturePrivacy,
		},
		SchemaConfig: gen.SchemaConfig{
			Table: schema.Name,
		},
		Templates: []*gen.Template{
			gen.MustParse(gen.NewTemplate("debug").ParseFiles("./debug.go.tmpl")),
		},
	}
	if genErr := entc.Generate("./schema", cfg); genErr != nil {
		log.Fatalln("running ent codegen:", genErr)
	}
}
