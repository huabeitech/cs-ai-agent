package factory

import (
	"context"
	"fmt"
	"strings"

	einoadapter "cs-agent/internal/ai/runtime/internal/impl/adapter"
	"cs-agent/internal/models"

	einoskill "github.com/cloudwego/eino/adk/middlewares/skill"
)

type selectedSkillBackend struct {
	frontMatter einoskill.FrontMatter
	skill       einoskill.Skill
}

func newSelectedSkillBackend(selectedSkill *models.SkillDefinition, toolDefinitions []einoadapter.MCPToolDefinition) (*selectedSkillBackend, error) {
	if selectedSkill == nil {
		return nil, fmt.Errorf("selected skill is nil")
	}
	skillName := strings.TrimSpace(selectedSkill.Code)
	if skillName == "" {
		return nil, fmt.Errorf("selected skill code is empty")
	}
	description := strings.TrimSpace(selectedSkill.Description)
	content := buildSelectedSkillDocument(selectedSkill, toolDefinitions)
	return &selectedSkillBackend{
		frontMatter: einoskill.FrontMatter{
			Name:        skillName,
			Description: description,
		},
		skill: einoskill.Skill{
			FrontMatter: einoskill.FrontMatter{
				Name:        skillName,
				Description: description,
			},
			Content: content,
		},
	}, nil
}

func (b *selectedSkillBackend) List(_ context.Context) ([]einoskill.FrontMatter, error) {
	if b == nil {
		return nil, nil
	}
	return []einoskill.FrontMatter{b.frontMatter}, nil
}

func (b *selectedSkillBackend) Get(_ context.Context, name string) (einoskill.Skill, error) {
	if b == nil {
		return einoskill.Skill{}, fmt.Errorf("selected skill backend is nil")
	}
	name = strings.TrimSpace(name)
	if name == "" || strings.EqualFold(name, b.frontMatter.Name) {
		return b.skill, nil
	}
	return einoskill.Skill{}, fmt.Errorf("skill %q not found", name)
}
