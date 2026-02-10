package db_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"sboard/panel/internal/db"
)

func TestSystemSettingsCRUD(t *testing.T) {
	store := setupStore(t)
	ctx := context.Background()

	_, err := store.GetSystemSetting(ctx, "timezone")
	require.ErrorIs(t, err, db.ErrNotFound)

	require.NoError(t, store.UpsertSystemSetting(ctx, "timezone", "Asia/Hong_Kong"))
	value, err := store.GetSystemSetting(ctx, "timezone")
	require.NoError(t, err)
	require.Equal(t, "Asia/Hong_Kong", value)

	require.NoError(t, store.UpsertSystemSetting(ctx, "timezone", "UTC"))
	value, err = store.GetSystemSetting(ctx, "timezone")
	require.NoError(t, err)
	require.Equal(t, "UTC", value)

	require.NoError(t, store.DeleteSystemSetting(ctx, "timezone"))
	_, err = store.GetSystemSetting(ctx, "timezone")
	require.ErrorIs(t, err, db.ErrNotFound)

	// 删除不存在的 key 应保持幂等。
	require.NoError(t, store.DeleteSystemSetting(ctx, "timezone"))
}
