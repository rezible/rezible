package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"fmt"
	"github.com/google/uuid"
)

// MeetingSchedule holds the schema definition for the MeetingSchedule entity.
type MeetingSchedule struct {
	ent.Schema
}

var (
	weekdays = map[string]bool{"sun": true, "mon": true, "tue": true, "wed": true, "thu": true, "fri": true, "sat": true}
)

func validateWeekdaysValue(v []string) error {
	for _, f := range v {
		if _, ok := weekdays[f]; !ok {
			return fmt.Errorf(`invalid weekday: %s`, f)
		}
	}
	return nil
}

// Fields of the MeetingSchedule.
func (MeetingSchedule) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.String("name"),
		field.String("description").Optional(),
		field.Int("begin_minute"), // eg 9am = 540
		field.Int("duration_minutes"),
		field.Time("start_date"),
		field.Enum("repeats").Values("daily", "weekly", "monthly"),
		field.Int("repetition_step").Default(1),                                // every N days/weeks/months
		field.Strings("week_days").Validate(validateWeekdaysValue).Optional(),  // for weekly, which days of week
		field.Enum("monthly_on").Values("same_day", "same_weekday").Optional(), // for monthly
		field.Time("until_date").Optional(),
		field.Int("num_repetitions").Optional(),
	}
}

// Edges of the MeetingSchedule.
func (MeetingSchedule) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("sessions", MeetingSession.Type),
		edge.To("owning_team", Team.Type),
	}
}
