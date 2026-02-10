package api

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"sboard/panel/internal/db"
)

const (
	defaultSystemTimezone = "UTC"
	systemTimezoneKey     = "timezone"
)

var (
	timezoneMu       sync.RWMutex
	timezoneName     = defaultSystemTimezone
	timezoneLocation = time.UTC
)

func InitSystemTimezone(ctx context.Context, store *db.Store) error {
	if store == nil {
		_, err := setSystemTimezone(defaultSystemTimezone)
		return err
	}

	value, err := store.GetSystemSetting(ctx, systemTimezoneKey)
	if err != nil {
		if !errors.Is(err, db.ErrNotFound) {
			return err
		}
		_, err = setSystemTimezone(defaultSystemTimezone)
		return err
	}

	if _, err := setSystemTimezone(value); err != nil {
		_, _ = setSystemTimezone(defaultSystemTimezone)
	}
	return nil
}

func normalizeSystemTimezone(raw string) (string, *time.Location, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		value = defaultSystemTimezone
	}

	loc, err := time.LoadLocation(value)
	if err != nil {
		return "", nil, errors.New("invalid timezone")
	}
	return value, loc, nil
}

func setSystemTimezone(raw string) (string, error) {
	normalized, loc, err := normalizeSystemTimezone(raw)
	if err != nil {
		return "", err
	}

	timezoneMu.Lock()
	timezoneName = normalized
	timezoneLocation = loc
	time.Local = loc
	timezoneMu.Unlock()

	return normalized, nil
}

func currentSystemTimezoneName() string {
	timezoneMu.RLock()
	defer timezoneMu.RUnlock()
	return timezoneName
}

func currentSystemTimezoneLocation() *time.Location {
	timezoneMu.RLock()
	defer timezoneMu.RUnlock()
	if timezoneLocation == nil {
		return time.UTC
	}
	return timezoneLocation
}

func formatTimeRFC3339OrEmpty(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.In(currentSystemTimezoneLocation()).Format(time.RFC3339)
}

func timeInSystemTimezonePtr(v *time.Time) *time.Time {
	if v == nil {
		return nil
	}
	loc := currentSystemTimezoneLocation()
	converted := v.In(loc)
	return &converted
}

func timeRFC3339OrEmpty(t time.Time) string {
	return formatTimeRFC3339OrEmpty(t)
}
