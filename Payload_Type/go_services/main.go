package main

import (
	payloadtypeFunctions "GoServices/basic_agent/agentfunctions"
	translatorFunctions "GoServices/basic_agent/agentfunctions"
	"GoServices/http/c2functions"
	"GoServices/my_auth"
	"GoServices/my_event_processor"
	"GoServices/my_logger"
	"GoServices/my_webhooks"
	"GoServices/no_actual_translation/translationfunctions"
	"github.com/MythicMeta/MythicContainer"
)

func main() {
	// load up the agent functions directory so all the init() functions execute
	c2functions.Initialize()
	payloadtypeFunctions.Initialize()
	translationfunctions.Initialize()
	my_webhooks.Initialize()
	my_logger.Initialize()
	my_event_processor.Initialize()
	translatorFunctions.Initialize()
	my_auth.Initialize()
	// sync over definitions and listen
	MythicContainer.StartAndRunForever([]MythicContainer.MythicServices{
		MythicContainer.MythicServiceC2,
		MythicContainer.MythicServiceTranslationContainer,
		MythicContainer.MythicServiceWebhook,
		MythicContainer.MythicServiceLogger,
		MythicContainer.MythicServicePayload,
		MythicContainer.MythicServiceEventing,
		MythicContainer.MythicServiceAuth,
	})
}
