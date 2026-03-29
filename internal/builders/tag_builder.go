package builders

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto/response"
	"time"
)

func BuildTagResponse(item *models.Tag) response.TagResponse {
	if item == nil {
		return response.TagResponse{}
	}
	return response.TagResponse{
		ID:        item.ID,
		ParentID:  item.ParentID,
		Name:      item.Name,
		Remark:    item.Remark,
		SortNo:    item.SortNo,
		Status:    item.Status,
		CreatedAt: item.CreatedAt.Format(time.DateTime),
		UpdatedAt: item.UpdatedAt.Format(time.DateTime),
	}
}

func BuildTagResponses(list []models.Tag) []response.TagResponse {
	if len(list) == 0 {
		return nil
	}
	results := make([]response.TagResponse, 0, len(list))
	for i := range list {
		results = append(results, BuildTagResponse(&list[i]))
	}
	return results
}

func BuildTagTreeResponses(list []models.Tag) []*response.TagTreeResponse {
	if len(list) == 0 {
		return nil
	}

	nodeMap := make(map[int64]*response.TagTreeResponse, len(list))
	roots := make([]*response.TagTreeResponse, 0)

	for i := range list {
		item := &list[i]
		nodeMap[item.ID] = &response.TagTreeResponse{
			ID:       item.ID,
			ParentID: item.ParentID,
			Name:     item.Name,
			Remark:   item.Remark,
			SortNo:   item.SortNo,
			Status:   item.Status,
			Children: make([]*response.TagTreeResponse, 0),
		}
	}

	for i := range list {
		item := &list[i]
		node := nodeMap[item.ID]
		if item.ParentID == 0 {
			roots = append(roots, node)
			continue
		}
		parent, ok := nodeMap[item.ParentID]
		if !ok {
			roots = append(roots, node)
			continue
		}
		parent.Children = append(parent.Children, node)
	}

	return roots
}
