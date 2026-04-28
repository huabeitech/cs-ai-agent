package services

import (
	"errors"
	"strings"
	"time"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/config"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/openidentity"
	"cs-agent/internal/repositories"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

const (
	customerSessionTokenType = "customer_session"
	customerSessionHeader    = "X-Customer-Session-Token"
	customerSessionExpHeader = "X-Customer-Session-Expires-At"
)

var CustomerSessionService = newCustomerSessionService()

func newCustomerSessionService() *customerSessionService {
	return &customerSessionService{}
}

type customerSessionService struct {
}

type customerSessionClaims struct {
	TokenType    string `json:"typ"`
	ChannelID    int64  `json:"channelId"`
	ChannelCode  string `json:"channelCode"`
	CustomerID   int64  `json:"customerId"`
	CustomerName string `json:"customerName"`
	IdentityKey  string `json:"identityKey"`
	jwt.RegisteredClaims
}

type CustomerSessionVerifyResult struct {
	ExternalUser *openidentity.ExternalUser
	Token        string
	ExpiresAt    time.Time
	Refreshed    bool
}

func (s *customerSessionService) Exchange(channel *models.Channel, externalUser openidentity.ExternalUser) (*response.CustomerSessionExchangeResponse, error) {
	if channel == nil || channel.Status != enums.StatusOk {
		return nil, errorsx.InvalidParam("接入渠道不存在或已停用")
	}
	var customerID int64
	if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		id, err := CustomerService.EnsureExternalCustomer(ctx, externalUser)
		if err != nil {
			return err
		}
		customerID = id
		return nil
	}); err != nil {
		return nil, err
	}
	customer := CustomerService.Get(customerID)
	if customer == nil || customer.Status == enums.StatusDeleted {
		return nil, errorsx.InvalidParam("客户不存在")
	}
	token, expiresAt, err := s.Sign(channel, customer, externalUser)
	if err != nil {
		return nil, err
	}
	return &response.CustomerSessionExchangeResponse{
		CustomerSessionToken: token,
		ExpiresAt:            expiresAt.Format(time.DateTime),
		IdentityKey:          s.identityKey(externalUser),
		Customer: response.CustomerSessionCustomerResponse{
			ID:   customer.ID,
			Name: strings.TrimSpace(customer.Name),
		},
	}, nil
}

func (s *customerSessionService) Sign(channel *models.Channel, customer *models.Customer, externalUser openidentity.ExternalUser) (string, time.Time, error) {
	cfg := config.Current().CustomerSession
	secret := strings.TrimSpace(cfg.Secret)
	if secret == "" {
		return "", time.Time{}, errorsx.BusinessError(1, "客服会话密钥未配置")
	}
	if channel == nil || customer == nil {
		return "", time.Time{}, errorsx.InvalidParam("客服会话参数不完整")
	}
	now := time.Now()
	expiresAt := now.Add(time.Duration(cfg.TTL()) * time.Minute)
	claims := customerSessionClaims{
		TokenType:    customerSessionTokenType,
		ChannelID:    channel.ID,
		ChannelCode:  strings.TrimSpace(channel.ChannelID),
		CustomerID:   customer.ID,
		CustomerName: strings.TrimSpace(customer.Name),
		IdentityKey:  s.identityKey(externalUser),
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, err
	}
	return token, expiresAt, nil
}

func (s *customerSessionService) VerifyRequest(ctx iris.Context, channel *models.Channel) (*CustomerSessionVerifyResult, error) {
	token := s.getCustomerSessionToken(ctx)
	if token == "" {
		return nil, errorsx.Unauthorized("客服会话不能为空")
	}
	claims, err := s.verifyToken(token)
	if err != nil {
		return nil, err
	}
	if channel == nil || channel.Status != enums.StatusOk {
		return nil, errorsx.InvalidParam("接入渠道不存在或已停用")
	}
	if claims.ChannelID != channel.ID || strings.TrimSpace(claims.ChannelCode) != strings.TrimSpace(channel.ChannelID) {
		return nil, errorsx.Unauthorized("客服会话校验失败")
	}
	customer := CustomerService.Get(claims.CustomerID)
	if customer == nil || customer.Status == enums.StatusDeleted {
		return nil, errorsx.Unauthorized("客服会话校验失败")
	}
	external, err := s.externalUserFromClaims(claims, customer)
	if err != nil {
		return nil, err
	}
	result := &CustomerSessionVerifyResult{
		ExternalUser: external,
		Token:        token,
		ExpiresAt:    claims.ExpiresAt.Time,
	}
	if s.shouldRefresh(claims.ExpiresAt.Time) {
		newToken, expiresAt, err := s.Sign(channel, customer, *external)
		if err != nil {
			return nil, err
		}
		result.Token = newToken
		result.ExpiresAt = expiresAt
		result.Refreshed = true
	}
	return result, nil
}

