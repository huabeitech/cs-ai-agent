package kb

import (
	"bytes"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/repositories"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mlogclub/simple/sqls"
	"golang.org/x/net/html"
	"gorm.io/gorm"
)

type Chapter struct {
	Title   string
	Link    string
	Content string
}

type InitResult struct {
	KnowledgeBaseID  int64
	TotalChapters    int
	CreatedDocuments int
	UpdatedDocuments int
}

func Init() (*InitResult, error) {
	chapters, err := read("水浒传")
	if err != nil {
		return nil, fmt.Errorf("read books failed: %w", err)
	}

	if len(chapters) == 0 {
		return nil, fmt.Errorf("no chapters found")
	}

	result := &InitResult{TotalChapters: len(chapters)}
	err = sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		kbModel, ensureErr := ensureKnowledgeBase(ctx.Tx)
		if ensureErr != nil {
			return ensureErr
		}
		result.KnowledgeBaseID = kbModel.ID

		for _, chapter := range chapters {
			created, upsertErr := upsertKnowledgeDocument(ctx.Tx, kbModel.ID, chapter)
			if upsertErr != nil {
				return upsertErr
			}
			if created {
				result.CreatedDocuments++
			} else {
				result.UpdatedDocuments++
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func ensureKnowledgeBase(db *gorm.DB) (*models.KnowledgeBase, error) {
	now := time.Now()
	item := repositories.KnowledgeBaseRepository.FindOne(db, sqls.NewCnd().Eq("name", "水浒传"))
	if item == nil {
		item = &models.KnowledgeBase{
			Name:                  "水浒传",
			Description:           "四大名著测试数据",
			Status:                enums.StatusOk,
			DefaultTopK:           10,
			DefaultScoreThreshold: 0.2,
			DefaultRerankLimit:    5,
			ChunkProvider:         string(enums.KnowledgeChunkProviderStructured),
			ChunkTargetTokens:     300,
			ChunkMaxTokens:        400,
			ChunkOverlapTokens:    40,
			AnswerMode:            int(enums.KnowledgeAnswerModeStrict),
			FallbackMode:          int(enums.KnowledgeFallbackModeNoAnswer),
			AuditFields: models.AuditFields{
				CreatedAt:      now,
				CreateUserID:   constants.SystemAuditUserID,
				CreateUserName: constants.SystemAuditUserName,
				UpdatedAt:      now,
				UpdateUserID:   constants.SystemAuditUserID,
				UpdateUserName: constants.SystemAuditUserName,
			},
		}
		if err := repositories.KnowledgeBaseRepository.Create(db, item); err != nil {
			return nil, err
		}
		return item, nil
	}

	err := repositories.KnowledgeBaseRepository.Updates(db, item.ID, map[string]any{
		"description":          "四大名著测试数据",
		"status":               enums.StatusOk,
		"chunk_provider":       string(enums.KnowledgeChunkProviderStructured),
		"chunk_target_tokens":  300,
		"chunk_max_tokens":     400,
		"chunk_overlap_tokens": 40,
		"update_user_id":       constants.SystemAuditUserID,
		"update_user_name":     constants.SystemAuditUserName,
		"updated_at":           now,
	})
	if err != nil {
		return nil, err
	}
	return repositories.KnowledgeBaseRepository.Get(db, item.ID), nil
}

func upsertKnowledgeDocument(db *gorm.DB, knowledgeBaseID int64, chapter Chapter) (bool, error) {
	now := time.Now()
	item := repositories.KnowledgeDocumentRepository.FindOne(db, sqls.NewCnd().Eq("knowledge_base_id", knowledgeBaseID).Eq("title", chapter.Title))
	if item == nil {
		item = &models.KnowledgeDocument{
			KnowledgeBaseID: knowledgeBaseID,
			Title:           chapter.Title,
			ContentType:     enums.KnowledgeDocumentContentTypeMarkdown,
			Content:         chapter.Content,
			Status:          enums.StatusOk,
			AuditFields: models.AuditFields{
				CreatedAt:      now,
				CreateUserID:   constants.SystemAuditUserID,
				CreateUserName: constants.SystemAuditUserName,
				UpdatedAt:      now,
				UpdateUserID:   constants.SystemAuditUserID,
				UpdateUserName: constants.SystemAuditUserName,
			},
		}
		if err := repositories.KnowledgeDocumentRepository.Create(db, item); err != nil {
			return false, err
		}
		return true, nil
	}

	err := repositories.KnowledgeDocumentRepository.Updates(db, item.ID, map[string]any{
		"content_type":     enums.KnowledgeDocumentContentTypeMarkdown,
		"content":          chapter.Content,
		"status":           enums.StatusOk,
		"update_user_id":   constants.SystemAuditUserID,
		"update_user_name": constants.SystemAuditUserName,
		"updated_at":       now,
	})
	if err != nil {
		return false, err
	}
	return false, nil
}

func read(name string) (chapters []Chapter, err error) {
	path := filepath.Join("cmd", "testdata", "kb", name+".html")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	doc, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	baseDir := filepath.Dir(path)
	seen := make(map[string]struct{})

	var walk func(*html.Node) error
	walk = func(n *html.Node) error {
		if n.Type == html.ElementNode && n.Data == "a" {
			parent := n.Parent
			if parent != nil && parent.Type == html.ElementNode && parent.Data == "span" && parent.Parent != nil && hasClass(parent.Parent, "chapter") {
				href := strings.TrimSpace(getAttr(n, "href"))
				if href != "" {
					resolved := filepath.Clean(filepath.Join(baseDir, href))
					if _, ok := seen[resolved]; !ok {
						seen[resolved] = struct{}{}

						title := cleanText(nodeText(parent))
						title = strings.ReplaceAll(title, "（原文）", "")
						title = strings.TrimSpace(title)

						content, readErr := readChapterContent(resolved)
						if readErr != nil {
							return readErr
						}

						chapters = append(chapters, Chapter{
							Title:   title,
							Link:    resolved,
							Content: content,
						})
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := walk(c); err != nil {
				return err
			}
		}
		return nil
	}

	err = walk(doc)
	if err != nil {
		return nil, err
	}

	return chapters, nil
}

func readChapterContent(link string) (content string, err error) {
	data, err := os.ReadFile(link)
	if err != nil {
		return "", err
	}

	doc, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	lines := make([]string, 0, 128)
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "h1" {
				title := cleanText(nodeText(n))
				title = strings.ReplaceAll(title, " 原文", "")
				title = strings.TrimSpace(title)
				if title != "" {
					lines = append(lines, title)
				}
			}

			if n.Data == "p" {
				if hasClass(n, "next") || getAttr(n, "id") == "home" || getAttr(n, "id") == "list" {
					return
				}
				if isInsideClass(n, "pn") {
					return
				}
				line := cleanText(nodeText(n))
				if line != "" {
					lines = append(lines, line)
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(doc)
	return strings.Join(lines, "\n"), nil
}

func getAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func hasClass(n *html.Node, className string) bool {
	classes := strings.Fields(getAttr(n, "class"))
	for _, c := range classes {
		if c == className {
			return true
		}
	}
	return false
}

func isInsideClass(n *html.Node, className string) bool {
	for p := n.Parent; p != nil; p = p.Parent {
		if hasClass(p, className) {
			return true
		}
	}
	return false
}

func nodeText(n *html.Node) string {
	if n == nil {
		return ""
	}

	var b strings.Builder
	var walk func(*html.Node)
	walk = func(cur *html.Node) {
		if cur.Type == html.TextNode {
			b.WriteString(cur.Data)
		}
		for c := cur.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(n)
	return b.String()
}

func cleanText(s string) string {
	s = strings.ReplaceAll(s, "\u00a0", " ")
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	return strings.Join(strings.Fields(s), " ")
}
