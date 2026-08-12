package sqlstore

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func TestIsYugabyteRetryable(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "serialization_failure",
			err:  &pgconn.PgError{Code: "40001"},
			want: true,
		},
		{
			name: "deadlock_detected",
			err:  &pgconn.PgError{Code: "40P01"},
			want: true,
		},
		{
			name: "unrelated_pg_error",
			err:  &pgconn.PgError{Code: "23505"}, // unique_violation
			want: false,
		},
		{
			name: "wrapped_retryable_pg_error",
			err:  errors.Join(errors.New("tx failed"), &pgconn.PgError{Code: "40001"}),
			want: true,
		},
		{
			name: "non_pg_error",
			err:  errors.New("connection reset"),
			want: false,
		},
		{
			name: "nil_error",
			err:  nil,
			want: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, c.want, isYugabyteRetryable(c.err))
		})
	}
}
