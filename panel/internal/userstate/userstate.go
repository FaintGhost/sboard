package userstate

import (
  "time"

  "sboard/panel/internal/db"
)

const (
  StatusActive          = "active"
  StatusDisabled        = "disabled"
  StatusExpired         = "expired"
  StatusTrafficExceeded = "traffic_exceeded"
)

func EffectiveStatus(user db.User, now time.Time) string {
  if user.Status == StatusDisabled {
    return StatusDisabled
  }
  if user.Status == StatusExpired {
    return StatusExpired
  }
  if user.Status == StatusTrafficExceeded {
    return StatusTrafficExceeded
  }
  if user.ExpireAt != nil && !user.ExpireAt.After(now) {
    return StatusExpired
  }
  if user.TrafficLimit > 0 && user.TrafficUsed >= user.TrafficLimit {
    return StatusTrafficExceeded
  }
  return StatusActive
}

func IsSubscriptionEligible(user db.User, now time.Time) bool {
  return EffectiveStatus(user, now) == StatusActive
}

