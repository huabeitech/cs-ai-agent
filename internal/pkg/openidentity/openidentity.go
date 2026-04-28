// Package openidentity 解析开放 IM 场景下的外部访客身份（HTTP Header / Query），与 JSON 请求体 DTO 解耦。
package openidentity

import (
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"errors"
	"net/url"
	"strings"

	"github.com/golang-jwt/jwt/v5"
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

type UserTokenClaims struct {
	UserID string `json:"userId"`
	Name   string `json:"name"`
	Exp    int64  `json:"exp"`
	Iat    int64  `json:"iat"`
}

type userTokenJWTClaims struct {
	UserID string `json:"userId"`
	Name   string `json:"name"`
	jwt.RegisteredClaims
}

// GetExternalInfo 从 Header（X-External-*）或 query（externalSource、externalId、externalName）解析身份。
func GetExternalInfo(ctx iris.Context) (*ExternalInfo, error) {
	return GetExternalInfoWithUserTokenSecret(ctx, "")
}

func GetExternalInfoWithUserTokenSecret(ctx iris.Context, userTokenSecret string) (*ExternalInfo, error) {
	if userToken := parseUserToken(ctx); userToken != "" {
		claims, err := VerifyUserToken(userToken, userTokenSecret)
		if err != nil {
			return nil, err
		}
		return &ExternalInfo{
			ExternalSource: enums.ExternalSourceUser,
			ExternalID:     claims.UserID,
			ExternalName:   claims.Name,
		}, nil
	}
	externalSource, err := parseExternalSource(ctx)
	if err != nil {
		return nil, err
	}
	if !enums.IsAllowedOpenImExternalSource(externalSource) {
		return nil, errorsx.InvalidParam("不支持的外部来源")
	}
	if externalSource == enums.ExternalSourceUser {
		return nil, errorsx.Unauthorized("用户身份不能为空")
	}
	externalID, err := parseExternalID(ctx)
	if err != nil {
		return nil, err
	}
	return &ExternalInfo{
		ExternalSource: externalSource,
		// TODO: 对接业务系统后，根据业务系统用户信息识别 user；未对接时统一按访客处理。
		ExternalID:   externalID,
		ExternalName: parseExternalName(ctx),
	}, nil
}

func VerifyUserToken(userToken, secret string) (*UserTokenClaims, error) {
	userToken = strings.TrimSpace(userToken)
	secret = strings.TrimSpace(secret)
	if userToken == "" {
		return nil, errorsx.Unauthorized("用户身份不能为空")
	}
	if secret == "" {
		return nil, errorsx.Unauthorized("用户身份校验未配置")
	}

	claims := &userTokenJWTClaims{}
	token, err := jwt.ParseWithClaims(userToken, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unsupported signing method")
		}
		return []byte(secret), nil
	}, jwt.WithExpirationRequired(), jwt.WithValidMethods([]string{
		jwt.SigningMethodHS256.Alg(),
		jwt.SigningMethodHS384.Alg(),
		jwt.SigningMethodHS512.Alg(),
	}))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errorsx.Unauthorized("用户身份已过期")
		}
		return nil, errorsx.Unauthorized("用户身份校验失败")
	}
	if token == nil || !token.Valid {
		return nil, errorsx.Unauthorized("用户身份校验失败")
	}

	userID := strings.TrimSpace(claims.UserID)
	name := strings.TrimSpace(claims.Name)
	if userID == "" {
		return nil, errorsx.Unauthorized("用户标识不能为空")
	}
	if name == "" {
		return nil, errorsx.Unauthorized("用户名称不能为空")
	}
	if claims.ExpiresAt == nil {
		return nil, errorsx.Unauthorized("用户身份已过期")
	}

	result := &UserTokenClaims{
		UserID: userID,
		Name:   name,
		Exp:    claims.ExpiresAt.Unix(),
	}
	if claims.IssuedAt != nil {
		result.Iat = claims.IssuedAt.Unix()
	}
	return result, nil
}

func parseUserToken(ctx iris.Context) string {
	auth := strings.TrimSpace(ctx.GetHeader("Authorization"))
	if len(auth) > 7 && strings.EqualFold(auth[:7], "Bearer ") {
		if token := strings.TrimSpace(auth[7:]); token != "" {
			return token
		}
	}
	userToken, _ := params.Get(ctx, "userToken")
	return strings.TrimSpace(userToken)
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
	return decodeExternalDisplayName(externalName)
}

// decodeExternalDisplayName 将客户端对 X-External-Name / externalName 做的 encodeURIComponent 还原为 UTF-8。
// 无百分号编码时 QueryUnescape 原样返回，解码失败则保留原串（兼容异常或旧客户端明文）。
func decodeExternalDisplayName(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	dec, err := url.QueryUnescape(s)
	if err != nil {
		return s
	}
	return strings.TrimSpace(dec)
}
