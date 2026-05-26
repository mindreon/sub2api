package repository_test

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

// TestCatalogModelRepoInterface verifies the repo satisfies the service interface at compile time.
func TestCatalogModelRepoInterface(t *testing.T) {
	t.Helper()
	var _ service.CatalogModelRepository = (*repository.CatalogModelRepo)(nil)
}
