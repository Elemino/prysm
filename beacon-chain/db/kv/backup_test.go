package kv

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/prysmaticlabs/prysm/shared/testutil"
	"github.com/prysmaticlabs/prysm/shared/testutil/require"
)

func TestStore_Backup(t *testing.T) {
	db, err := NewKVStore(context.Background(), t.TempDir())
	require.NoError(t, err, "Failed to instantiate DB")
	ctx := context.Background()

	head := testutil.NewBeaconBlock()
	head.Block.Slot = 5000

	require.NoError(t, db.SaveBlock(ctx, head))
	root, err := head.Block.HashTreeRoot()
	require.NoError(t, err)
	st, err := testutil.NewBeaconState()
	require.NoError(t, err)
	require.NoError(t, db.SaveState(ctx, st, root))
	require.NoError(t, db.SaveHeadBlockRoot(ctx, root))

	require.NoError(t, db.Backup(ctx, ""))

	backupsPath := filepath.Join(db.databasePath, backupsDirectoryName)
	files, err := ioutil.ReadDir(backupsPath)
	require.NoError(t, err)
	require.NotEqual(t, 0, len(files), "No backups created")
	require.NoError(t, db.Close(), "Failed to close database")

	oldFilePath := filepath.Join(backupsPath, files[0].Name())
	newFilePath := filepath.Join(backupsPath, DatabaseFileName)
	// We rename the file to match the database file name
	// our NewKVStore function expects when opening a database.
	require.NoError(t, os.Rename(oldFilePath, newFilePath))

	backedDB, err := NewKVStore(ctx, backupsPath)
	require.NoError(t, err, "Failed to instantiate DB")
	t.Cleanup(func() {
		require.NoError(t, backedDB.Close(), "Failed to close database")
	})
	require.Equal(t, true, backedDB.HasState(ctx, root))
}
