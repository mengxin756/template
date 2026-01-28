package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// User holds the schema definition for the User entity.
type User struct{ ent.Schema }

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().MaxLen(50).Comment("用户姓名"),
		field.String("email").Unique().MaxLen(100).Comment("用户邮箱"),
		field.String("password").NotEmpty().MaxLen(100).Comment("用户密码"),
		field.String("status").Default("active").Comment("用户状态: active, inactive, banned"),
		field.Time("created_at").Default(time.Now).Immutable().Comment("创建时间"),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).Comment("更新时间"),
	}
}

// Indexes of the User.
func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email"),
		index.Fields("status"),
		index.Fields("created_at"),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return nil
}
