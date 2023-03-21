package my_webhooks

import (
	"github.com/MythicMeta/MythicContainer/webhookstructs"
)

func Initialize() {
	myWebhooks := webhookstructs.WebhookDefinition{
		NewFeedbackFunction: newfeedbackWebhook,
		NewCallbackFunction: newCallbackWebhook,
		NewStartupFunction:  newStartupMessage,
	}
	webhookstructs.AllWebhookData.Get("my_webhooks").AddWebhookDefinition(myWebhooks)
}
