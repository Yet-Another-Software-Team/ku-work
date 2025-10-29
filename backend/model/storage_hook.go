package model

import "context"

// StorageDeleteFunc is the signature for a storage deletion hook.
// It is called by models to remove the underlying stored object.
type StorageDeleteFunc func(ctx context.Context, fileID string) error

// storageDeleteHook holds the currently-registered storage deletion hook.
var storageDeleteHook StorageDeleteFunc

// SetStorageDeleteHook registers the provided hook.
func SetStorageDeleteHook(h StorageDeleteFunc) {
	storageDeleteHook = h
}

// CallStorageDeleteHook invokes the registered storage deletion hook.
// If no hook is registered, it returns nil.
func CallStorageDeleteHook(ctx context.Context, fileID string) error {
	if storageDeleteHook == nil {
		return nil
	}
	return storageDeleteHook(ctx, fileID)
}
