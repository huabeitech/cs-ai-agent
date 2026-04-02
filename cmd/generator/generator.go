package main

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/codegen"
)

func main() {
	codegen.GenerateWithOption(
		codegen.Options{
			BaseDir:    "./",
			PkgName:    "cs-agent",
			Version:    1,
			Repository: true,
			Service:    true,
			Controller: false,
			WebIndex:   false,
			WebEdit:    false,
		},
		codegen.GetGenerateStruct(&models.Migration{}),
		codegen.GetGenerateStruct(&models.User{}),
		codegen.GetGenerateStruct(&models.UserIdentity{}),
		codegen.GetGenerateStruct(&models.Company{}),
		codegen.GetGenerateStruct(&models.Customer{}),
		codegen.GetGenerateStruct(&models.CustomerIdentity{}),
		codegen.GetGenerateStruct(&models.CustomerContact{}),
		codegen.GetGenerateStruct(&models.Role{}),
		codegen.GetGenerateStruct(&models.Permission{}),
		codegen.GetGenerateStruct(&models.UserRole{}),
		codegen.GetGenerateStruct(&models.RolePermission{}),
		codegen.GetGenerateStruct(&models.UserPermission{}),
		codegen.GetGenerateStruct(&models.LoginSession{}),
		codegen.GetGenerateStruct(&models.LoginCredentialLog{}),
		codegen.GetGenerateStruct(&models.Asset{}),
		codegen.GetGenerateStruct(&models.Tag{}),
		codegen.GetGenerateStruct(&models.Conversation{}),
		codegen.GetGenerateStruct(&models.ConversationParticipant{}),
		codegen.GetGenerateStruct(&models.Message{}),
		codegen.GetGenerateStruct(&models.ConversationAssignment{}),
		codegen.GetGenerateStruct(&models.ConversationTag{}),
		codegen.GetGenerateStruct(&models.QuickReply{}),
		codegen.GetGenerateStruct(&models.AIAgent{}),
		codegen.GetGenerateStruct(&models.WidgetSite{}),
		codegen.GetGenerateStruct(&models.ConversationEventLog{}),
		codegen.GetGenerateStruct(&models.Ticket{}),
		codegen.GetGenerateStruct(&models.TicketCategory{}),
		codegen.GetGenerateStruct(&models.TicketResolutionCode{}),
		codegen.GetGenerateStruct(&models.TicketSLAConfig{}),
		codegen.GetGenerateStruct(&models.TicketComment{}),
		codegen.GetGenerateStruct(&models.TicketWatcher{}),
		codegen.GetGenerateStruct(&models.TicketEventLog{}),
		codegen.GetGenerateStruct(&models.TicketSLARecord{}),
		codegen.GetGenerateStruct(&models.TicketRelation{}),
		codegen.GetGenerateStruct(&models.AgentProfile{}),
		codegen.GetGenerateStruct(&models.AgentTeam{}),
		codegen.GetGenerateStruct(&models.AgentTeamSchedule{}),
		codegen.GetGenerateStruct(&models.AIConfig{}),
		codegen.GetGenerateStruct(&models.SkillDefinition{}),
		codegen.GetGenerateStruct(&models.SkillRunLog{}),
		codegen.GetGenerateStruct(&models.AgentRunLog{}),
		codegen.GetGenerateStruct(&models.SystemConfig{}),
	)

}
