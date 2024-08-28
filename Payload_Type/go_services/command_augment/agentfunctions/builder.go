package agentfunctions

import (
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"path/filepath"
)

var payloadDefinition = agentstructs.PayloadType{
	Name:          "MyCustomCommands",
	Author:        "@its_a_feature_",
	SupportedOS:   []string{agentstructs.SUPPORTED_OS_LINUX, agentstructs.SUPPORTED_OS_MACOS},
	Description:   "Extra commands I want added to all agents",
	AgentType:     agentstructs.AgentTypeCommandAugment,
	MessageFormat: agentstructs.MessageFormatJSON,
}

func Initialize() {
	agentstructs.AllPayloadData.Get("MyCustomCommands").AddPayloadDefinition(payloadDefinition)
	agentstructs.AllPayloadData.Get("MyCustomCommands").AddIcon(filepath.Join(".", "command_augment", "agentfunctions", "myCustomCommands.svg"))
}