func (s *customerSessionService) SetRefreshHeaders(ctx iris.Context, result *CustomerSessionVerifyResult) {
	if ctx == nil || result == nil || !result.Refreshed {
		return
	}
	ctx.Header(customerSessionHeader, result.Token)
	ctx.Header(customerSessionExpHeader, result.ExpiresAt.Format(time.DateTime))
}

func (s *customerSessionService) verifyToken(rawToken string) (*customerSessionClaims, error) {
	cfg := config.Current().CustomerSession
	secret := strings.TrimSpace(cfg.Secret)
	if secret == "" {
		return nil, errorsx.BusinessError(1, "客服会话密钥未配置")
	}
	claims := &customerSessionClaims{}
	token, err := jwt.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (any, error) {
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
			return nil, errorsx.Unauthorized("客服会话已过期")
		}
		return nil, errorsx.Unauthorized("客服会话校验失败")
	}
	if token == nil || !token.Valid || claims.TokenType != customerSessionTokenType || claims.ExpiresAt == nil {
		return nil, errorsx.Unauthorized("客服会话校验失败")
	}
	if claims.ChannelID <= 0 || strings.TrimSpace(claims.ChannelCode) == "" || claims.CustomerID <= 0 || strings.TrimSpace(claims.IdentityKey) == "" {
		return nil, errorsx.Unauthorized("客服会话校验失败")
	}
	return claims, nil
}

func (s *customerSessionService) externalUserFromClaims(claims *customerSessionClaims, customer *models.Customer) (*openidentity.ExternalUser, error) {
	identityKey := strings.TrimSpace(claims.IdentityKey)
	parts := strings.SplitN(identityKey, ":", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[1]) == "" {
		return nil, errorsx.Unauthorized("客服会话校验失败")
	}
	var source enums.ExternalSource
	switch parts[0] {
	case "user":
		source = enums.ExternalSourceUser
	case "guest":
		source = enums.ExternalSourceGuest
	default:
		return nil, errorsx.Unauthorized("客服会话校验失败")
	}
	identity := repositories.CustomerIdentityRepository.GetBy(sqls.DB(), source, parts[1])
	if identity == nil || identity.CustomerID != claims.CustomerID {
		return nil, errorsx.Unauthorized("客服会话校验失败")
	}
	name := strings.TrimSpace(claims.CustomerName)
	if customer != nil && strings.TrimSpace(customer.Name) != "" {
		name = strings.TrimSpace(customer.Name)
	}
	return &openidentity.ExternalUser{
		ExternalSource: source,
		ExternalID:     parts[1],
		ExternalName:   name,
	}, nil
}

func (s *customerSessionService) shouldRefresh(expiresAt time.Time) bool {
	threshold := config.Current().CustomerSession.RefreshThreshold()
	return time.Until(expiresAt) <= time.Duration(threshold)*time.Minute
}

func (s *customerSessionService) identityKey(externalUser openidentity.ExternalUser) string {
	switch externalUser.ExternalSource {
	case enums.ExternalSourceUser:
		return "user:" + strings.TrimSpace(externalUser.ExternalID)
	default:
		return "guest:" + strings.TrimSpace(externalUser.ExternalID)
	}
}

func (s *customerSessionService) getCustomerSessionToken(ctx iris.Context) string {
	auth := strings.TrimSpace(ctx.GetHeader("Authorization"))
	if len(auth) > 7 && strings.EqualFold(auth[:7], "Bearer ") {
		if token := strings.TrimSpace(auth[7:]); token != "" {
			return token
		}
	}
	token, _ := params.Get(ctx, "customerSessionToken")
	return strings.TrimSpace(token)
}
