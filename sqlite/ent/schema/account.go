package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Account holds the schema definition for the Account entity.
type Account struct {
	ent.Schema
}

// Fields of the Account.
func (Account) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Immutable().Unique(),
		field.String("username").Immutable().Unique(),
		field.String("email").Unique(),
		field.String("password"),
		field.String("private_key"),
	}
}

// Edges of the Account.
func (Account) Edges() []ent.Edge {
	return nil
}
