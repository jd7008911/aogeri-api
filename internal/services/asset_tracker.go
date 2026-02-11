// internal/services/asset_tracker.go
package services

import "context"

// AssetTracker is a lightweight stub for the asset tracking service.
type AssetTracker struct {
	ctx context.Context
}

func NewAssetTracker(ctx context.Context) *AssetTracker {
	return &AssetTracker{ctx: ctx}
}

// Start begins background tracking (stub).
func (a *AssetTracker) Start() error {
	// no-op for now
	return nil
}
