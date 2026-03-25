package utils

import (
	"crypto/rand"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/errorsx"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cast"
)

func NormalizeNullableString(value *string) *string {
	if value == nil {
		return nil
	}
	v := strings.TrimSpace(*value)
	if v == "" {
		return nil
	}
	return &v
}

func BuildAuditFields(operator *dto.AuthPrincipal) models.AuditFields {
	now := time.Now()
	fields := models.AuditFields{
		CreatedAt: now,
		UpdatedAt: now,
	}
	if operator != nil {
		fields.CreateUserID = operator.UserID
		fields.CreateUserName = operator.Username
		fields.UpdateUserID = operator.UserID
		fields.UpdateUserName = operator.Username
	}
	return fields
}

func FormatTimePtr(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format(time.DateTime)
}

func GenerateRandomPassword(length int) (string, error) {
	if length <= 0 {
		return "", errorsx.InvalidParam("密码长度不合法")
	}

	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789"
	buf := make([]byte, length)
	random := make([]byte, length)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	for i := range buf {
		buf[i] = charset[int(random[i])%len(charset)]
	}
	return string(buf), nil
}

func JoinInt64s(values []int64) string {
	if len(values) == 0 {
		return ""
	}
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, cast.ToString(value))
	}
	return strings.Join(parts, ",")
}

func SplitInt64s(raw string) []int64 {
	parts := strings.Split(strings.TrimSpace(raw), ",")
	ret := make([]int64, 0, len(parts))
	seen := make(map[int64]struct{})
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			continue
		}
		id, err := strconv.ParseInt(value, 10, 64)
		if err != nil || id <= 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		ret = append(ret, id)
	}
	return ret
}
