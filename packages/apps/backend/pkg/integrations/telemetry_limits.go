package integrations

import (
	"fmt"
	"math"
	"strings"
	"time"

	rez "github.com/rezible/rezible"
)

// Limits on log and metric reads. A request outside them is rez.ErrInvalidInput naming the limit; the
// others adjust the result, and the result's ReadRecord.Limits says so.
const (
	MaxTelemetryReadRange = 6 * time.Hour
	DefaultLogSearchLimit = 50
	MaxLogSearchLimit     = 200
	MaxLogLineLength      = 2000 // runes
	MaxTelemetryPoints    = 300  // per series; sets the minimum step
	MaxMetricSeries       = 20
	MaxMetricNames        = 200
	TelemetryReadTimeout  = 10 * time.Second
)

func (p LogSearchParams) Validate() error {
	if serviceErr := validateReadService(p.Service); serviceErr != nil {
		return serviceErr
	}
	if p.Limit < 0 {
		return fmt.Errorf("%w: limit must not be negative", rez.ErrInvalidInput)
	}
	if p.Limit > MaxLogSearchLimit {
		return fmt.Errorf("%w: limit of %d is above the maximum of %d lines", rez.ErrInvalidInput, p.Limit, MaxLogSearchLimit)
	}
	return validateReadRange(p.Start, p.End)
}

// EffectiveLimit is the number of lines the search may return.
func (p LogSearchParams) EffectiveLimit() int {
	if p.Limit == 0 {
		return DefaultLogSearchLimit
	}
	return p.Limit
}

func (p LogCountParams) Validate() error {
	if serviceErr := validateReadService(p.Service); serviceErr != nil {
		return serviceErr
	}
	if p.Step < 0 {
		return fmt.Errorf("%w: step must not be negative", rez.ErrInvalidInput)
	}
	return validateReadRange(p.Start, p.End)
}

func (p MetricNamesParams) Validate() error {
	if serviceErr := validateReadService(p.Service); serviceErr != nil {
		return serviceErr
	}
	return validateReadRange(p.Start, p.End)
}

func (p MetricRangeParams) Validate() error {
	if strings.TrimSpace(p.Query) == "" {
		return fmt.Errorf("%w: query is required", rez.ErrInvalidInput)
	}
	if p.Step < 0 {
		return fmt.Errorf("%w: step must not be negative", rez.ErrInvalidInput)
	}
	return validateReadRange(p.Start, p.End)
}

func validateReadService(service string) error {
	if strings.TrimSpace(service) == "" {
		return fmt.Errorf("%w: service is required", rez.ErrInvalidInput)
	}
	return nil
}

func validateReadRange(start, end time.Time) error {
	if !end.After(start) {
		return fmt.Errorf("%w: end must be after start", rez.ErrInvalidInput)
	}
	if end.Sub(start) > MaxTelemetryReadRange {
		return fmt.Errorf("%w: range of %s is longer than the %s limit", rez.ErrInvalidInput, end.Sub(start), MaxTelemetryReadRange)
	}
	return nil
}

// ReadStep returns the step, in whole seconds, for a range read of at most MaxTelemetryPoints points. A
// zero requested step uses the minimum. When a smaller step is raised, it also returns the limit to record.
func ReadStep(start, end time.Time, requested time.Duration) (time.Duration, string) {
	minSeconds := max(1, math.Ceil(end.Sub(start).Seconds()/MaxTelemetryPoints))
	minStep := time.Duration(minSeconds) * time.Second
	if requested == 0 {
		return minStep, ""
	}
	if requested < minStep {
		limit := fmt.Sprintf("step raised from %s to %s to stay within %d points", requested, minStep, MaxTelemetryPoints)
		return minStep, limit
	}
	return time.Duration(math.Ceil(requested.Seconds())) * time.Second, ""
}
