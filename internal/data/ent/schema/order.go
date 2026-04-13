package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Order holds the schema definition for the Order entity.
type Order struct{ ent.Schema }

// Fields of the Order.
func (Order) Fields() []ent.Field {
	return []ent.Field{
		field.String("order_no").Unique().Comment("order number"),
		field.Int("user_id").Comment("user ID"),
		field.String("status").Default("pending").Comment("order status: pending, paid, processing, shipped, completed, cancelled, refunded"),
		field.Float("total_amount").Comment("total amount"),
		field.Float("discount_amount").Default(0).Comment("discount amount"),
		field.Float("final_amount").Comment("final amount after discount"),
		field.String("payment_method").Optional().Comment("payment method: wallet, points, mixed, external"),
		field.Float("wallet_amount").Default(0).Comment("amount paid from wallet"),
		field.Float("points_amount").Default(0).Comment("points value used"),
		field.Float("external_amount").Default(0).Comment("external payment amount"),
		field.Time("paid_at").Optional().Comment("payment time"),
		field.Float("refund_amount").Default(0).Comment("refund amount"),
		field.String("refund_reason").Optional().Comment("refund reason"),
		field.Time("refunded_at").Optional().Comment("refund time"),
		field.String("remark").Optional().Comment("order remark"),
		field.Time("created_at").Default(time.Now).Immutable().Comment("created time"),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).Comment("updated time"),
	}
}

// Indexes of the Order.
func (Order) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_no").Unique(),
		index.Fields("user_id"),
		index.Fields("status"),
		index.Fields("created_at"),
	}
}

// Edges of the Order.
func (Order) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("orders").Unique().Field("user_id"),
		edge.To("items", OrderItem.Type),
	}
}

// OrderItem holds the schema definition for the OrderItem entity.
type OrderItem struct{ ent.Schema }

// Fields of the OrderItem.
func (OrderItem) Fields() []ent.Field {
	return []ent.Field{
		field.Int("order_id").Comment("order ID"),
		field.Int("product_id").Comment("product ID"),
		field.String("product_name").Comment("product name"),
		field.Int("quantity").Comment("quantity"),
		field.Float("unit_price").Comment("unit price"),
		field.Float("discount").Default(0).Comment("discount rate"),
		field.Float("subtotal").Comment("subtotal"),
	}
}

// Indexes of the OrderItem.
func (OrderItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_id"),
		index.Fields("product_id"),
	}
}

// Edges of the OrderItem.
func (OrderItem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("order", Order.Type).Ref("items").Unique().Field("order_id"),
	}
}
