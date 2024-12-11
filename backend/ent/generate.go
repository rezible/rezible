package ent

//go:generate go run -mod=mod entgo.io/ent/cmd/ent generate --feature intercept,sql/upsert,sql/modifier --template ./debug.go.tmpl ./schema
