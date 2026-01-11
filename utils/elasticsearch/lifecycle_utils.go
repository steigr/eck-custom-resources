package elasticsearch

import "time"

// GetResourceCreatedAt extracts the created_at timestamp from a resource's _meta field
// Returns the timestamp or nil if not found or invalid
func GetResourceCreatedAt(meta map[string]any) (*time.Time, error) {
	return getTimestampFromMeta(meta, "created_at")
}

// GetResourceUpdatedAt extracts the updated_at timestamp from a resource's _meta field
// Returns the timestamp or nil if not found or invalid
func GetResourceUpdatedAt(meta map[string]any) (*time.Time, error) {
	return getTimestampFromMeta(meta, "updated_at")
}

// getTimestampFromMeta extracts a timestamp from a _meta map by key
// Returns the timestamp or nil if not found or invalid
func getTimestampFromMeta(meta map[string]any, key string) (*time.Time, error) {
	if meta == nil {
		return nil, nil
	}

	timestampStr, ok := meta[key].(string)
	if !ok {
		return nil, nil
	}

	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		return nil, err
	}

	return &timestamp, nil
}
