package ent

//go:generate go run -mod=mod entgo.io/ent/cmd/ent generate --feature schema/snapshot,sql/versioned-migration,entql,intercept,sql/upsert,sql/modifier,privacy --template ./debug.go.tmpl ./schema
