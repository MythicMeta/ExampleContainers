package main

import (
	basicAgent "GoServices/basic_agent/agentfunctions"
	httpfunctions "GoServices/http/c2functions"
	"GoServices/my_logger"
	"GoServices/my_webhooks"
	mytranslatorfunctions "GoServices/no_actual_translation/translationfunctions"
	"github.com/MythicMeta/MythicContainer"
)

func main() {
	// load up the agent functions directory so all the init() functions execute
	httpfunctions.Initialize()
	basicAgent.Initialize()
	mytranslatorfunctions.Initialize()
	my_webhooks.Initialize()
	my_logger.Initialize()
	// sync over definitions and listen
	MythicContainer.StartAndRunForever([]MythicContainer.MythicServices{
		MythicContainer.MythicServiceC2,
		MythicContainer.MythicServiceTranslationContainer,
		MythicContainer.MythicServiceWebhook,
		MythicContainer.MythicServiceLogger,
		MythicContainer.MythicServicePayload,
	})
}
