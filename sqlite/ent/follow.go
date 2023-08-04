// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/Mushus/activitypub/sqlite/ent/follow"
)

// Follow is the model entity for the Follow schema.
type Follow struct {
	config `json:"-"`
	// ID of the ent.
	ID string `json:"id,omitempty"`
	// FromID holds the value of the "fromID" field.
	FromID string `json:"fromID,omitempty"`
	// ToID holds the value of the "toID" field.
	ToID         string `json:"toID,omitempty"`
	selectValues sql.SelectValues
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Follow) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case follow.FieldID, follow.FieldFromID, follow.FieldToID:
			values[i] = new(sql.NullString)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Follow fields.
func (f *Follow) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case follow.FieldID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value.Valid {
				f.ID = value.String
			}
		case follow.FieldFromID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field fromID", values[i])
			} else if value.Valid {
				f.FromID = value.String
			}
		case follow.FieldToID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field toID", values[i])
			} else if value.Valid {
				f.ToID = value.String
			}
		default:
			f.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the Follow.
// This includes values selected through modifiers, order, etc.
func (f *Follow) Value(name string) (ent.Value, error) {
	return f.selectValues.Get(name)
}

// Update returns a builder for updating this Follow.
// Note that you need to call Follow.Unwrap() before calling this method if this Follow
// was returned from a transaction, and the transaction was committed or rolled back.
func (f *Follow) Update() *FollowUpdateOne {
	return NewFollowClient(f.config).UpdateOne(f)
}

// Unwrap unwraps the Follow entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (f *Follow) Unwrap() *Follow {
	_tx, ok := f.config.driver.(*txDriver)
	if !ok {
		panic("ent: Follow is not a transactional entity")
	}
	f.config.driver = _tx.drv
	return f
}

// String implements the fmt.Stringer.
func (f *Follow) String() string {
	var builder strings.Builder
	builder.WriteString("Follow(")
	builder.WriteString(fmt.Sprintf("id=%v, ", f.ID))
	builder.WriteString("fromID=")
	builder.WriteString(f.FromID)
	builder.WriteString(", ")
	builder.WriteString("toID=")
	builder.WriteString(f.ToID)
	builder.WriteByte(')')
	return builder.String()
}

// Follows is a parsable slice of Follow.
type Follows []*Follow
