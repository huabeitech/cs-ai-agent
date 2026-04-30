package repositories

import (
	"fmt"
	"strings"
	"testing"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"

	"github.com/glebarez/sqlite"
	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
)

func TestKnowledgeDocumentRepositoryFindPageListByCndOmitsContentAndPaginates(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:knowledge_document_repository_test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite error = %v", err)
	}
	if err := db.AutoMigrate(&models.KnowledgeDocument{}); err != nil {
		t.Fatalf("auto migrate error = %v", err)
	}
	if err := db.Exec("DELETE FROM knowledge_documents").Error; err != nil {
		t.Fatalf("clean knowledge documents error = %v", err)
	}
	sqls.SetDB(db)

	for i := 1; i <= 3; i++ {
		item := &models.KnowledgeDocument{
			KnowledgeBaseID: 1,
			Title:           fmt.Sprintf("doc-%d", i),
			ContentType:     enums.KnowledgeDocumentContentTypeMarkdown,
			Content:         strings.Repeat("large document content ", 200),
			Status:          enums.StatusOk,
			IndexStatus:     enums.KnowledgeDocumentIndexStatusPending,
		}
		if err := db.Create(item).Error; err != nil {
			t.Fatalf("create document %d error = %v", i, err)
		}
	}

	list, paging := KnowledgeDocumentRepository.FindPageListByCnd(db, sqls.NewCnd().Eq("knowledge_base_id", 1).Asc("id").Page(1, 2))
	if len(list) != 2 {
		t.Fatalf("len(list) = %d, want 2", len(list))
	}
	if paging.Total != 3 {
		t.Fatalf("paging.Total = %d, want 3", paging.Total)
	}
	for _, item := range list {
		if item.Content != "" {
			t.Fatalf("list item content should be empty, got %q", item.Content)
		}
	}
}
