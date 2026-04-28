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

// ExternalUser 外部访客身份（IM 客户），与站内 AuthPrincipal 区分。
type ExternalUser struct {
	ExternalSource enums.ExternalSource `json:"externalSource"`
	ExternalID     string               `json:"externalId"`
	ExternalName   string               `json:"externalName"`
}

type UserTokenClaims struct {
	UserID string `json:"userId"`
	Name   string `json:"name"`
	jwt.RegisteredClaims
}

func GetExternalUser(ctx iris.Context, secret string) (*ExternalUser, error) {
	if userToken := getUserToken(ctx); strs.IsNotBlank(userToken) {
		claims, err := verifyUserToken(userToken, secret)
		if err != nil {
			return nil, err
		}
		return &ExternalUser{
			ExternalSource: enums.ExternalSourceUser,
			ExternalID:     claims.UserID,
			ExternalName:   claims.Name,
		}, nil
	}
	return getGuestUser(ctx)
}

func verifyUserToken(userToken, secret string) (*UserTokenClaims, error) {
	if strs.IsBlank(userToken) {
		return nil, errorsx.Unauthorized("用户身份不能为空")
	}
	if strs.IsBlank(secret) {
		return nil, errorsx.Unauthorized("用户身份校验未配置")
	}

	claims := &UserTokenClaims{}
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

	if strs.IsBlank(claims.UserID) {
		return nil, errorsx.Unauthorized("用户标识不能为空")
	}
	if strs.IsBlank(claims.Name) {
		return nil, errorsx.Unauthorized("用户名称不能为空")
	}
	if claims.ExpiresAt == nil {
		return nil, errorsx.Unauthorized("用户身份已过期")
	}

	return claims, nil
}

func getUserToken(ctx iris.Context) string {
	auth := strings.TrimSpace(ctx.GetHeader("Authorization"))
	if len(auth) > 7 && strings.EqualFold(auth[:7], "Bearer ") {
		if token := strings.TrimSpace(auth[7:]); token != "" {
			return token
		}
	}
	userToken, _ := params.Get(ctx, "userToken")
	return strings.TrimSpace(userToken)
}

func getGuestUser(ctx iris.Context) (*ExternalUser, error) {
	externalID := getExternalID(ctx)
	if strs.IsBlank(externalID) {
		return nil, errorsx.Unauthorized("用户标识不能为空")
	}
	return &ExternalUser{
		ExternalSource: enums.ExternalSourceGuest,
		ExternalID:     externalID,
		ExternalName:   getExternalName(ctx),
	}, nil
}

func getExternalID(ctx iris.Context) string {
	externalID := ctx.GetHeader("X-External-Id")
	if strs.IsBlank(externalID) {
		externalID, _ = params.Get(ctx, "externalId")
	}
	return externalID
}

func getExternalName(ctx iris.Context) string {
	externalName := ctx.GetHeader("X-External-Name")
	if strs.IsBlank(externalName) {
		externalName, _ = params.Get(ctx, "externalName")
	}
	if strs.IsNotBlank(externalName) {
		externalName, _ = url.QueryUnescape(externalName)
	}
	return externalName
}

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
