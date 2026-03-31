//go:build ignore

package main

import (
	"log"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
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
		Templates: []*gen.Template{
			gen.MustParse(gen.NewTemplate("debug").ParseFiles("./debug.go.tmpl")),
		},
	}
	if genErr := entc.Generate("./schema", cfg); genErr != nil {
		log.Fatalln("running ent codegen:", genErr)
	}
}
