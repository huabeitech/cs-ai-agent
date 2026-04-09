package vectordb

import (
	"context"
	"fmt"

	"cs-agent/internal/pkg/config"
	"cs-agent/internal/pkg/enums"
)

var defaultProvider Provider

func Init(cfg *config.VectorDBConfig) error {
	if cfg == nil || cfg.Type == "" {
		return nil
	}

	var err error
	switch enums.VectorDBType(cfg.Type) {
	case enums.VectorDBTypeQdrant:
		defaultProvider, err = NewQdrantProvider(cfg)
	default:
		return fmt.Errorf("unsupported vectordb type: %s", cfg.Type)
	}
	return err
}

func GetProvider() Provider {
	return defaultProvider
}

func Close() error {
	if defaultProvider != nil {
		return defaultProvider.Close()
	}
	return nil
}

func CreateCollection(ctx context.Context, name string, dimension int) error {
	if defaultProvider == nil {
		return fmt.Errorf("vectordb provider not initialized")
	}
	return defaultProvider.CreateCollection(ctx, name, dimension)
}

func DeleteCollection(ctx context.Context, name string) error {
	if defaultProvider == nil {
		return fmt.Errorf("vectordb provider not initialized")
	}
	return defaultProvider.DeleteCollection(ctx, name)
}

func GetCollection(ctx context.Context, name string) (*CollectionInfo, error) {
	if defaultProvider == nil {
		return nil, fmt.Errorf("vectordb provider not initialized")
	}
	return defaultProvider.GetCollection(ctx, name)
}

func ListCollections(ctx context.Context) ([]string, error) {
	if defaultProvider == nil {
		return nil, fmt.Errorf("vectordb provider not initialized")
	}
	return defaultProvider.ListCollections(ctx)
}

func UpsertVectors(ctx context.Context, collectionName string, vectors []Vector) error {
	if defaultProvider == nil {
		return fmt.Errorf("vectordb provider not initialized")
	}
	return defaultProvider.UpsertVectors(ctx, collectionName, vectors)
}

func DeleteVectors(ctx context.Context, collectionName string, ids []string) error {
	if defaultProvider == nil {
		return fmt.Errorf("vectordb provider not initialized")
	}
	return defaultProvider.DeleteVectors(ctx, collectionName, ids)
}

func Search(ctx context.Context, req *SearchRequest) ([]SearchResult, error) {
	if defaultProvider == nil {
		return nil, fmt.Errorf("vectordb provider not initialized")
	}
	return defaultProvider.Search(ctx, req)
}
