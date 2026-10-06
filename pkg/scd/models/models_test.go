package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/stretchr/testify/require"
)

func TestOVNFromTimeIsValid(t *testing.T) {
	require.True(t, NewOVNFromTime(time.Now(), uuid.New().String()).Valid())
}

func TestNewOVNFromUUIDv7Suffix(t *testing.T) {
	type cases []struct {
		name string

		now    time.Time
		oiID   dssmodels.ID
		suffix string

		ovn string
	}

	t.Run("valid", func(t *testing.T) {
		testCases := cases{{
			name:   "exact",
			now:    time.Date(2024, time.September, 10, 13, 02, 42, int(408*time.Millisecond), time.UTC),
			oiID:   "bd65d3de-f52e-419d-acfb-ad85d557de99",
			suffix: "0191dc07-76e8-7546-84f2-e739e9f44d77", // 2024-09-10T13:02:42.408Z
			ovn:    "bd65d3de-f52e-419d-acfb-ad85d557de99_0191dc07-76e8-7546-84f2-e739e9f44d77",
		}, {
			name:   "before",
			now:    time.Date(2024, time.September, 10, 13, 02, 24, 0, time.UTC),
			oiID:   "e72589d4-8c14-4d6f-bd9c-1bfb8704e332",
			suffix: "0191dc07-2f57-79fd-b021-80456ceb627f", // 2024-09-10T13:02:24.087Z
			ovn:    "e72589d4-8c14-4d6f-bd9c-1bfb8704e332_0191dc07-2f57-79fd-b021-80456ceb627f",
		}, {
			name:   "after",
			now:    time.Date(2024, time.September, 10, 13, 02, 48, 0, time.UTC),
			oiID:   "f577437f-bc6b-4826-9c6b-7831b78eabcc",
			suffix: "0191dc07-8a71-7a12-87ed-9baa6e889874", // 2024-09-10T13:02:47.409Z
			ovn:    "f577437f-bc6b-4826-9c6b-7831b78eabcc_0191dc07-8a71-7a12-87ed-9baa6e889874",
		}, {
			name:   "before - max skew",
			now:    time.Date(2024, time.September, 10, 12, 57, 25, 0, time.UTC),
			oiID:   "e72589d4-8c14-4d6f-bd9c-1bfb8704e332",
			suffix: "0191dc07-2f57-79fd-b021-80456ceb627f", // 2024-09-10T13:02:24.087Z
			ovn:    "e72589d4-8c14-4d6f-bd9c-1bfb8704e332_0191dc07-2f57-79fd-b021-80456ceb627f",
		}, {
			name:   "after - max skew",
			now:    time.Date(2024, time.September, 10, 13, 07, 47, 0, time.UTC),
			oiID:   "f577437f-bc6b-4826-9c6b-7831b78eabcc",
			suffix: "0191dc07-8a71-7a12-87ed-9baa6e889874", // 2024-09-10T13:02:47.409Z
			ovn:    "f577437f-bc6b-4826-9c6b-7831b78eabcc_0191dc07-8a71-7a12-87ed-9baa6e889874",
		}}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				ovn, err := NewOVNFromUUIDv7Suffix(testCase.now, testCase.oiID, testCase.suffix)
				require.NoError(t, err)
				require.EqualValues(t, testCase.ovn, ovn)
			})
		}
	})

	t.Run("invalid", func(t *testing.T) {
		testCases := cases{{
			name:   "before - past skew",
			now:    time.Date(2024, time.September, 10, 12, 57, 24, 0, time.UTC),
			oiID:   "e72589d4-8c14-4d6f-bd9c-1bfb8704e332",
			suffix: "0191dc07-2f57-79fd-b021-80456ceb627f", // 2024-09-10T13:02:24.087Z
		}, {
			name:   "after - past skew",
			now:    time.Date(2024, time.September, 10, 13, 07, 48, 0, time.UTC),
			oiID:   "f577437f-bc6b-4826-9c6b-7831b78eabcc",
			suffix: "0191dc07-8a71-7a12-87ed-9baa6e889874", // 2024-09-10T13:02:47.409Z
		}, {
			name:   "before - long past skew",
			now:    time.Date(2024, time.September, 10, 11, 57, 24, 0, time.UTC),
			oiID:   "e72589d4-8c14-4d6f-bd9c-1bfb8704e332",
			suffix: "0191dc07-2f57-79fd-b021-80456ceb627f", // 2024-09-10T13:02:24.087Z
		}, {
			name:   "after - long past skew",
			now:    time.Date(2024, time.September, 10, 14, 07, 48, 0, time.UTC),
			oiID:   "f577437f-bc6b-4826-9c6b-7831b78eabcc",
			suffix: "0191dc07-8a71-7a12-87ed-9baa6e889874", // 2024-09-10T13:02:47.409Z
		}, {
			name:   "uuidv4",
			now:    time.Date(2024, time.September, 10, 13, 02, 24, 0, time.UTC),
			oiID:   "e72589d4-8c14-4d6f-bd9c-1bfb8704e332",
			suffix: "44299cb9-a722-4d9c-87bc-537a5aeb2b73",
		}, {
			name:   "not uuid",
			now:    time.Date(2024, time.September, 10, 13, 02, 24, 0, time.UTC),
			oiID:   "not_a_uuid",
			suffix: "44299cb9-a722-4d9c-87bc-537a5aeb2b73",
		}}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				_, err := NewOVNFromUUIDv7Suffix(testCase.now, testCase.oiID, testCase.suffix)
				require.Error(t, err)
			})
		}
	})
}
