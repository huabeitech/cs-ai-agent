package kb

import (
	"bytes"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/repositories"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mlogclub/simple/sqls"
	"golang.org/x/net/html"
	"gorm.io/gorm"
)

type Chapter struct {
	Title   string
	Link    string
	Content string
}

type InitResult struct {
	DocumentKnowledgeBaseID int64
	FAQKnowledgeBaseID      int64
	TotalChapters           int
	CreatedDocuments        int
	UpdatedDocuments        int
	TotalFAQs               int
	CreatedFAQs             int
	UpdatedFAQs             int
}

type faqSeed struct {
	Question         string
	Answer           string
	SimilarQuestions []string
	Remark           string
}

func Init() (*InitResult, error) {
	chapters, err := read("水浒传")
	if err != nil {
		return nil, fmt.Errorf("read books failed: %w", err)
	}

	if len(chapters) == 0 {
		return nil, fmt.Errorf("no chapters found")
	}

	faqSeeds := knowledgeFAQSeeds()
	result := &InitResult{
		TotalChapters: len(chapters),
		TotalFAQs:     len(faqSeeds),
	}
	err = sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		kbModel, ensureErr := ensureKnowledgeBase(ctx.Tx)
		if ensureErr != nil {
			return ensureErr
		}
		result.DocumentKnowledgeBaseID = kbModel.ID

		for _, chapter := range chapters {
			created, upsertErr := upsertKnowledgeDocument(ctx.Tx, kbModel.ID, chapter)
			if upsertErr != nil {
				return upsertErr
			}
			if created {
				result.CreatedDocuments++
			} else {
				result.UpdatedDocuments++
			}
		}

		faqKnowledgeBase, ensureFAQErr := ensureFAQKnowledgeBase(ctx.Tx)
		if ensureFAQErr != nil {
			return ensureFAQErr
		}
		result.FAQKnowledgeBaseID = faqKnowledgeBase.ID

		for _, faq := range faqSeeds {
			created, upsertErr := upsertKnowledgeFAQ(ctx.Tx, faqKnowledgeBase.ID, faq)
			if upsertErr != nil {
				return upsertErr
			}
			if created {
				result.CreatedFAQs++
			} else {
				result.UpdatedFAQs++
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func ensureKnowledgeBase(db *gorm.DB) (*models.KnowledgeBase, error) {
	now := time.Now()
	item := repositories.KnowledgeBaseRepository.FindOne(db, sqls.NewCnd().Eq("name", "水浒传"))
	if item == nil {
		item = &models.KnowledgeBase{
			Name:                  "水浒传",
			Description:           "四大名著测试数据",
			Status:                enums.StatusOk,
			DefaultTopK:           10,
			DefaultScoreThreshold: 0.2,
			DefaultRerankLimit:    5,
			ChunkProvider:         string(enums.KnowledgeChunkProviderStructured),
			ChunkTargetTokens:     300,
			ChunkMaxTokens:        400,
			ChunkOverlapTokens:    40,
			AnswerMode:            int(enums.KnowledgeAnswerModeStrict),
			FallbackMode:          int(enums.KnowledgeFallbackModeNoAnswer),
			AuditFields: models.AuditFields{
				CreatedAt:      now,
				CreateUserID:   constants.SystemAuditUserID,
				CreateUserName: constants.SystemAuditUserName,
				UpdatedAt:      now,
				UpdateUserID:   constants.SystemAuditUserID,
				UpdateUserName: constants.SystemAuditUserName,
			},
		}
		if err := repositories.KnowledgeBaseRepository.Create(db, item); err != nil {
			return nil, err
		}
		return item, nil
	}

	err := repositories.KnowledgeBaseRepository.Updates(db, item.ID, map[string]any{
		"description":          "四大名著测试数据",
		"status":               enums.StatusOk,
		"chunk_provider":       string(enums.KnowledgeChunkProviderStructured),
		"chunk_target_tokens":  300,
		"chunk_max_tokens":     400,
		"chunk_overlap_tokens": 40,
		"update_user_id":       constants.SystemAuditUserID,
		"update_user_name":     constants.SystemAuditUserName,
		"updated_at":           now,
	})
	if err != nil {
		return nil, err
	}
	return repositories.KnowledgeBaseRepository.Get(db, item.ID), nil
}

func ensureFAQKnowledgeBase(db *gorm.DB) (*models.KnowledgeBase, error) {
	now := time.Now()
	item := repositories.KnowledgeBaseRepository.FindOne(db, sqls.NewCnd().Eq("name", "贝壳客服平台 FAQ"))
	if item == nil {
		item = &models.KnowledgeBase{
			Name:                  "贝壳客服平台 FAQ",
			Description:           "模拟真实客服场景的 FAQ 测试数据，覆盖账号、坐席、机器人、知识库、工单、计费与发票等常见问题。",
			KnowledgeType:         string(enums.KnowledgeBaseTypeFAQ),
			Status:                enums.StatusOk,
			DefaultTopK:           8,
			DefaultScoreThreshold: 0.35,
			DefaultRerankLimit:    5,
			ChunkProvider:         string(enums.KnowledgeChunkProviderFAQ),
			ChunkTargetTokens:     0,
			ChunkMaxTokens:        0,
			ChunkOverlapTokens:    0,
			AnswerMode:            int(enums.KnowledgeAnswerModeStrict),
			FallbackMode:          int(enums.KnowledgeFallbackModeNoAnswer),
			Remark:                "测试数据初始化自动生成",
			AuditFields: models.AuditFields{
				CreatedAt:      now,
				CreateUserID:   constants.SystemAuditUserID,
				CreateUserName: constants.SystemAuditUserName,
				UpdatedAt:      now,
				UpdateUserID:   constants.SystemAuditUserID,
				UpdateUserName: constants.SystemAuditUserName,
			},
		}
		if err := repositories.KnowledgeBaseRepository.Create(db, item); err != nil {
			return nil, err
		}
		return item, nil
	}

	err := repositories.KnowledgeBaseRepository.Updates(db, item.ID, map[string]any{
		"description":             "模拟真实客服场景的 FAQ 测试数据，覆盖账号、坐席、机器人、知识库、工单、计费与发票等常见问题。",
		"knowledge_type":          string(enums.KnowledgeBaseTypeFAQ),
		"status":                  enums.StatusOk,
		"default_top_k":           8,
		"default_score_threshold": 0.35,
		"default_rerank_limit":    5,
		"chunk_provider":          string(enums.KnowledgeChunkProviderFAQ),
		"chunk_target_tokens":     0,
		"chunk_max_tokens":        0,
		"chunk_overlap_tokens":    0,
		"answer_mode":             int(enums.KnowledgeAnswerModeStrict),
		"fallback_mode":           int(enums.KnowledgeFallbackModeNoAnswer),
		"remark":                  "测试数据初始化自动生成",
		"update_user_id":          constants.SystemAuditUserID,
		"update_user_name":        constants.SystemAuditUserName,
		"updated_at":              now,
	})
	if err != nil {
		return nil, err
	}
	return repositories.KnowledgeBaseRepository.Get(db, item.ID), nil
}

func upsertKnowledgeDocument(db *gorm.DB, knowledgeBaseID int64, chapter Chapter) (bool, error) {
	now := time.Now()
	item := repositories.KnowledgeDocumentRepository.FindOne(db, sqls.NewCnd().Eq("knowledge_base_id", knowledgeBaseID).Eq("title", chapter.Title))
	if item == nil {
		item = &models.KnowledgeDocument{
			KnowledgeBaseID: knowledgeBaseID,
			Title:           chapter.Title,
			ContentType:     enums.KnowledgeDocumentContentTypeMarkdown,
			Content:         chapter.Content,
			Status:          enums.StatusOk,
			AuditFields: models.AuditFields{
				CreatedAt:      now,
				CreateUserID:   constants.SystemAuditUserID,
				CreateUserName: constants.SystemAuditUserName,
				UpdatedAt:      now,
				UpdateUserID:   constants.SystemAuditUserID,
				UpdateUserName: constants.SystemAuditUserName,
			},
		}
		if err := repositories.KnowledgeDocumentRepository.Create(db, item); err != nil {
			return false, err
		}
		return true, nil
	}

	err := repositories.KnowledgeDocumentRepository.Updates(db, item.ID, map[string]any{
		"content_type":     enums.KnowledgeDocumentContentTypeMarkdown,
		"content":          chapter.Content,
		"status":           enums.StatusOk,
		"update_user_id":   constants.SystemAuditUserID,
		"update_user_name": constants.SystemAuditUserName,
		"updated_at":       now,
	})
	if err != nil {
		return false, err
	}
	return false, nil
}

func upsertKnowledgeFAQ(db *gorm.DB, knowledgeBaseID int64, seed faqSeed) (bool, error) {
	now := time.Now()
	similarQuestions, err := json.Marshal(seed.SimilarQuestions)
	if err != nil {
		return false, err
	}

	item := repositories.KnowledgeFAQRepository.Find(db, sqls.NewCnd().
		Eq("knowledge_base_id", knowledgeBaseID).
		Eq("question", seed.Question))
	if len(item) == 0 {
		faq := &models.KnowledgeFAQ{
			KnowledgeBaseID:  knowledgeBaseID,
			Question:         seed.Question,
			Answer:           seed.Answer,
			SimilarQuestions: string(similarQuestions),
			Status:           enums.StatusOk,
			IndexStatus:      enums.KnowledgeDocumentIndexStatusPending,
			Remark:           seed.Remark,
			AuditFields: models.AuditFields{
				CreatedAt:      now,
				CreateUserID:   constants.SystemAuditUserID,
				CreateUserName: constants.SystemAuditUserName,
				UpdatedAt:      now,
				UpdateUserID:   constants.SystemAuditUserID,
				UpdateUserName: constants.SystemAuditUserName,
			},
		}
		if err := repositories.KnowledgeFAQRepository.Create(db, faq); err != nil {
			return false, err
		}
		return true, nil
	}

	if err := repositories.KnowledgeFAQRepository.Updates(db, item[0].ID, map[string]any{
		"answer":            seed.Answer,
		"similar_questions": string(similarQuestions),
		"status":            enums.StatusOk,
		"index_status":      enums.KnowledgeDocumentIndexStatusPending,
		"indexed_at":        nil,
		"index_error":       "",
		"remark":            seed.Remark,
		"update_user_id":    constants.SystemAuditUserID,
		"update_user_name":  constants.SystemAuditUserName,
		"updated_at":        now,
	}); err != nil {
		return false, err
	}
	return false, nil
}

func knowledgeFAQSeeds() []faqSeed {
	return []faqSeed{
		{Question: "登录后台时提示账号或密码错误怎么办？", Answer: "请先确认账号是否输入完整，邮箱登录要区分公司域名，密码需注意大小写。如果连续输错 5 次，系统会临时锁定 15 分钟。仍无法登录时，可让企业管理员在“组织设置-成员管理”中重置密码，或通过登录页的“忘记密码”重新设置。", SimilarQuestions: []string{"后台密码错误怎么处理", "账号登录不上去怎么办", "提示用户名或密码错误"}, Remark: "账号登录"},
		{Question: "忘记登录密码后怎么找回？", Answer: "在登录页点击“忘记密码”，输入绑定邮箱后获取验证码即可重置密码。如果账号未绑定邮箱或邮箱无法使用，请联系企业管理员协助重置。为了安全，重置成功后系统会让该账号在所有设备重新登录。", SimilarQuestions: []string{"忘记密码了怎么办", "怎么重置后台密码", "找回登录密码"}, Remark: "账号登录"},
		{Question: "为什么收不到登录验证码邮件？", Answer: "请先检查垃圾邮箱、广告邮件和企业邮箱的安全隔离区。若 5 分钟内仍未收到，建议确认邮箱地址是否填写正确，并联系企业 IT 将平台发信域名加入白名单。如果多次重发都未收到，可能是该邮箱服务商限流，建议改用备用邮箱。", SimilarQuestions: []string{"验证码邮件收不到", "邮箱没有收到验证码", "登录验证码不见了"}, Remark: "账号登录"},
		{Question: "同一个账号可以多人同时登录吗？", Answer: "不建议多人共用同一个坐席账号。平台默认允许同账号在多个设备登录，但会记录登录日志并触发异常提醒。为保证操作留痕、权限隔离和会话分配准确，建议每位坐席使用独立账号。", SimilarQuestions: []string{"一个账号能不能多人共用", "支持多人同时登录同一账号吗", "账号能在多台电脑登录吗"}, Remark: "账号登录"},
		{Question: "新成员加入后如何开通后台账号？", Answer: "企业管理员进入“组织设置-成员管理”，点击“新增成员”，填写姓名、邮箱、所属团队和角色后保存。系统会自动发送激活邮件，成员首次登录时设置密码即可。若你们开通了单点登录，也可以直接从企业身份系统同步成员。", SimilarQuestions: []string{"怎么给新客服开账号", "新增员工账号在哪里", "成员怎么加入后台"}, Remark: "成员管理"},
		{Question: "成员离职后如何停用账号？", Answer: "请在“组织设置-成员管理”中找到对应成员，点击“停用”即可。停用后该账号无法继续登录，但历史会话、工单处理记录和质检数据会保留，不会影响报表统计。若后续确认不再使用，也可以在完成交接后删除账号。", SimilarQuestions: []string{"离职员工账号怎么处理", "怎么禁用成员账号", "停用客服账号"}, Remark: "成员管理"},
		{Question: "角色权限修改后多久生效？", Answer: "角色权限保存后通常即时生效。已在线的成员可能需要刷新页面或重新登录，才能拿到最新权限菜单和接口授权。如果修改后仍能访问原页面，请清理浏览器缓存后再试。", SimilarQuestions: []string{"权限修改什么时候生效", "调整角色后没变化", "角色更新后要重登吗"}, Remark: "成员管理"},
		{Question: "支持企业微信或钉钉单点登录吗？", Answer: "支持对接企业微信、钉钉和标准 SAML/OIDC 单点登录。开通后成员可通过企业身份系统直接登录平台，无需单独维护密码。具体配置需要管理员在“安全设置-单点登录”中填写回调地址、应用凭证并完成测试。", SimilarQuestions: []string{"能接企业微信登录吗", "支持钉钉 SSO 吗", "单点登录怎么接"}, Remark: "成员管理"},
		{Question: "坐席在线、忙碌、离线状态有什么区别？", Answer: "在线表示可正常接待新会话，忙碌表示当前暂不分配新会话但仍可处理已有会话，离线表示不参与会话分配也不接收实时提醒。若开启自动状态切换，长时间无操作或退出登录后，系统会自动变更为离线。", SimilarQuestions: []string{"客服状态怎么理解", "在线忙碌离线区别", "坐席状态说明"}, Remark: "坐席接待"},
		{Question: "会话是怎么分配给坐席的？", Answer: "默认按技能组和轮询策略分配，也可结合坐席当前负载、最近响应时长和优先级规则进行智能分流。若客户命中了指定渠道、语言或标签条件，系统会优先路由到匹配该条件的团队或坐席。", SimilarQuestions: []string{"客户咨询怎么分配", "会话路由规则是什么", "新会话按什么分给客服"}, Remark: "坐席接待"},
		{Question: "如何把会话转接给其他团队？", Answer: "在会话详情页点击“转接”，选择目标团队或指定坐席，并填写转接备注即可。转接后，原坐席仍可在历史记录中查看会话内容，但新消息会优先提醒接收方。若目标团队离线人数过多，建议先确认有人值班。", SimilarQuestions: []string{"会话怎么转给别的组", "转接客服在哪里", "咨询如何分配给其他团队"}, Remark: "坐席接待"},
		{Question: "客户长时间不回复，会话会自动结束吗？", Answer: "可以。在“接待设置-会话规则”里可配置无响应超时时间，例如客户 30 分钟未回复后自动关闭会话。关闭后客户再次发消息会新建会话，历史内容仍可在客户轨迹中查看。", SimilarQuestions: []string{"会话会自动关闭吗", "多久不回复算结束", "无响应自动结束如何设置"}, Remark: "坐席接待"},
		{Question: "坐席最大并发接待数在哪里设置？", Answer: "管理员可在“接待设置-分配规则”中设置团队级默认并发数，也可以在成员详情中为个别坐席单独调整。建议新手坐席从 3 到 5 路并发开始，成熟团队再根据平均处理时长逐步放开。", SimilarQuestions: []string{"并发接待数怎么改", "一个客服最多接多少会话", "坐席并发设置"}, Remark: "坐席接待"},
		{Question: "如何查看客服首响应时长和平均处理时长？", Answer: "进入“数据报表-会话分析”，可按团队、坐席、渠道和日期查看首响应时长、平均响应时长、平均处理时长和超时占比。报表支持导出 CSV，便于做班次复盘或绩效分析。", SimilarQuestions: []string{"首响时长在哪里看", "客服处理时长报表", "怎么查响应效率"}, Remark: "坐席接待"},
		{Question: "客户排队时能看到前面还有多少人吗？", Answer: "支持。你可以在 Web 小部件或 H5 咨询页里开启“排队展示”，客户会看到预估等待人数和等待时长。若不希望暴露具体排队数据，也可以只展示“当前咨询较多，请耐心等待”的文案。", SimilarQuestions: []string{"排队人数能展示吗", "客户能看到等待人数吗", "排队提示怎么设置"}, Remark: "坐席接待"},
		{Question: "机器人什么时候会接管会话？", Answer: "当渠道启用了 AI 机器人优先接待，且用户消息命中机器人的服务时段、语言和业务范围时，会先由机器人回复。若客户主动要求人工、命中转人工规则，或机器人连续多轮未命中知识，系统会自动切换到人工。", SimilarQuestions: []string{"机器人接待规则是什么", "什么情况下先走 AI", "机器人何时转人工"}, Remark: "AI机器人"},
		{Question: "机器人回答不准时应该先调哪里？", Answer: "优先检查知识库是否覆盖了该问题、FAQ 表述是否过于口语化、召回阈值是否设得过高，以及是否开启了错误的知识库。若命中很多但答案仍跑偏，再看模型提示词、回答模式和重排策略。", SimilarQuestions: []string{"AI 回复不准确怎么办", "机器人答非所问怎么调", "先检查哪些配置"}, Remark: "AI机器人"},
		{Question: "怎么限制机器人只回答售后相关问题？", Answer: "可以给该机器人绑定专门的售后知识库，并在意图路由里限定只处理售后渠道、售后入口或带售后标签的会话。同时建议把兜底策略设为“引导转人工”，避免机器人回答超出范围的问题。", SimilarQuestions: []string{"只让机器人回答售后", "机器人回答范围怎么限制", "怎么绑定指定业务"}, Remark: "AI机器人"},
		{Question: "机器人可以识别情绪激动的客户并转人工吗？", Answer: "可以。你可以在策略中配置敏感词、情绪标签和高风险场景，例如投诉、退款争议、辱骂等。一旦命中条件，系统会立即转人工并在会话顶部提示“高风险用户”，方便优先接待。", SimilarQuestions: []string{"负面情绪自动转人工", "投诉用户能自动升级吗", "敏感客户转人工"}, Remark: "AI机器人"},
		{Question: "AI 生成答案时会引用知识来源吗？", Answer: "支持。在开启知识引用后，机器人会在调试信息或内部坐席辅助面板中展示命中的知识条目、相似度分数和引用片段。面向终端用户是否展示引用，可按渠道单独控制。", SimilarQuestions: []string{"AI 能显示知识来源吗", "答案引用在哪里看", "机器人回复能带出处吗"}, Remark: "AI机器人"},
		{Question: "如何测试机器人而不影响线上客户？", Answer: "建议使用“调试问答”或专门的测试渠道。测试渠道可以绑定同一套知识库和机器人配置，但不进入正式报表，也不会分配给真实坐席。上线前至少验证高频问题、异常问法和兜底流程。", SimilarQuestions: []string{"怎么做机器人测试", "测试 AI 会不会影响线上", "有没有沙箱渠道"}, Remark: "AI机器人"},
		{Question: "机器人支持多语言回复吗？", Answer: "支持中文、英文及其他已接入模型支持的语种，但前提是知识库内容本身有对应语言版本。若只上传中文知识，机器人对英文问题通常无法稳定给出英文业务答案，建议按语言拆分知识库。", SimilarQuestions: []string{"AI 支持英文吗", "机器人能多语言吗", "英文客户怎么接待"}, Remark: "AI机器人"},
		{Question: "机器人能读取客户订单信息后再回答吗？", Answer: "可以，但需要先对接业务系统接口，并在机器人工具能力中开放“查询订单”“查询物流”等受控动作。建议只返回回答所需的最小字段，并记录调用日志，避免暴露不必要的隐私信息。", SimilarQuestions: []string{"AI 能查订单吗", "机器人结合业务数据回答", "怎么让机器人调用接口"}, Remark: "AI机器人"},
		{Question: "知识库和 FAQ 知识库有什么区别？", Answer: "文档知识库更适合长文档、帮助中心、制度说明等内容，系统会做分段切片后检索；FAQ 知识库适合标准问答，一问一答更利于高频咨询命中。若内容天然就是固定问法和固定口径，优先用 FAQ。", SimilarQuestions: []string{"FAQ 知识库和文档库区别", "什么场景用 FAQ", "知识库类型怎么选"}, Remark: "知识库"},
		{Question: "上传帮助文档后多久可以被机器人检索到？", Answer: "通常上传成功后会进入索引队列，文本量不大时 1 到 3 分钟内即可完成。若文档很长、图片较多或当前索引任务拥堵，时间会更久。你可以在知识库列表里查看索引状态。", SimilarQuestions: []string{"文档多久生效", "知识库上传后多久可用", "索引需要多久"}, Remark: "知识库"},
		{Question: "FAQ 支持批量导入吗？", Answer: "支持。你可以下载模板后按“问题、答案、相似问、备注”四列填写，再通过 FAQ 页面批量导入。相似问建议用换行或分号分隔，导入前最好先去重，避免同一问题出现多个口径。", SimilarQuestions: []string{"FAQ 能批量上传吗", "FAQ 导入模板在哪", "问答怎么批量导入"}, Remark: "知识库"},
		{Question: "为什么上传的文档显示索引失败？", Answer: "常见原因包括文档为空、正文被复制成图片、内容长度超限、向量服务不可用或模型配置异常。可以先查看索引错误提示，再尝试把文档转换成纯文本或 Markdown 重新上传。", SimilarQuestions: []string{"文档索引失败怎么办", "知识库处理失败原因", "上传后报索引错误"}, Remark: "知识库"},
		{Question: "如何让一个机器人只使用指定知识库？", Answer: "在机器人配置页里把“可用知识库”限制为目标知识库即可。若同时勾选多个知识库，系统会按统一策略检索。为避免跨业务串答，建议按产品线、渠道或场景拆分知识库并独立绑定。", SimilarQuestions: []string{"机器人绑定知识库", "只用一个知识库回答", "知识库范围怎么限制"}, Remark: "知识库"},
		{Question: "知识库内容更新后需要手动重建索引吗？", Answer: "平台默认会在你更新文档或 FAQ 后自动重新索引，不需要手动全量重建。只有在向量配置、分块策略或模型发生较大变化时，才建议执行一次全量重建以保证检索效果一致。", SimilarQuestions: []string{"更新文档后要重建吗", "知识库修改会自动生效吗", "FAQ 改完要重新索引吗"}, Remark: "知识库"},
		{Question: "相似问应该怎么写效果更好？", Answer: "建议保留用户真实提问方式，包括口语、省略句和错别字变体，但不要为了凑数量写过多语义无关的问法。每条 FAQ 一般补 3 到 8 个高频相似问即可，重点覆盖客户常见表达。", SimilarQuestions: []string{"相似问怎么整理", "FAQ 别名怎么写", "类似问题怎么配"}, Remark: "知识库"},
		{Question: "知识库支持权限隔离吗？", Answer: "支持。你可以按团队、角色或机器人配置知识库可见范围。比如售后团队只能看到售后知识库，财务 FAQ 只对财务机器人和财务坐席开放，这样可以减少误检索和越权查看。", SimilarQuestions: []string{"知识库能分权限吗", "不同团队看不同知识", "FAQ 可以隔离吗"}, Remark: "知识库"},
		{Question: "网站咨询按钮怎么嵌入到官网？", Answer: "进入“渠道接入-Web Widget”，复制系统生成的脚本代码，粘贴到官网页面的 `</body>` 前即可。若你们站点启用了 CSP，需要把平台域名加入允许列表，否则组件可能加载失败。", SimilarQuestions: []string{"官网怎么挂咨询入口", "Web Widget 怎么接", "网站客服按钮嵌入"}, Remark: "渠道接入"},
		{Question: "Web Widget 的颜色和文案可以自定义吗？", Answer: "可以。你可以在渠道配置里修改主色、标题、欢迎语、按钮文案、是否展示头像和工作时间提示。保存后前端会在几分钟内刷新缓存，若你希望立即生效，可手动清理页面缓存。", SimilarQuestions: []string{"咨询浮窗能改样式吗", "按钮文案怎么改", "Widget 主题色设置"}, Remark: "渠道接入"},
		{Question: "支持把客服入口嵌到微信 H5 页面吗？", Answer: "支持，但需要使用适配移动端的 H5 咨询页或自定义嵌入页。若在微信内打开，建议同时开启微信浏览器兼容模式，并测试键盘弹起、页面滚动和文件上传权限是否正常。", SimilarQuestions: []string{"H5 页面能接客服吗", "微信里能打开咨询页吗", "移动端客服入口"}, Remark: "渠道接入"},
		{Question: "客户消息提醒可以推送到企业微信吗？", Answer: "支持把新会话、超时未回复、工单升级等提醒推送到企业微信机器人或应用消息。建议只推送关键事件，避免通知过载影响值班人员判断。", SimilarQuestions: []string{"消息提醒发企业微信", "能推送到企微吗", "新会话通知怎么接"}, Remark: "渠道接入"},
		{Question: "访客进入咨询前能先收集手机号吗？", Answer: "可以。你可以在欢迎页开启预采集表单，要求客户填写手机号、订单号、邮箱等信息后再进入会话。这样有助于后续识别身份和分配对应业务团队。", SimilarQuestions: []string{"咨询前收集手机号", "先填表再聊天", "访客信息预采集"}, Remark: "渠道接入"},
		{Question: "支持接入 WhatsApp 或 Telegram 吗？", Answer: "平台可以通过开放接口或第三方集成中间层接入海外渠道，但具体能力取决于你们当前套餐和所选服务商。若是正式商用，建议先确认消息模板、号码资质和当地合规要求。", SimilarQuestions: []string{"能接 WhatsApp 吗", "支持 Telegram 吗", "海外渠道接入"}, Remark: "渠道接入"},
		{Question: "为什么网站上看不到客服浮窗？", Answer: "先检查脚本是否成功加载、站点域名是否在渠道白名单内，以及浏览器是否拦截了第三方脚本。若开启了广告拦截插件或严格 CSP，也可能导致组件被屏蔽。", SimilarQuestions: []string{"网页不显示咨询按钮", "Widget 没出来", "客服浮窗不见了"}, Remark: "渠道接入"},
		{Question: "不同站点可以共用一个客服渠道吗？", Answer: "可以共用，但更建议按站点或品牌拆分渠道，这样可以分别配置欢迎语、机器人、工作时间和报表来源。若多个站点业务差异较大，共用一个渠道会影响会话分流和数据分析。", SimilarQuestions: []string{"多个官网能共用渠道吗", "不同域名用一个 Widget", "站点渠道怎么规划"}, Remark: "渠道接入"},
		{Question: "工单和实时会话有什么关系？", Answer: "实时会话适合即时咨询，工单适合需要跨班次跟进、跨部门协作或需要留痕审批的问题。会话中如果发现问题无法当场解决，可以一键转为工单，并保留原始聊天记录作为上下文。", SimilarQuestions: []string{"为什么还需要工单", "会话和工单区别", "聊天怎么转工单"}, Remark: "工单"},
		{Question: "如何把会话升级成工单？", Answer: "在会话详情页点击“创建工单”，系统会自动带出客户信息、会话摘要和最近消息。你只需补充工单类型、优先级、负责人和期望完成时间即可。", SimilarQuestions: []string{"聊天转工单在哪里", "会话升级工单", "怎么建售后单"}, Remark: "工单"},
		{Question: "工单支持 SLA 超时提醒吗？", Answer: "支持。你可以为不同工单类型配置首次响应时限、处理时限和升级规则，临近超时时会给负责人和主管发送提醒，超时后也可自动升级到上级处理。", SimilarQuestions: []string{"工单超时提醒", "SLA 怎么配置", "工单逾期通知"}, Remark: "工单"},
		{Question: "工单能分配给外部协作人吗？", Answer: "目前标准成员体系主要面向内部账号。如果需要外部协作，可为供应商或合作方单独开受限角色账号，并限制其仅查看被指派工单，避免访问其他客户数据。", SimilarQuestions: []string{"工单给外包处理", "外部人员能看工单吗", "供应商协作权限"}, Remark: "工单"},
		{Question: "工单状态有哪些推荐用法？", Answer: "常见做法是设置为“待受理、处理中、待客户反馈、已解决、已关闭”。其中“待客户反馈”适合需要客户补充材料的场景，“已解决”表示业务已处理完成但仍保留回访窗口。", SimilarQuestions: []string{"工单状态怎么设计", "售后单流程建议", "工单字段如何规划"}, Remark: "工单"},
		{Question: "能否查看工单处理的完整操作记录？", Answer: "可以。每张工单都保留状态变更、指派变更、备注、附件上传和评论记录，方便审计和复盘。管理员还可以导出操作日志做质检或合规留存。", SimilarQuestions: []string{"工单处理日志", "谁改过工单怎么查", "工单历史记录"}, Remark: "工单"},
		{Question: "工单附件支持哪些格式？", Answer: "常见图片、PDF、Excel、Word 和压缩包都支持，单文件大小上限由你们当前存储配置决定。若附件包含客户证件或敏感资料，建议同步开启下载权限控制和水印。", SimilarQuestions: []string{"工单能上传什么文件", "附件格式限制", "售后凭证支持哪些类型"}, Remark: "工单"},
		{Question: "重复提交的工单可以自动合并吗？", Answer: "可以通过规则按手机号、订单号、邮箱或自定义字段检测重复，并提示坐席合并处理。是否自动合并建议谨慎开启，避免把不同问题错误归并到同一张工单。", SimilarQuestions: []string{"重复工单怎么处理", "能自动识别重复吗", "相同订单重复建单"}, Remark: "工单"},
		{Question: "客户信息可以从 CRM 自动同步过来吗？", Answer: "支持通过开放 API、Webhook 或中间件同步客户主数据，例如姓名、手机号、会员等级、所属销售和最近订单。同步后这些字段可以直接在会话侧边栏展示，减少客服来回切系统查询。", SimilarQuestions: []string{"CRM 能同步到客服吗", "客户资料自动带入", "怎么对接用户信息"}, Remark: "集成"},
		{Question: "平台提供开放 API 吗？", Answer: "提供。你可以通过 API 创建会话、发送消息、查询客户、同步工单和拉取报表。正式对接前建议先在测试环境验证签名、限流和错误码处理，再切换到生产。", SimilarQuestions: []string{"有没有开放接口", "客服系统 API 文档", "能程序化调用吗"}, Remark: "集成"},
		{Question: "Webhook 可以推送哪些事件？", Answer: "常见事件包括新会话创建、会话关闭、客户留言、工单创建、工单状态变更、机器人转人工和客户满意度回收等。你可以按需订阅，避免把所有事件都推到业务系统。", SimilarQuestions: []string{"Webhook 支持什么事件", "事件推送列表", "回调通知有哪些"}, Remark: "集成"},
		{Question: "API 调用频率有限制吗？", Answer: "有。默认按应用和接口维度做限流，避免高峰期影响平台稳定。若你们需要批量同步历史数据，建议走离线导入或提前联系技术支持申请更高配额。", SimilarQuestions: []string{"接口限流是多少", "API 有 QPS 限制吗", "批量同步会不会被限流"}, Remark: "集成"},
		{Question: "如何验证开放 API 的签名是否正确？", Answer: "请先确认时间戳、随机串、请求体摘要和签名算法与文档一致。排查时建议先用平台提供的示例请求对比，再检查服务端是否在参与签名的原始字符串里改动了空格、换行或字段顺序。", SimilarQuestions: []string{"API 签名不通过", "签名校验失败怎么办", "接口鉴权报错"}, Remark: "集成"},
		{Question: "可以把会话记录同步到内部 BI 系统吗？", Answer: "可以。你可以通过报表导出、API 增量拉取或消息回调三种方式同步。若是 BI 场景，建议每天离线拉取聚合数据，避免用高频实时接口增加系统压力。", SimilarQuestions: []string{"会话数据怎么同步 BI", "报表能对接数仓吗", "聊天记录导入分析系统"}, Remark: "集成"},
		{Question: "支持和订单系统做查单联动吗？", Answer: "支持。接入后坐席在会话侧边栏即可查看订单状态、支付情况、发货时间和物流单号。若你们希望机器人直接回答订单问题，还需要额外配置工具调用权限。", SimilarQuestions: []string{"能和订单系统打通吗", "聊天页直接查订单", "查单联动怎么做"}, Remark: "集成"},
		{Question: "为什么 API 请求返回 401 或 403？", Answer: "401 一般表示签名错误、时间戳过期或访问令牌无效；403 多数是应用没有该接口权限或 IP 白名单未放通。建议先检查应用状态、签名配置和请求来源 IP。", SimilarQuestions: []string{"接口返回 401 怎么办", "API 403 原因", "鉴权失败怎么排查"}, Remark: "集成"},
		{Question: "聊天记录会保存多久？", Answer: "默认保存期限取决于你们的套餐和合规配置，常见为 180 天到长期保存可选。若业务涉及金融、教育或跨境场景，建议结合行业监管要求单独设置归档周期。", SimilarQuestions: []string{"会话记录保留多久", "聊天消息存多久", "历史记录保存周期"}, Remark: "数据安全"},
		{Question: "导出聊天记录时会包含客户手机号吗？", Answer: "默认按权限控制。普通坐席导出的明细通常会对手机号、邮箱等敏感字段做脱敏，只有拥有高级数据权限的管理员才能导出完整信息。", SimilarQuestions: []string{"导出会脱敏吗", "手机号会不会直接导出", "聊天记录敏感信息处理"}, Remark: "数据安全"},
		{Question: "平台支持数据脱敏吗？", Answer: "支持。你可以对手机号、身份证号、银行卡号、邮箱和地址启用显示脱敏、日志脱敏以及导出脱敏。对于高敏字段，建议同时配置按角色可见范围。", SimilarQuestions: []string{"客户信息能脱敏吗", "隐私字段隐藏", "敏感数据保护"}, Remark: "数据安全"},
		{Question: "是否支持按角色限制查看聊天记录？", Answer: "支持。你可以限制普通坐席只能查看自己接待过的会话，主管查看本团队，管理员查看全局。对于投诉、法务等敏感会话，也可以单独设置更严格的访问范围。", SimilarQuestions: []string{"聊天记录权限隔离", "谁能看全部会话", "会话查看范围"}, Remark: "数据安全"},
		{Question: "客户要求删除个人数据时怎么处理？", Answer: "管理员可在客户资料页发起“数据删除”或“匿名化处理”。系统会按配置清空或打码可识别字段，同时保留必要的审计记录，以满足合规要求和内部追溯。", SimilarQuestions: []string{"用户要求删数据", "隐私删除怎么做", "客户信息匿名化"}, Remark: "数据安全"},
		{Question: "系统有操作日志吗？", Answer: "有。成员登录、权限变更、知识库编辑、工单操作、导出报表等关键动作都会进入审计日志。管理员可以按时间、成员、对象类型筛选并导出。", SimilarQuestions: []string{"后台操作有记录吗", "谁改了配置怎么查", "审计日志在哪里"}, Remark: "数据安全"},
		{Question: "支持设置 IP 白名单吗？", Answer: "支持。你可以在安全设置里为后台登录和开放 API 分别配置 IP 白名单。若你们办公网络经常变动，建议至少给高权限账号启用 MFA，避免完全依赖固定 IP。", SimilarQuestions: []string{"后台能限制 IP 吗", "接口白名单怎么配", "登录来源限制"}, Remark: "数据安全"},
		{Question: "聊天内容会不会被平台拿去训练公共模型？", Answer: "默认不会。客户数据仅用于你们自身的业务处理和已授权的产品功能，不会擅自用于公共模型训练。若你们开通了定制优化服务，也会以合同和配置项约定的数据范围为准。", SimilarQuestions: []string{"聊天数据会训练模型吗", "数据会不会外泄", "平台会拿客户数据训练吗"}, Remark: "数据安全"},
		{Question: "如何给客户打标签？", Answer: "可以在客户详情页手动添加标签，也可以通过规则根据来源渠道、访问页面、下单次数、会员等级或对话关键词自动打标签。标签通常用于分流、营销和服务分层。", SimilarQuestions: []string{"客户标签怎么加", "支持自动标签吗", "用户标签规则"}, Remark: "客户管理"},
		{Question: "客户历史会话在哪里看？", Answer: "打开客户资料页即可看到该客户的历史会话、工单、满意度评价和最近访问记录。若同一个客户用多个渠道接入，建议先配置身份合并规则，避免历史被拆散。", SimilarQuestions: []string{"怎么查客户历史咨询", "用户轨迹在哪里", "以前的聊天记录怎么看"}, Remark: "客户管理"},
		{Question: "一个客户在多个渠道咨询，会被识别成同一个人吗？", Answer: "可以，但需要提前配置统一身份标识，例如手机号、会员 ID、邮箱或外部用户 ID。若不同渠道没有共同标识，系统会默认视为不同访客。", SimilarQuestions: []string{"多渠道客户合并", "同一个人跨渠道识别", "用户身份统一"}, Remark: "客户管理"},
		{Question: "如何筛选高价值客户并优先接待？", Answer: "你可以结合会员等级、近 90 天消费金额、订单频次或 VIP 标签建立高价值客户规则，并在路由策略里设置优先分配到专属团队或高级坐席。", SimilarQuestions: []string{"VIP 客户优先接待", "高价值用户怎么识别", "客户分层服务"}, Remark: "客户管理"},
		{Question: "客户昵称乱码或显示异常怎么办？", Answer: "优先确认上游渠道返回的编码是否为 UTF-8，以及是否包含平台不支持的特殊字符。若只是个别历史数据异常，可通过客户资料页手动更正；若批量异常，建议检查同步接口。", SimilarQuestions: []string{"昵称显示乱码", "客户名称异常", "中文昵称不正常"}, Remark: "客户管理"},
		{Question: "客户满意度评价怎么配置？", Answer: "在“服务评价”里可设置评价时机、问题文案、星级或满意/不满意选项，以及低分自动预警规则。通常建议在会话关闭后立即触发，回收率会更高。", SimilarQuestions: []string{"满意度怎么开", "服务评价配置", "会话结束后评分"}, Remark: "客户管理"},
		{Question: "可以屏蔽骚扰客户吗？", Answer: "支持。你可以把手机号、邮箱、Cookie 或访客 ID 加入黑名单，命中后系统可直接拒绝接入、静默排队，或只允许其提交留言不进入人工接待。", SimilarQuestions: []string{"骚扰用户怎么封禁", "黑名单在哪里设置", "恶意访客屏蔽"}, Remark: "客户管理"},
		{Question: "客户资料支持自定义字段吗？", Answer: "支持。常见如订单等级、合同编号、所属城市、设备型号等都可以配置为自定义字段，并在导入、API 同步和页面展示中统一使用。", SimilarQuestions: []string{"客户字段能自定义吗", "自定义用户属性", "资料页增加业务字段"}, Remark: "客户管理"},
		{Question: "如何设置自动欢迎语？", Answer: "在渠道或机器人配置中启用欢迎语后，可按工作时间、渠道、客户标签或访问页面展示不同文案。建议文案简洁，并明确告知当前是机器人还是人工服务。", SimilarQuestions: []string{"自动回复欢迎语", "开场白在哪里配", "客户进入先发一句"}, Remark: "自动化"},
		{Question: "支持非工作时间自动回复吗？", Answer: "支持。你可以为夜间、节假日或团队休息时段配置离线回复，告知客户预计处理时间、紧急联系电话或留言入口。必要时也可在离线时继续让机器人先接待。", SimilarQuestions: []string{"下班后自动回复", "非工作时间留言", "夜间自动应答"}, Remark: "自动化"},
		{Question: "如何设置关键词自动转人工？", Answer: "在机器人策略里新增转人工规则，输入关键词或短语即可，例如“投诉”“退款”“人工客服”“发票重开”等。建议同时加入同义词和常见口语表达，减少漏判。", SimilarQuestions: []string{"关键词触发人工", "哪些词会转接客服", "自动转人工规则"}, Remark: "自动化"},
		{Question: "会话超时未回复能自动提醒坐席吗？", Answer: "可以。你可以按首响超时、处理中超时和即将 SLA 超时三个阶段配置提醒，支持站内提醒、邮件和企业微信通知。", SimilarQuestions: []string{"超时提醒怎么配", "客服久未回复提醒", "消息超时通知"}, Remark: "自动化"},
		{Question: "能按客户标签分配不同的机器人吗？", Answer: "支持。你可以在路由规则里按客户标签、渠道来源或页面入口命中不同机器人，例如新客走导购机器人，老客走售后机器人。", SimilarQuestions: []string{"不同用户进不同 AI", "按标签分机器人", "机器人路由规则"}, Remark: "自动化"},
		{Question: "能自动给会话生成摘要吗？", Answer: "支持。在开启 AI 摘要后，系统会在会话结束时生成问题摘要、处理结果和待跟进事项，便于转工单、交班和质检。", SimilarQuestions: []string{"聊天自动总结", "会话摘要功能", "交班摘要怎么生成"}, Remark: "自动化"},
		{Question: "自动化规则执行顺序是怎样的？", Answer: "通常按“接入识别 -> 路由分配 -> 机器人应答 -> 转人工/升级 -> 会后自动化”的顺序执行。若多条规则都命中，系统会按优先级和创建顺序决定实际结果。", SimilarQuestions: []string{"规则先后顺序", "自动化命中顺序", "多个规则冲突怎么办"}, Remark: "自动化"},
		{Question: "支持根据访问页面触发不同欢迎语吗？", Answer: "支持。你可以在 Web 渠道里按 URL 路径或页面分组配置欢迎语，比如商品页引导咨询库存，支付页引导咨询优惠和支付问题。", SimilarQuestions: []string{"不同页面不同文案", "页面维度欢迎语", "按 URL 展示话术"}, Remark: "自动化"},
		{Question: "如何查看会话量趋势和高峰时段？", Answer: "在“数据报表-流量分析”里可按小时、日期和渠道查看会话量、访客量、排队峰值和人工接待率。高峰时段建议结合排班数据一起分析。", SimilarQuestions: []string{"会话高峰怎么看", "流量趋势报表", "哪个时间段最忙"}, Remark: "数据报表"},
		{Question: "报表里的首响应时长为什么和人工感觉不一致？", Answer: "因为报表一般按系统时间自动计算，且会区分机器人首响、人工首响和工作时间内首响。建议先确认你看的指标口径，再核对是否排除了离线时段和转接等待时间。", SimilarQuestions: []string{"首响数据不对", "报表口径怎么看", "为什么和实际感受不同"}, Remark: "数据报表"},
		{Question: "机器人命中率和解决率分别是什么意思？", Answer: "命中率指机器人对进入其服务范围的消息中，成功检索并给出答案的比例；解决率指客户在机器人处理后未再转人工且会话完成的比例。命中率高不代表解决率一定高，需要结合看。", SimilarQuestions: []string{"命中率是什么意思", "AI 解决率怎么理解", "机器人指标解释"}, Remark: "数据报表"},
		{Question: "支持按团队导出日报吗？", Answer: "支持。大部分报表都支持按团队、坐席和渠道维度筛选后导出，也可通过定时报表每天自动发送到指定邮箱。", SimilarQuestions: []string{"团队日报导出", "报表自动发邮件", "每天定时发送数据"}, Remark: "数据报表"},
		{Question: "如何看哪些 FAQ 最常被命中？", Answer: "进入知识检索日志或 FAQ 报表，可查看各 FAQ 的命中次数、平均分数和最近命中时间。命中高但满意度低的条目通常最值得优先优化。", SimilarQuestions: []string{"FAQ 命中排行", "哪些知识用得最多", "热门 FAQ 在哪看"}, Remark: "数据报表"},
		{Question: "坐席质检结果可以纳入绩效吗？", Answer: "可以。平台支持按会话抽检结果、服务评价、超时率和处理时长生成坐席评分，但是否直接纳入绩效建议由你们内部制度决定。", SimilarQuestions: []string{"质检分能做绩效吗", "客服评分怎么计算", "质检报表用途"}, Remark: "数据报表"},
		{Question: "导出的报表时间是按哪个时区算的？", Answer: "默认按当前企业设置的时区计算，并以服务器完成汇总的时间窗口为准。若你们团队跨国协作，建议统一企业时区并在报表标题中注明。", SimilarQuestions: []string{"报表时区怎么定", "导出时间不一致", "跨时区数据口径"}, Remark: "数据报表"},
		{Question: "历史报表可以保留多久？", Answer: "报表明细和汇总保留周期取决于套餐和数据保留策略，核心运营报表一般支持长期查看，超大体量的明细日志可能会按月归档。", SimilarQuestions: []string{"报表数据保存多久", "能看多久之前的数据", "历史统计保留"}, Remark: "数据报表"},
		{Question: "如何升级套餐或增加坐席数？", Answer: "联系客户成功经理或在“账单中心-套餐管理”发起升级申请即可。新增坐席一般按剩余计费周期折算费用，升级后权限和资源配额会即时扩容。", SimilarQuestions: []string{"怎么加购买席位", "套餐升级在哪里", "新增坐席费用"}, Remark: "计费发票"},
		{Question: "试用版和正式版有什么主要区别？", Answer: "试用版通常限制坐席数、历史数据保留、开放 API、品牌自定义和高级 AI 能力。正式版支持更完整的权限体系、报表、集成和 SLA 服务。", SimilarQuestions: []string{"试用版限制有哪些", "正式版多了什么", "免费版和付费版区别"}, Remark: "计费发票"},
		{Question: "账单是按自然月还是合同周期结算？", Answer: "标准订阅一般按合同周期结算，可选月付或年付；增购席位和资源通常会按当前合同剩余天数做比例折算。具体以合同和商务报价单为准。", SimilarQuestions: []string{"怎么结算费用", "按月还是按年收费", "增购怎么计费"}, Remark: "计费发票"},
		{Question: "可以开增值税专用发票吗？", Answer: "可以。请在“账单中心-发票信息”中维护完整的开票资料，包括公司名称、税号、开户地址、电话和开户行信息。审核通过后，发票会按账单周期开具。", SimilarQuestions: []string{"能开发票吗", "专票怎么申请", "开票资料在哪里填"}, Remark: "计费发票"},
		{Question: "发票抬头填错了可以重开吗？", Answer: "如发票尚未寄出或尚未作废，可联系商务或财务协助修改；若纸质票已开具且不符合重开条件，需要按财税流程先退回原票再重开。电子发票是否可重开以当前税控规则为准。", SimilarQuestions: []string{"发票信息错了怎么办", "开错票能重开吗", "发票抬头修改"}, Remark: "计费发票"},
		{Question: "欠费后会立刻停服务吗？", Answer: "通常不会立即停服，平台会先进入宽限期并多次提醒管理员续费。宽限期结束后，部分高级能力可能先受限，继续欠费才会影响登录或接待功能。", SimilarQuestions: []string{"欠费多久停用", "没续费会怎样", "账单逾期影响"}, Remark: "计费发票"},
		{Question: "AI 调用量超额后怎么收费？", Answer: "若套餐包含固定 AI 配额，超出部分一般按实际调用量或令牌量计费。建议在账单中心开启用量预警，避免月底集中超额。", SimilarQuestions: []string{"AI 超额怎么计费", "模型调用量收费", "超出配额怎么办"}, Remark: "计费发票"},
		{Question: "如何查看本月已经使用了多少 AI 配额？", Answer: "在“账单中心-用量明细”里可以看到按天汇总的模型调用次数、Token 用量和按机器人/知识库拆分的消耗情况。", SimilarQuestions: []string{"AI 配额怎么看", "本月用了多少 token", "模型用量报表"}, Remark: "计费发票"},
		{Question: "聊天消息发出去但客户收不到怎么办？", Answer: "先确认消息是否发送成功、客户渠道是否在线，以及第三方通道是否有回执失败。若只在某个渠道异常，通常是渠道侧限流、模板审核或网络问题。", SimilarQuestions: []string{"消息发不出去", "客户收不到回复", "发送成功但对方没收到"}, Remark: "故障排查"},
		{Question: "上传文件时报“格式不支持”怎么处理？", Answer: "先核对文件扩展名和实际内容是否一致，例如部分截图工具导出的文件后缀和真实格式不匹配。若格式本身受支持，建议检查文件是否损坏或超出大小限制。", SimilarQuestions: []string{"附件上传失败", "文件格式不支持", "为什么不能传图片"}, Remark: "故障排查"},
		{Question: "页面很卡或者消息列表加载很慢是什么原因？", Answer: "常见原因包括浏览器缓存过多、同时打开过多会话标签、网络延迟较高或页面中存在超大附件。建议先刷新页面、关闭不必要标签页，并确认本地网络和浏览器版本。", SimilarQuestions: []string{"后台卡顿怎么办", "消息加载慢", "聊天页很慢"}, Remark: "故障排查"},
		{Question: "为什么刚配置好的机器人没有生效？", Answer: "通常需要确认机器人是否绑定到了正确渠道、服务时间是否覆盖当前时段，以及该渠道是否仍在走人工优先。若配置无误仍未生效，可查看调试日志确认是否命中了路由条件。", SimilarQuestions: []string{"机器人不工作", "AI 没有接管会话", "配置完没效果"}, Remark: "故障排查"},
		{Question: "知识库明明有答案，机器人还是说不知道，怎么办？", Answer: "先看该问题是否命中了正确知识库，其次检查 FAQ 问法和用户真实提问是否差距过大，再检查阈值是否过高导致被过滤。必要时补充相似问，或降低最低相似度阈值后重新验证。", SimilarQuestions: []string{"知识命不中怎么办", "FAQ 有内容却答不出", "AI 说查不到答案"}, Remark: "故障排查"},
		{Question: "浏览器提示证书风险，平台打不开怎么办？", Answer: "请先确认访问的域名是否正确、企业网络是否做了 HTTPS 解密，以及本机时间是否准确。若是自定义域名场景，还要检查证书是否已过期或未正确绑定。", SimilarQuestions: []string{"证书错误怎么处理", "网站打不开提示不安全", "HTTPS 有问题"}, Remark: "故障排查"},
		{Question: "为什么导出的 CSV 打开后中文乱码？", Answer: "通常是因为本地表格软件默认不是 UTF-8 编码。建议优先用支持 UTF-8 的方式导入，或在导出后用支持编码选择的工具打开。", SimilarQuestions: []string{"CSV 中文乱码", "导出文件打不开中文", "报表编码问题"}, Remark: "故障排查"},
		{Question: "系统升级会影响线上接待吗？", Answer: "常规小版本升级一般采用滚动发布，不会中断在线接待；涉及数据库结构或底层组件的大版本升级，平台会提前通知维护窗口，并建议你们避开业务高峰。", SimilarQuestions: []string{"升级会停机吗", "发版影响线上吗", "系统维护通知"}, Remark: "故障排查"},
		{Question: "支持自定义品牌 Logo 和登录页吗？", Answer: "支持自定义企业 Logo、登录页欢迎语和部分品牌色，高级套餐还支持自定义域名和白标展示。具体可配置项取决于你们的套餐级别。", SimilarQuestions: []string{"能换成自己品牌吗", "登录页能自定义吗", "支持白标吗"}, Remark: "品牌配置"},
		{Question: "可以按品牌线配置不同客服团队吗？", Answer: "可以。很多企业会按品牌、区域或产品线拆分团队和渠道，再为每个团队配置独立机器人、工作时间和报表视图。", SimilarQuestions: []string{"多品牌怎么管理", "不同业务线分团队", "品牌独立配置"}, Remark: "品牌配置"},
		{Question: "自定义域名需要准备什么？", Answer: "通常需要你们提供已备案域名、DNS 管理权限以及用于回源或 CNAME 的配置能力。若开启 HTTPS，还需要按指引完成证书校验或授权托管证书。", SimilarQuestions: []string{"绑定自定义域名", "专属域名怎么配", "客服系统独立域名"}, Remark: "品牌配置"},
		{Question: "Logo 替换后前端多久刷新？", Answer: "通常几分钟内会生效，具体取决于 CDN 缓存时间。若你在后台已经看到新 Logo，但前台仍未更新，建议清空浏览器缓存或稍后再试。", SimilarQuestions: []string{"换 Logo 后没生效", "品牌图标多久更新", "前端缓存多久"}, Remark: "品牌配置"},
		{Question: "支持多组织或多租户统一管理吗？", Answer: "支持企业下管理多个组织，但权限和数据隔离方式需按实际业务设计。若是完全独立运营的品牌或国家站点，通常建议拆成独立组织。", SimilarQuestions: []string{"多租户支持吗", "多个子公司统一管", "多组织架构"}, Remark: "品牌配置"},
		{Question: "如何申请产品培训或上线辅导？", Answer: "你可以联系客户成功经理预约标准培训、管理员培训或机器人调优辅导。首次上线建议安排一次管理员培训和一次一线坐席培训，能明显减少上线初期问题。", SimilarQuestions: []string{"有没有培训服务", "上线辅导怎么预约", "员工使用培训"}, Remark: "客户成功"},
		{Question: "遇到紧急故障，最快如何联系支持团队？", Answer: "若购买了企业服务，可通过专属工单通道、服务群或紧急支持电话联系。提交时请尽量附上问题时间、组织 ID、影响范围、截图和复现步骤，便于快速定位。", SimilarQuestions: []string{"紧急问题联系谁", "系统故障怎么报", "售后支持入口"}, Remark: "客户成功"},
		{Question: "产品更新公告在哪里看？", Answer: "你可以在后台首页公告栏、帮助中心更新日志或服务群中查看版本发布说明。涉及影响配置或操作习惯的变更，平台一般会提前通知。", SimilarQuestions: []string{"版本更新在哪里看", "发布说明入口", "新功能公告"}, Remark: "客户成功"},
		{Question: "能提供上线前的最佳实践建议吗？", Answer: "可以。标准建议包括先梳理高频问题 FAQ、配置清晰的转人工策略、按业务拆分知识库、先从一个渠道灰度上线，再逐步扩展到全部渠道。", SimilarQuestions: []string{"上线前准备什么", "机器人落地建议", "客服系统实施建议"}, Remark: "客户成功"},
		{Question: "平台支持数据迁移服务吗？", Answer: "支持按项目评估。常见迁移内容包括历史客户资料、会话记录、FAQ、工单和成员账号。由于不同系统字段差异较大，迁移前通常需要做一次字段映射确认。", SimilarQuestions: []string{"从旧系统迁移数据", "历史消息能导入吗", "数据迁移服务"}, Remark: "客户成功"},
		{Question: "如何判断当前 FAQ 是否需要优化？", Answer: "可以优先看三类信号：命中高但转人工率高、命中高但满意度低、以及客户经常追问同一问题。出现这些情况时，通常说明答案不够完整、口径不一致，或相似问覆盖不够。", SimilarQuestions: []string{"FAQ 优化依据", "哪些问答该先改", "知识库效果怎么评估"}, Remark: "知识运营"},
		{Question: "FAQ 的答案建议写多长？", Answer: "建议先给出结论，再补充步骤和注意事项。大多数客服 FAQ 控制在 80 到 220 字效果较好，太短容易信息不全，太长又不利于机器人稳定引用和客户快速阅读。", SimilarQuestions: []string{"FAQ 答案长度建议", "回答写多长合适", "问答内容怎么控制"}, Remark: "知识运营"},
		{Question: "一个问题有多个业务口径，FAQ 应该怎么处理？", Answer: "不要把多个冲突口径塞进同一条 FAQ。更合理的做法是按前置条件拆分，比如“个人版如何退款”和“企业版如何退款”分别建条目，并在答案开头明确适用范围。", SimilarQuestions: []string{"FAQ 口径冲突怎么办", "同一问题多个答案", "知识条目怎么拆"}, Remark: "知识运营"},
		{Question: "FAQ 更新后如何通知坐席口径变化？", Answer: "建议结合版本说明、站内公告和班前会同步。对核心条目可以开启“重点知识变更提醒”，让坐席在登录后先看到更新摘要。", SimilarQuestions: []string{"知识更新怎么通知客服", "FAQ 改动如何同步", "口径变更提醒"}, Remark: "知识运营"},
		{Question: "谁适合负责 FAQ 日常维护？", Answer: "通常由业务负责人提供口径，知识运营或客服主管负责整理成标准 FAQ，再由一线坐席持续反馈命中和遗漏情况。最好明确唯一责任人，避免长期无人维护。", SimilarQuestions: []string{"FAQ 谁来维护", "知识库负责人怎么定", "问答内容归谁管"}, Remark: "知识运营"},
		{Question: "知识库应该多久复盘一次？", Answer: "高频业务建议每周看一次命中和转人工数据，每月至少做一次系统复盘。活动期、发版期和政策调整期需要更高频地更新知识。", SimilarQuestions: []string{"FAQ 多久更新一次", "知识运营复盘频率", "什么时候维护知识库"}, Remark: "知识运营"},
		{Question: "如何区分应该写 FAQ 还是快捷回复？", Answer: "固定且经常重复的问题，优先写 FAQ 供机器人和检索使用；只在人工接待时需要快速发送、但未必适合机器人直接回答的内容，更适合放快捷回复。", SimilarQuestions: []string{"FAQ 和快捷回复区别", "什么内容放快捷回复", "知识条目怎么分类"}, Remark: "知识运营"},
		{Question: "低频但高风险的问题也要录入 FAQ 吗？", Answer: "建议录入，但要在答案中明确风险提示和升级路径，比如退款争议、隐私投诉、法务取证等。即便命中率不高，这类问题也能显著降低误答风险。", SimilarQuestions: []string{"低频问题有必要建 FAQ 吗", "高风险问答怎么处理", "投诉类知识要不要建"}, Remark: "知识运营"},
	}
}

func read(name string) (chapters []Chapter, err error) {
	path := filepath.Join("cmd", "testdata", "kb", name+".html")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	doc, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	baseDir := filepath.Dir(path)
	seen := make(map[string]struct{})

	var walk func(*html.Node) error
	walk = func(n *html.Node) error {
		if n.Type == html.ElementNode && n.Data == "a" {
			parent := n.Parent
			if parent != nil && parent.Type == html.ElementNode && parent.Data == "span" && parent.Parent != nil && hasClass(parent.Parent, "chapter") {
				href := strings.TrimSpace(getAttr(n, "href"))
				if href != "" {
					resolved := filepath.Clean(filepath.Join(baseDir, href))
					if _, ok := seen[resolved]; !ok {
						seen[resolved] = struct{}{}

						title := cleanText(nodeText(parent))
						title = strings.ReplaceAll(title, "（原文）", "")
						title = strings.TrimSpace(title)

						content, readErr := readChapterContent(resolved)
						if readErr != nil {
							return readErr
						}

						chapters = append(chapters, Chapter{
							Title:   title,
							Link:    resolved,
							Content: content,
						})
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := walk(c); err != nil {
				return err
			}
		}
		return nil
	}

	err = walk(doc)
	if err != nil {
		return nil, err
	}

	return chapters, nil
}

func readChapterContent(link string) (content string, err error) {
	data, err := os.ReadFile(link)
	if err != nil {
		return "", err
	}

	doc, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	lines := make([]string, 0, 128)
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "h1" {
				title := cleanText(nodeText(n))
				title = strings.ReplaceAll(title, " 原文", "")
				title = strings.TrimSpace(title)
				if title != "" {
					lines = append(lines, title)
				}
			}

			if n.Data == "p" {
				if hasClass(n, "next") || getAttr(n, "id") == "home" || getAttr(n, "id") == "list" {
					return
				}
				if isInsideClass(n, "pn") {
					return
				}
				line := cleanText(nodeText(n))
				if line != "" {
					lines = append(lines, line)
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(doc)
	return strings.Join(lines, "\n"), nil
}

func getAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func hasClass(n *html.Node, className string) bool {
	classes := strings.Fields(getAttr(n, "class"))
	for _, c := range classes {
		if c == className {
			return true
		}
	}
	return false
}

func isInsideClass(n *html.Node, className string) bool {
	for p := n.Parent; p != nil; p = p.Parent {
		if hasClass(p, className) {
			return true
		}
	}
	return false
}

func nodeText(n *html.Node) string {
	if n == nil {
		return ""
	}

	var b strings.Builder
	var walk func(*html.Node)
	walk = func(cur *html.Node) {
		if cur.Type == html.TextNode {
			b.WriteString(cur.Data)
		}
		for c := cur.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(n)
	return b.String()
}

func cleanText(s string) string {
	s = strings.ReplaceAll(s, "\u00a0", " ")
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	return strings.Join(strings.Fields(s), " ")
}
