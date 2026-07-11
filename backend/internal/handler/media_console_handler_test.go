package handler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseMediaTaskTimeRange(t *testing.T) {
	t.Run("empty range", func(t *testing.T) {
		from, to, err := parseMediaTaskTimeRange("", "")
		require.NoError(t, err)
		require.Nil(t, from)
		require.Nil(t, to)
	})

	t.Run("valid inclusive range", func(t *testing.T) {
		from, to, err := parseMediaTaskTimeRange(
			"2026-07-01T00:00:00Z",
			"2026-07-31T23:59:59Z",
		)
		require.NoError(t, err)
		require.Equal(t, time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC), *from)
		require.Equal(t, time.Date(2026, 7, 31, 23, 59, 59, 0, time.UTC), *to)
	})

	t.Run("invalid from", func(t *testing.T) {
		_, _, err := parseMediaTaskTimeRange("2026-07-01", "")
		require.EqualError(t, err, "invalid created_from: use RFC3339 format")
	})

	t.Run("invalid to", func(t *testing.T) {
		_, _, err := parseMediaTaskTimeRange("", "tomorrow")
		require.EqualError(t, err, "invalid created_to: use RFC3339 format")
	})

	t.Run("inverted range", func(t *testing.T) {
		_, _, err := parseMediaTaskTimeRange(
			"2026-07-31T23:59:59Z",
			"2026-07-01T00:00:00Z",
		)
		require.EqualError(t, err, "created_from must not be after created_to")
	})
}
