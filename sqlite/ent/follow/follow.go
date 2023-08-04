// Code generated by ent, DO NOT EDIT.

package follow

import (
	"entgo.io/ent/dialect/sql"
)

const (
	// Label holds the string label denoting the follow type in the database.
	Label = "follow"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldFromID holds the string denoting the fromid field in the database.
	FieldFromID = "from_id"
	// FieldToID holds the string denoting the toid field in the database.
	FieldToID = "to_id"
	// Table holds the table name of the follow in the database.
	Table = "follows"
)

// Columns holds all SQL columns for follow fields.
var Columns = []string{
	FieldID,
	FieldFromID,
	FieldToID,
}

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	return false
}

// OrderOption defines the ordering options for the Follow queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByFromID orders the results by the fromID field.
func ByFromID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldFromID, opts...).ToFunc()
}

// ByToID orders the results by the toID field.
func ByToID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldToID, opts...).ToFunc()
}