from mythic_container.WebhookBase import *
from mythic_container.MythicGoRPC.send_mythic_rpc_task_search import *


class MyWebhook(Webhook):

    async def new_startup(self, inputMsg: WebhookMessage) -> None:
        message = {
            "channel": f"#{self.getWebhookChannel(inputMsg=inputMsg)}",
            "username": "Mythic",
            "icon_emoji": ":mythic:",
            "attachments": [
                {
                    "fallback": "Mythic Webhook Started!",
                    "color": "#ff0000",
                    "blocks": [
                        {
                            "type": "section",
                            "text": {
                                "type": "mrkdwn",
                                "text": "<!here> Mythic Started!"
                            }
                        },
                        {
                            "type": "divider"
                        },
                        {
                            "type": "section",
                            "fields": [
                                {
                                    "type": "mrkdwn",
                                    "text": f"Mythic Online for operation {inputMsg.OperationName}!"
                                }
                            ]
                        }
                    ]
                }
            ]
        }
        await sendWebhookMessage(contents=message, url=self.getWebhookURL(inputMsg=inputMsg))

    async def new_callback(self, inputMsg: WebhookMessage) -> None:
        integrityLevelString = "MEDIUM"
        if inputMsg.Data.IntegrityLevel < 2:
            integrityLevelString = "LOW"
        elif inputMsg.Data.IntegrityLevel == 3:
            integrityLevelString = "HIGH"
        elif inputMsg.Data.IntegrityLevel > 3:
            integrityLevelString = "SYSTEM"
        message = {
            "channel": f"#{self.getWebhookChannel(inputMsg=inputMsg)}",
            "username": "Mythic",
            "icon_emoji": ":mythic:",
            "attachments": [
                {
                    "fallback": "New Callback!",
                    "color": "#b366ff",
                    "blocks": [
                        {
                            "type": "section",
                            "text": {
                                "type": "mrkdwn",
                                "text": "<!here> You have a new callback!"
                            }
                        },
                        {
                            "type": "divider"
                        },
                        {
                            "type": "section",
                            "fields": [
                                {
                                    "type": "mrkdwn",
                                    "text": f"*Operation*\n{inputMsg.OperationName}"
                                },
                                {
                                    "type": "mrkdwn",
                                    "text": f"*Callback ID*\n{inputMsg.Data.DisplayID}"
                                },
                                {
                                    "type": "mrkdwn",
                                    "text": f"*Integrity Level*\n{integrityLevelString}"
                                },
                                {
                                    "type": "mrkdwn",
                                    "text": f"*IP*\n{inputMsg.Data.IPs}"
                                },
                                {
                                    "type": "mrkdwn",
                                    "text": f"*Type*\n{inputMsg.Data.PayloadType}"
                                }
                            ]
                        },
                        {
                            "type": "divider"
                        },
                        {
                            "type": "section",
                            "text": {
                                "type": "mrkdwn",
                                "text": f"{inputMsg.Data.Description}"
                            }
                        }
                    ]
                }
            ]
        }
        await sendWebhookMessage(contents=message, url=self.getWebhookURL(inputMsg=inputMsg))

    async def new_feedback(self, inputMsg: WebhookMessage) -> None:
        feedbackMsgType = "Unknown Type"
        feedbackMsgColor = "#ffff00"
        feedbackMsgTitle = f"*{inputMsg.OperatorUsername}* submitted an unknown type: {inputMsg.Data.FeedbackType}"
        if inputMsg.Data.FeedbackType == "bug":
            feedbackMsgType = "Bug Report!"
            feedbackMsgColor = "#ff0000"
            feedbackMsgTitle = f"<!here> *{inputMsg.OperatorUsername}* submitted a bug report! :bug:"
        elif inputMsg.Data.FeedbackType == "feature_request":
            feedbackMsgType = "Feature Request!"
            feedbackMsgColor = "#00cc00"
            feedbackMsgTitle = f"*{inputMsg.OperatorUsername}* submitted a feature request! :new:"
        elif inputMsg.Data.FeedbackType == "confusing_ui":
            feedbackMsgType = "Confusing UI!"
            feedbackMsgColor = "#ff9900"
            feedbackMsgTitle = f"*{inputMsg.OperatorUsername}* has issues with the UI! :interrobang:"
        elif inputMsg.Data.FeedbackType == "detection":
            feedbackMsgType = "We got caught! :bomb:"
            feedbackMsgColor = "#ff0000"
            feedbackMsgTitle = f"<!here> *{inputMsg.OperatorUsername}* noticed we were detected! :bomb:"

        msgBlocks = [
            {
                "type": "section",
                "text": {
                    "type": "mrkdwn",
                    "text": f"{feedbackMsgTitle}"
                }
            },
            {
                "type": "divider"
            },
        ]
        if inputMsg.Data.TaskID is not None:
            taskDataResponse = await SendMythicRPCTaskSearch(MythicRPCTaskSearchMessage(TaskID=inputMsg.Data.TaskID))
            if taskDataResponse.Success and len(taskDataResponse.Tasks) == 1:
                newBlock = {
                    "type": "section",
                    "fields": [
                        {
                            "type": "mrkdwn",
                            "text": f"*Operation*\n{inputMsg.OperationName}"
                        },
                        {
                            "type": "mrkdwn",
                            "text": f"*Callback / Task*\n{taskDataResponse.Tasks[0].CallbackID} / {taskDataResponse.Tasks[0].DisplayID}"
                        },
                        {
                            "type": "mrkdwn",
                            "text": f"*Command*\n{taskDataResponse.Tasks[0].CommandName}"
                        },
                        {
                            "type": "mrkdwn",
                            "text": f"*Status*\n{taskDataResponse.Tasks[0].Status}"
                        }
                    ]
                }
                msgBlocks.append(newBlock)
            else:
                logger.error(f"Failed to search for task data: {taskDataResponse.Error}")
        else:
            newBlock = {
                "type": "section",
                "fields": [
                    {
                        "type": "mrkdwn",
                        "text": f"*Operation*\n{inputMsg.OperationName}"
                    },
                ]
            }
            msgBlocks.append(newBlock)
        msgBlocks.append({
            "type": "divider"
        })
        msgBlocks.append(
            {
                "type": "section",
                "text": {
                    "type": "mrkdwn",
                    "text": f"{inputMsg.Data.Message}"
                }
            })
        message = {
            "channel": f"#{self.getWebhookChannel(inputMsg=inputMsg)}",
            "username": "Mythic",
            "icon_emoji": ":mythic:",
            "attachments": [
                {
                    "fallback": f"{feedbackMsgType}",
                    "color": f"{feedbackMsgColor}",
                    "blocks": msgBlocks
                }
            ]
        }

        await sendWebhookMessage(contents=message, url=self.getWebhookURL(inputMsg=inputMsg))

    async def new_alert(self, inputMsg: WebhookMessage) -> None:
        message = {
            "channel": f"#{self.getWebhookChannel(inputMsg=inputMsg)}",
            "username": "Mythic",
            "icon_emoji": ":mythic:",
            "attachments": [
                {
                    "fallback": "New Event Alert!",
                    "color": "#ff0000",
                    "blocks": [
                        {
                            "type": "section",
                            "text": {
                                "type": "mrkdwn",
                                "text": f"Source: {inputMsg.Data.Source}"
                            }
                        },
                        {
                            "type": "divider"
                        },
                        {
                            "type": "section",
                            "fields": [
                                {
                                    "type": "mrkdwn",
                                    "text": f"{inputMsg.Data.Message}!"
                                }
                            ]
                        }
                    ]
                }
            ]
        }
        await sendWebhookMessage(contents=message, url=self.getWebhookURL(inputMsg=inputMsg))

    async def new_custom(self, inputMsg: WebhookMessage) -> None:
        block_pieces = []
        for key, val in inputMsg.Data.items():
            block_pieces.append({
                "type": "mrkdwn",
                "text": f"*{key}*\n{val}"
            })
        message = {
            "channel": f"#{self.getWebhookChannel(inputMsg=inputMsg)}",
            "username": "Mythic",
            "icon_emoji": ":mythic:",
            "attachments": [
                {
                    "fallback": f"{inputMsg.OperatorUsername} Message!",
                    "color": "#ff0000",
                    "blocks": [
                        {
                            "type": "section",
                            "fields": block_pieces
                        }
                    ]
                }
            ]
        }
        await sendWebhookMessage(contents=message, url=self.getWebhookURL(inputMsg=inputMsg))