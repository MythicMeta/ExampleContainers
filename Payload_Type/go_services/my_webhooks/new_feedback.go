package my_webhooks

import (
	"fmt"
	"github.com/MythicMeta/MythicContainer/logging"
	"github.com/MythicMeta/MythicContainer/mythicrpc"
	"github.com/MythicMeta/MythicContainer/webhookstructs"
)

func newfeedbackWebhook(input webhookstructs.NewFeedbackWebookMessage) {
	newMessage := webhookstructs.GetNewDefaultWebhookMessage()
	newMessage.Channel = webhookstructs.AllWebhookData.Get("my_webhooks").GetWebhookChannel(input, webhookstructs.WEBHOOK_TYPE_NEW_FEEDBACK)
	var webhookURL = webhookstructs.AllWebhookData.Get("my_webhooks").GetWebhookURL(input, webhookstructs.WEBHOOK_TYPE_NEW_FEEDBACK)
	if webhookURL == "" {
		logging.LogError(nil, "No webhook url specified for operation or locally")
		go mythicrpc.SendMythicRPCOperationEventLogCreate(mythicrpc.MythicRPCOperationEventLogCreateMessage{
			OperationId:  &input.OperationID,
			Message:      "No webhook url specified, can't send webhook message",
			MessageLevel: mythicrpc.MESSAGE_LEVEL_WARNING,
		})
		return
	}
	switch input.Data.FeedbackType {
	case "bug":
		newMessage.Attachments[0].Title = "Bug Report!"
		newMessage.Attachments[0].Color = "#ff0000"
		if newMessage.Attachments[0].Blocks != nil {
			(*newMessage.Attachments[0].Blocks)[0].Text.Text = fmt.Sprintf("<!here> *%s* submitted a bug report! :bug:", input.OperatorUsername)
		}

	case "feature_request":
		newMessage.Attachments[0].Title = "Feature Request!"
		newMessage.Attachments[0].Color = "#00cc00"
		if newMessage.Attachments[0].Blocks != nil {
			(*newMessage.Attachments[0].Blocks)[0].Text.Text = fmt.Sprintf("*%s* submitted a feature request! :new:", input.OperatorUsername)
		}
	case "confusing_ui":
		newMessage.Attachments[0].Title = "Confusing UI!"
		newMessage.Attachments[0].Color = "#ff9900"
		if newMessage.Attachments[0].Blocks != nil {
			(*newMessage.Attachments[0].Blocks)[0].Text.Text = fmt.Sprintf("*%s* has issues with the UI! :interrobang:", input.OperatorUsername)
		}
	case "detection":
		newMessage.Attachments[0].Title = "We got caught! :bomb:"
		newMessage.Attachments[0].Color = "#ff0000"
		if newMessage.Attachments[0].Blocks != nil {
			(*newMessage.Attachments[0].Blocks)[0].Text.Text = fmt.Sprintf("<!here> *%s* noticed we were detected! :bomb:", input.OperatorUsername)
		}
	default:
		newMessage.Attachments[0].Title = "Unknown Type"
		newMessage.Attachments[0].Color = "#ffff00"
		if newMessage.Attachments[0].Blocks != nil {
			(*newMessage.Attachments[0].Blocks)[0].Text.Text = fmt.Sprintf("*%s* submitted an unknown type: %s", input.OperatorUsername, input.Data.FeedbackType)
		}
	}
	// construct the fields list
	fieldsBlockStarter := []webhookstructs.SlackWebhookMessageAttachmentBlockText{
		{
			Type: "mrkdwn",
			Text: fmt.Sprintf("*Operation*\n%s", input.OperationName),
		},
	}
	if input.Data.TaskID != nil {

		if TaskData, err := mythicrpc.SendMythicRPCTaskSearch(mythicrpc.MythicRPCTaskSearchMessage{
			TaskID: *input.Data.TaskID,
		}); err != nil {
			logging.LogError(err, "Failed to fetch task information")
		} else {
			fieldsBlockStarter = append(fieldsBlockStarter,
				webhookstructs.SlackWebhookMessageAttachmentBlockText{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*Callback / Task*\n%d / %d", TaskData.Tasks[0].CallbackID, *input.Data.TaskID),
				})
			fieldsBlockStarter = append(fieldsBlockStarter,
				webhookstructs.SlackWebhookMessageAttachmentBlockText{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*Command*\n%s", TaskData.Tasks[0].CommandName),
				})
			fieldsBlockStarter = append(fieldsBlockStarter,
				webhookstructs.SlackWebhookMessageAttachmentBlockText{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*Status*\n%s", TaskData.Tasks[0].Status),
				})
		}
	}
	fieldBlock := webhookstructs.SlackWebhookMessageAttachmentBlock{
		Type:   "section",
		Fields: &fieldsBlockStarter,
	}
	messageBlockText := webhookstructs.SlackWebhookMessageAttachmentBlockText{
		Type: "mrkdwn",
		Text: fmt.Sprintf("%s", input.Data.Message),
	}
	messageBlock := webhookstructs.SlackWebhookMessageAttachmentBlock{
		Type: "section",
		Text: &messageBlockText,
	}
	dividerBlock := webhookstructs.SlackWebhookMessageAttachmentBlock{
		Type: "divider",
	}
	// add the block to the blocks list
	tempBlockList := append(*(newMessage.Attachments[0].Blocks), fieldBlock, dividerBlock, messageBlock)
	newMessage.Attachments[0].Blocks = &tempBlockList
	// now actually send the message
	/*
		logging.LogDebug("webhook about to fire", "url", webhookURL, "message", newMessage)
		messageBytes, _ := json.MarshalIndent(newMessage, "", "  ")
		fmt.Printf("%s", string(messageBytes))

	*/

	webhookstructs.SubmitWebRequest("POST", webhookURL, newMessage)
}
