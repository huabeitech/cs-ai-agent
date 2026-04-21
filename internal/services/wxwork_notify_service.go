package services

import (
	"fmt"
	"strings"

	"cs-agent/internal/pkg/config"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/repositories"
	"cs-agent/internal/wxwork"

	"github.com/mlogclub/simple/sqls"
	wxmessage "github.com/silenceper/wechat/v2/work/message"
)

var WxWorkNotifyService = newWxWorkNotifyService()

type wxWorkMessageSender interface {
	SendText(request wxmessage.SendTextRequest) (*wxmessage.SendResponse, error)
}

type wxWorkNotifyRecipients struct {
	ToUsers   []string
	ToParties []string
	ToTags    []string
}

type wxWorkNotifyService struct {
	senderFactory func() (wxWorkMessageSender, error)
}

func newWxWorkNotifyService() *wxWorkNotifyService {
	return &wxWorkNotifyService{
		senderFactory: func() (wxWorkMessageSender, error) {
			if !wxwork.Enabled() || wxwork.GetWorkCli() == nil {
				return nil, fmt.Errorf("wxwork is not enabled")
			}
			return wxwork.GetWorkCli().GetMessage(), nil
		},
	}
}

func (s *wxWorkNotifyService) Enabled() bool {
	if !wxwork.Enabled() {
		return false
	}
	return config.Current().WxWork.Notify.Enabled
}

func (s *wxWorkNotifyService) SendTextToAssigneeOrDefault(assigneeID int64, title, body string) error {
	if !s.Enabled() {
		return nil
	}
	recipients := s.resolveRecipientsByUserIDs([]int64{assigneeID})
	if recipients.empty() {
		recipients = s.defaultRecipients()
	}
	if recipients.empty() {
		return nil
	}
	return s.sendText(title, body, recipients)
}

func (s *wxWorkNotifyService) sendText(title, body string, recipients wxWorkNotifyRecipients) error {
	if !s.Enabled() {
		return nil
	}
	content := s.buildTextContent(title, body)
	if content == "" {
		return nil
	}
	sender, err := s.senderFactory()
	if err != nil {
		return err
	}
	cfg := config.Current().WxWork
	req := wxmessage.SendTextRequest{
		SendRequestCommon: &wxmessage.SendRequestCommon{
			ToUser:                 strings.Join(recipients.ToUsers, "|"),
			ToParty:                strings.Join(recipients.ToParties, "|"),
			ToTag:                  strings.Join(recipients.ToTags, "|"),
			AgentID:                strings.TrimSpace(cfg.AgentID),
			Safe:                   boolToInt(cfg.Notify.Safe),
			EnableDuplicateCheck:   boolToInt(cfg.Notify.EnableDuplicateCheck),
			DuplicateCheckInterval: s.normalizeDuplicateCheckInterval(cfg.Notify.DuplicateCheckInterval),
		},
		Text: wxmessage.TextField{Content: content},
	}
	_, err = sender.SendText(req)
	return err
}

func (s *wxWorkNotifyService) resolveRecipientsByUserIDs(userIDs []int64) wxWorkNotifyRecipients {
	userIDs = uniqueInt64s(userIDs)
	if len(userIDs) == 0 {
		return wxWorkNotifyRecipients{}
	}
	cfg := config.Current().WxWork
	identities := repositories.UserIdentityRepository.Find(sqls.DB(), sqls.NewCnd().
		Eq("provider", enums.ThirdProviderWxWork).
		Eq("provider_corp_id", strings.TrimSpace(cfg.CorpID)).
		Eq("status", enums.StatusOk).
		In("user_id", userIDs).
		Asc("id"))
	recipients := wxWorkNotifyRecipients{}
	for i := range identities {
		if receiver := strings.TrimSpace(identities[i].ProviderUserID); receiver != "" {
			recipients.ToUsers = append(recipients.ToUsers, receiver)
		}
	}
	recipients.ToUsers = uniqueStrings(recipients.ToUsers)
	return recipients
}

func (s *wxWorkNotifyService) defaultRecipients() wxWorkNotifyRecipients {
	cfg := config.Current().WxWork.Notify
	return wxWorkNotifyRecipients{
		ToUsers:   uniqueStrings(cfg.ToUsers),
		ToParties: uniqueStrings(cfg.ToParties),
		ToTags:    uniqueStrings(cfg.ToTags),
	}
}

func (s *wxWorkNotifyService) buildTextContent(title, body string) string {
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)
	switch {
	case title == "" && body == "":
		return ""
	case title == "":
		return truncateRunes(body, 1024)
	case body == "":
		return truncateRunes(title, 1024)
	default:
		return truncateRunes(title+"\n\n"+body, 1024)
	}
}

func (s *wxWorkNotifyService) normalizeDuplicateCheckInterval(value int) int {
	if value <= 0 {
		return 1800
	}
	if value > 14400 {
		return 14400
	}
	return value
}

func (r wxWorkNotifyRecipients) empty() bool {
	return len(r.ToUsers) == 0 && len(r.ToParties) == 0 && len(r.ToTags) == 0
}

func uniqueStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	ret := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		item := strings.TrimSpace(value)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		ret = append(ret, item)
	}
	return ret
}

func uniqueInt64s(values []int64) []int64 {
	if len(values) == 0 {
		return nil
	}
	ret := make([]int64, 0, len(values))
	seen := make(map[int64]struct{}, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		ret = append(ret, value)
	}
	return ret
}

func truncateRunes(value string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= max {
		return string(runes)
	}
	return string(runes[:max])
}

func defaultIfBlank(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return value
	}
	return strings.TrimSpace(fallback)
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
