package app

import (
	"context"
)

func (s Service) ClearCache(_ context.Context, _ CacheClearRequest) (CacheClearResult, error) {
	result, err := s.cache.Clear()
	if err != nil {
		return CacheClearResult{}, err
	}

	return CacheClearResult{
		CacheRoot:      result.CacheRoot,
		RemovedEntries: result.RemovedEntries,
	}, nil
}
