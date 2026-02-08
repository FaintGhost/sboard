package userstate

import (
  "testing"
  "time"

  "sboard/panel/internal/db"
  "github.com/stretchr/testify/require"
)

func TestEffectiveStatus(t *testing.T) {
  now := time.Date(2026, 2, 8, 12, 0, 0, 0, time.UTC)

  cases := []struct {
    name     string
    user     db.User
    expected string
  }{
    {
      name: "disabled takes priority",
      user: db.User{Status: StatusDisabled},
      expected: StatusDisabled,
    },
    {
      name: "manual expired preserved",
      user: db.User{Status: StatusExpired},
      expected: StatusExpired,
    },
    {
      name: "manual traffic exceeded preserved",
      user: db.User{Status: StatusTrafficExceeded},
      expected: StatusTrafficExceeded,
    },
    {
      name: "expire_at reached",
      user: db.User{Status: StatusActive, ExpireAt: ptrTime(now)},
      expected: StatusExpired,
    },
    {
      name: "traffic exceeded by counters",
      user: db.User{Status: StatusActive, TrafficLimit: 100, TrafficUsed: 100},
      expected: StatusTrafficExceeded,
    },
    {
      name: "active default",
      user: db.User{Status: StatusActive},
      expected: StatusActive,
    },
  }

  for _, tc := range cases {
    t.Run(tc.name, func(t *testing.T) {
      got := EffectiveStatus(tc.user, now)
      require.Equal(t, tc.expected, got)
    })
  }
}

func TestIsSubscriptionEligible(t *testing.T) {
  now := time.Date(2026, 2, 8, 12, 0, 0, 0, time.UTC)

  active := db.User{Status: StatusActive}
  require.True(t, IsSubscriptionEligible(active, now))

  expired := db.User{Status: StatusActive, ExpireAt: ptrTime(now.Add(-time.Minute))}
  require.False(t, IsSubscriptionEligible(expired, now))

  exceeded := db.User{Status: StatusActive, TrafficLimit: 100, TrafficUsed: 100}
  require.False(t, IsSubscriptionEligible(exceeded, now))
}

func ptrTime(t time.Time) *time.Time {
  return &t
}

