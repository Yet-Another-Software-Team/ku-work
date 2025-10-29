package model

import "context"

// StorageDeleteFunc is the signature for a storage deletion hook.
//
// Models can call CallStorageDeleteHook when they need to remove the underlying
// stored object (for example when a parent model is deleted). The hook should
// be registered by the application startup code (for example, the file-handling
// service) to perform the actual deletion in the configured storage backend.
type StorageDeleteFunc func(ctx context.Context, fileID string) error

// storageDeleteHook holds the currently-registered storage deletion hook.
// When nil, deletion calls are treated as successful (no-op).
var storageDeleteHook StorageDeleteFunc

// SetStorageDeleteHook registers the provided hook. Call this during application
// initialization after constructing the storage provider. Passing nil clears the hook.
func SetStorageDeleteHook(h StorageDeleteFunc) {
	storageDeleteHook = h
}

// CallStorageDeleteHook invokes the registered storage deletion hook if present.
// If no hook is registered, this returns nil to keep deletion idempotent and to
// avoid introducing a hard dependency between model code and storage implementation.
func CallStorageDeleteHook(ctx context.Context, fileID string) error {
	if storageDeleteHook == nil {
		// No-op when no hook is registered.
		return nil
	}
	return storageDeleteHook(ctx, fileID)
}
