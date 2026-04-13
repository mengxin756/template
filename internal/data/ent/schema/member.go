package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Member holds the schema definition for the Member entity.
type Member struct{ ent.Schema }

// Fields of the Member.
func (Member) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_id").Unique().Comment("associated user ID"),
		field.String("level").Default("normal").Comment("member level: normal, silver, gold, platinum"),
		field.String("status").Default("active").Comment("member status: active, expired, suspended, cancelled"),
		field.Int("total_points").Default(0).Comment("total accumulated points"),
		field.Int("current_points").Default(0).Comment("current available points"),
		field.Time("expired_at").Default(func() time.Time { return time.Now().AddDate(1, 0, 0) }).Comment("membership expiration time"),
		field.Time("created_at").Default(time.Now).Immutable().Comment("created time"),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).Comment("updated time"),
	}
}

// Indexes of the Member.
func (Member) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id").Unique(),
		index.Fields("status"),
		index.Fields("level"),
	}
}

// Edges of the Member.
func (Member) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("member").Unique().Field("user_id"),
	}
}
