// Package openidentity 解析开放 IM 场景下的外部访客身份（HTTP Header / Query），与 JSON 请求体 DTO 解耦。
package openidentity

import (
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"strings"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/web/params"
)

// ExternalInfo 外部访客身份（IM 客户），与站内 AuthPrincipal 区分。
type ExternalInfo struct {
	ExternalSource enums.ExternalSource `json:"externalSource"`
	ExternalID     string               `json:"externalId"`
	ExternalName   string               `json:"externalName"`
}

// GetExternalInfo 从 Header（X-External-*）或 query（externalSource、externalId、externalName）解析身份。
func GetExternalInfo(ctx iris.Context) (*ExternalInfo, error) {
	externalSource, err := parseExternalSource(ctx)
	if err != nil {
		return nil, err
	}
	if !enums.IsAllowedOpenImExternalSource(externalSource) {
		return nil, errorsx.InvalidParam("不支持的外部来源")
	}
	externalID, err := parseExternalID(ctx)
	if err != nil {
		return nil, err
	}
	return &ExternalInfo{
		ExternalSource: externalSource,
		ExternalID:     externalID,
		ExternalName:   parseExternalName(ctx),
	}, nil
}

func parseExternalSource(ctx iris.Context) (enums.ExternalSource, error) {
	externalSource := ctx.GetHeader("X-External-Source")
	if strs.IsBlank(externalSource) {
		externalSource, _ = params.Get(ctx, "externalSource")
	}
	if strs.IsBlank(externalSource) {
		return "", errorsx.Unauthorized("用户来源不能为空")
	}
	return enums.ExternalSource(strings.TrimSpace(externalSource)), nil
}

func parseExternalID(ctx iris.Context) (string, error) {
	externalID := ctx.GetHeader("X-External-Id")
	if strs.IsBlank(externalID) {
		externalID, _ = params.Get(ctx, "externalId")
	}
	if strs.IsBlank(externalID) {
		return "", errorsx.Unauthorized("用户标识不能为空")
	}
	return strings.TrimSpace(externalID), nil
}

func parseExternalName(ctx iris.Context) string {
	externalName := ctx.GetHeader("X-External-Name")
	if strs.IsBlank(externalName) {
		externalName, _ = params.Get(ctx, "externalName")
	}
	return strings.TrimSpace(externalName)
}
