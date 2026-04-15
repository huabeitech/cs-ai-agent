package dashboard

import (
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type DashboardController struct {
	Ctx iris.Context
}

func (c *DashboardController) GetOverview() *web.JsonResult {
	rangeValue, _ := params.Get(c.Ctx, "range")
	return web.JsonData(services.DashboardService.GetOverview(rangeValue))
}
