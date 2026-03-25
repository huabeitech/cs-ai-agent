package errorsx

import "github.com/mlogclub/simple/web"

const (
	CodeInvalidParam       = 1000
	CodeBusinessError      = 2000
	CodeAuthUnauthorized   = 3000
	CodeAuthForbidden      = 3001
	CodeAuthInvalidToken   = 3002
	CodeAuthInvalidAccount = 3003
)

func InvalidParam(message string) error {
	return web.NewError(CodeInvalidParam, message)
}

func BusinessError(code int, message string) error {
	return web.NewError(CodeBusinessError+code, message)
}

func Unauthorized(message string) error {
	return web.NewError(CodeAuthUnauthorized, message)
}

func Forbidden(message string) error {
	return web.NewError(CodeAuthForbidden, message)
}

func InvalidToken(message string) error {
	return web.NewError(CodeAuthInvalidToken, message)
}

func InvalidAccount(message string) error {
	return web.NewError(CodeAuthInvalidAccount, message)
}
