package irisx

import (
	"cs-agent/internal/pkg/openidentity"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/web/params"
)

const (
	ctxKeyExternalUser = "externalUser"
)

func SetExternalUser(ctx iris.Context, ext *openidentity.ExternalUser) {
	ctx.Values().Set(ctxKeyExternalUser, ext)
}

func GetExternalUser(ctx iris.Context) *openidentity.ExternalUser {
	v := ctx.Values().Get(ctxKeyExternalUser)
	ext, _ := v.(*openidentity.ExternalUser)
	return ext
}

func GetChannelID(ctx iris.Context) string {
	if channelID := ctx.GetHeader("X-Channel-ID"); strs.IsNotBlank(channelID) {
		return channelID
	}
	if channelID, _ := params.Get(ctx, "channelId"); strs.IsNotBlank(channelID) {
		return channelID
	}
	return ""
}
