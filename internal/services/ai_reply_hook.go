package services

import "cs-agent/internal/models"

var TriggerAIReplyAsyncHook func(conversation models.Conversation, message models.Message)
