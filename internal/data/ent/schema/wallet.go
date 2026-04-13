package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Wallet holds the schema definition for the Wallet entity.
type Wallet struct{ ent.Schema }

// Fields of the Wallet.
func (Wallet) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_id").Unique().Comment("associated user ID"),
		field.Float("balance").Default(0).Comment("wallet balance"),
		field.Float("frozen_amount").Default(0).Comment("frozen amount"),
		field.String("currency").Default("CNY").Comment("currency: CNY, USD, EUR"),
		field.Int("version").Default(1).Comment("optimistic lock version"),
		field.Time("created_at").Default(time.Now).Immutable().Comment("created time"),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).Comment("updated time"),
	}
}

// Indexes of the Wallet.
func (Wallet) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id").Unique(),
	}
}

// Edges of the Wallet.
func (Wallet) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("wallet").Unique().Field("user_id"),
		edge.To("transactions", Transaction.Type),
	}
}

// Transaction holds the schema definition for the Transaction entity.
type Transaction struct{ ent.Schema }

// Fields of the Transaction.
func (Transaction) Fields() []ent.Field {
	return []ent.Field{
		field.Int("wallet_id").Comment("wallet ID"),
		field.String("transaction_type").Comment("transaction type: deposit, withdraw, payment, refund, etc."),
		field.Float("amount").Comment("transaction amount"),
		field.Float("balance").Comment("balance after transaction"),
		field.String("status").Default("pending").Comment("transaction status: pending, completed, failed, cancelled"),
		field.String("reference_type").Optional().Comment("reference type: order, refund, etc."),
		field.String("reference_id").Optional().Comment("reference ID"),
		field.String("description").Optional().Comment("transaction description"),
		field.String("operated_by").Default("user").Comment("operator: user or system"),
		field.Time("operated_at").Default(time.Now).Comment("operation time"),
		field.Time("created_at").Default(time.Now).Immutable().Comment("created time"),
	}
}

// Indexes of the Transaction.
func (Transaction) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("wallet_id"),
		index.Fields("transaction_type"),
		index.Fields("reference_type", "reference_id"),
		index.Fields("created_at"),
	}
}

// Edges of the Transaction.
func (Transaction) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("wallet", Wallet.Type).Ref("transactions").Unique().Field("wallet_id"),
	}
}
