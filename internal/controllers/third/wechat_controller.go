package third

import (
	"cs-agent/internal/wxwork"
	"fmt"
	"net/http"

	"github.com/kataras/iris/v12"
	"github.com/silenceper/wechat/v2/work/kf"
)

type WechatController struct {
	Ctx iris.Context
}

// GetCallback GET请求用于校验回调是否配置正确
func (c *WechatController) GetCallback() {
	cli, err := wxwork.GetWorkCli().GetKF()
	if err != nil {
		c.Ctx.StopWithError(http.StatusInternalServerError, err)
		return
	}
	options := kf.SignatureOptions{}
	if err := c.Ctx.ReadForm(&options); err != nil {
		c.Ctx.StopWithError(http.StatusUnauthorized, err)
		return
	}
	// 调用VerifyURL方法校验当前请求，如果合法则把解密后的内容作为响应返回给微信服务器
	echo, err := cli.VerifyURL(options)
	if err == nil {
		c.Ctx.WriteString(echo)
	} else {
		c.Ctx.StopWithError(http.StatusUnauthorized, err)
	}
}

// PostCallback POST请求用于接收回调
func (c *WechatController) PostCallback() {
	cli, err := wxwork.GetWorkCli().GetKF()
	if err != nil {
		c.Ctx.StopWithError(http.StatusInternalServerError, err)
		return
	}
	var (
		message kf.CallbackMessage
		body    []byte
	)
	// 读取原始消息内容
	body, err = c.Ctx.GetBody()
	if err != nil {
		c.Ctx.StopWithError(http.StatusInternalServerError, err)
		return
	}
	// 解析原始数据
	message, err = cli.GetCallbackMessage(body)
	if err != nil {
		c.Ctx.StopWithError(http.StatusInternalServerError, err)
		return
	}

	// TODO 消息处理
	fmt.Println(message)

	c.Ctx.WriteString("ok")
}
