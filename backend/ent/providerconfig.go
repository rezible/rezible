// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/providerconfig"
)

// ProviderConfig is the model entity for the ProviderConfig schema.
type ProviderConfig struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// ProviderType holds the value of the "provider_type" field.
	ProviderType providerconfig.ProviderType `json:"provider_type,omitempty"`
	// ProviderName holds the value of the "provider_name" field.
	ProviderName string `json:"provider_name,omitempty"`
	// ProviderConfig holds the value of the "provider_config" field.
	ProviderConfig []byte `json:"provider_config,omitempty"`
	// Enabled holds the value of the "enabled" field.
	Enabled bool `json:"enabled,omitempty"`
	// UpdatedAt holds the value of the "updated_at" field.
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
	selectValues sql.SelectValues
}

// scanValues returns the types for scanning values from sql.Rows.
func (*ProviderConfig) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case providerconfig.FieldProviderConfig:
			values[i] = new([]byte)
		case providerconfig.FieldEnabled:
			values[i] = new(sql.NullBool)
		case providerconfig.FieldProviderType, providerconfig.FieldProviderName:
			values[i] = new(sql.NullString)
		case providerconfig.FieldUpdatedAt:
			values[i] = new(sql.NullTime)
		case providerconfig.FieldID:
			values[i] = new(uuid.UUID)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the ProviderConfig fields.
func (pc *ProviderConfig) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case providerconfig.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				pc.ID = *value
			}
		case providerconfig.FieldProviderType:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field provider_type", values[i])
			} else if value.Valid {
				pc.ProviderType = providerconfig.ProviderType(value.String)
			}
		case providerconfig.FieldProviderName:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field provider_name", values[i])
			} else if value.Valid {
				pc.ProviderName = value.String
			}
		case providerconfig.FieldProviderConfig:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field provider_config", values[i])
			} else if value != nil {
				pc.ProviderConfig = *value
			}
		case providerconfig.FieldEnabled:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field enabled", values[i])
			} else if value.Valid {
				pc.Enabled = value.Bool
			}
		case providerconfig.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				pc.UpdatedAt = value.Time
			}
		default:
			pc.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the ProviderConfig.
// This includes values selected through modifiers, order, etc.
func (pc *ProviderConfig) Value(name string) (ent.Value, error) {
	return pc.selectValues.Get(name)
}

// Update returns a builder for updating this ProviderConfig.
// Note that you need to call ProviderConfig.Unwrap() before calling this method if this ProviderConfig
// was returned from a transaction, and the transaction was committed or rolled back.
func (pc *ProviderConfig) Update() *ProviderConfigUpdateOne {
	return NewProviderConfigClient(pc.config).UpdateOne(pc)
}

// Unwrap unwraps the ProviderConfig entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (pc *ProviderConfig) Unwrap() *ProviderConfig {
	_tx, ok := pc.config.driver.(*txDriver)
	if !ok {
		panic("ent: ProviderConfig is not a transactional entity")
	}
	pc.config.driver = _tx.drv
	return pc
}

// String implements the fmt.Stringer.
func (pc *ProviderConfig) String() string {
	var builder strings.Builder
	builder.WriteString("ProviderConfig(")
	builder.WriteString(fmt.Sprintf("id=%v, ", pc.ID))
	builder.WriteString("provider_type=")
	builder.WriteString(fmt.Sprintf("%v", pc.ProviderType))
	builder.WriteString(", ")
	builder.WriteString("provider_name=")
	builder.WriteString(pc.ProviderName)
	builder.WriteString(", ")
	builder.WriteString("provider_config=")
	builder.WriteString(fmt.Sprintf("%v", pc.ProviderConfig))
	builder.WriteString(", ")
	builder.WriteString("enabled=")
	builder.WriteString(fmt.Sprintf("%v", pc.Enabled))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(pc.UpdatedAt.Format(time.ANSIC))
	builder.WriteByte(')')
	return builder.String()
}

// ProviderConfigs is a parsable slice of ProviderConfig.
type ProviderConfigs []*ProviderConfig
