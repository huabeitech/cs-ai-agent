package dashboard

import (
	"cs-agent/internal/builders"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type CustomerContactController struct {
	Ctx iris.Context
}

// AnyList GET/POST /customer-contact/list?customerId=
func (c *CustomerContactController) AnyList() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionCustomerView); err != nil {
		return web.JsonError(err)
	}
	customerID, _ := params.GetInt64(c.Ctx, "customerId")
	if customerID <= 0 {
		return web.JsonErrorMsg("customerId 必填")
	}
	list := services.CustomerContactService.FindActiveByCustomerID(customerID)
	return web.JsonData(builders.BuildCustomerContactList(list))
}

func (c *CustomerContactController) PostCreate() *web.JsonResult {
	user, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionCustomerUpdate)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.CreateCustomerContactRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	item, err := services.CustomerContactService.CreateCustomerContact(req, user)
	if err != nil {
		return web.JsonError(err)
	}
	ret := builders.BuildCustomerContactResponse(item)
	return web.JsonData(&ret)
}

func (c *CustomerContactController) PostUpdate() *web.JsonResult {
	user, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionCustomerUpdate)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.UpdateCustomerContactRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.CustomerContactService.UpdateCustomerContact(req, user); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *CustomerContactController) PostDelete() *web.JsonResult {
	user, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionCustomerUpdate)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.DeleteCustomerContactRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.CustomerContactService.DeleteCustomerContact(req.ID, user); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}
