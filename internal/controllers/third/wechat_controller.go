package third

import "github.com/kataras/iris/v12"

type WechatController struct {
	Ctx iris.Context
}

// GetCallback GET请求用于校验回调是否配置正确
func (c *WechatController) GetCallback() {}

// PostCallback POST请求用于接收回调
func (c *WechatController) PostCallback() {

}
