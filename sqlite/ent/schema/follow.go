package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/Mushus/activitypub"
)

// Follow holds the schema definition for the Follow entity.
type Follow struct {
	ent.Schema
}

// Fields of the Follow.
func (Follow) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Immutable().Unique(),
		field.String("fromID").Immutable(),
		field.String("toID").Immutable(),
		field.Int("status").Default(activitypub.FollowStatusFollowing.Value()),
	}
}

// Edges of the Follow.
func (Follow) Edges() []ent.Edge {
	return nil
}

// Indexes of the Street.
func (Follow) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("fromID", "toID").Unique(),
	}
}
