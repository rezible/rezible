package errs

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCodeOfFindsOutermostError(t *testing.T) {
	inner := New(CodeNotFound, "")
	outer := Wrap(fmt.Errorf("load team: %w", inner), CodeForbidden, "")

	require.Equal(t, CodeNotFound, CodeOf(fmt.Errorf("get team: %w", inner)))
	require.Equal(t, CodeForbidden, CodeOf(fmt.Errorf("get team: %w", outer)))
	require.Equal(t, CodeInternal, CodeOf(errors.New("connection refused")))
}

func TestPublicMessage(t *testing.T) {
	require.Equal(t, "Team names must be unique.", PublicMessage(fmt.Errorf("create: %w", New(CodeConflict, "Team names must be unique."))))
	require.Equal(t, defaultMessages[CodeConflict], PublicMessage(New(CodeConflict, "")))
	require.Equal(t, defaultMessages[CodeInternal], PublicMessage(errors.New("pq: relation does not exist")))
}

func TestErrorIncludesCause(t *testing.T) {
	cause := errors.New("pq: relation does not exist")
	require.Equal(t, "not_found: pq: relation does not exist", Wrap(cause, CodeNotFound, "").Error())
	require.Equal(t, "Team is archived.: pq: relation does not exist", Wrap(cause, CodeConflict, "Team is archived.").Error())
	require.Equal(t, "not_found", New(CodeNotFound, "").Error())
}

func TestWrappedSentinelMatches(t *testing.T) {
	sentinel := New(CodeNotFound, "")
	require.ErrorIs(t, fmt.Errorf("get team: %w", sentinel), sentinel)
	require.ErrorIs(t, Wrap(fmt.Errorf("get team: %w", sentinel), CodeForbidden, ""), sentinel)
	require.NotErrorIs(t, fmt.Errorf("get team: %w", New(CodeNotFound, "")), sentinel)
}
