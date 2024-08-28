package my_event_processor

import (
	"GoServices/mythicGraphql"
	"context"
	"github.com/MythicMeta/MythicContainer/eventingstructs"
	"github.com/MythicMeta/MythicContainer/logging"
	"github.com/MythicMeta/MythicContainer/utils/sharedStructs"
)

/*
GraphQL queries identified in event_graphq.go and `genqlient.graphql` are used to generate Go code and placed in `generated.go`.
`genqlient.yaml` identifies that these are the files to process and where the resulting Go code should go.
`schema.graphql` is generated from Mythic Scripting (can be done via Jupyter container) and function to generate this is identified in Mythic's changelog docs.
*/
func Initialize() {
	myEventerName := "opsecChecker"
	myEventer := eventingstructs.EventingDefinition{
		Name:        myEventerName,
		Description: "A custom eventing container to handle additional processing on output",
		CustomFunctions: []eventingstructs.CustomFunctionDefinition{
			{
				Name:        "HealthInspectorOutputEDRCheck",
				Description: "check for EDR or bad strings in HealthInspector output",
				Function: func(input eventingstructs.NewCustomEventingMessage) eventingstructs.NewCustomEventingMessageResponse {
					response := eventingstructs.NewCustomEventingMessageResponse{
						EventStepInstanceID: input.EventStepInstanceID,
						Success:             true,
					}
					logging.LogInfo("called eventing function", "function", "opsecString", "input", input.Inputs)
					client := mythicGraphql.NewClient("https://127.0.0.1:7443/graphql/", input.Inputs["APIToken"].(string))
					/*
						payloadSearch, err := GetPayloadData(context.Background(), client, input.Inputs["PayloadUUID"].(string))
						if err != nil {
							logging.LogError(err, "failed to find payload")
							return response
						}
						if len(payloadSearch.Payload) == 0 {
							logging.LogInfo("payload doesn't exist", "payload", payloadSearch.Payload)
							return response
						}

					*/
					tagTypeID := 0
					tagTypes, err := getTagTypes(context.Background(), client, "EDR Detected")
					if err != nil {
						logging.LogError(err, "failed to fetch tag types")
					} else {
						logging.LogInfo("got tagTypes via graphql", "tagTypes", tagTypes)
						if len(tagTypes.Tagtype) == 0 {
							logging.LogInfo("about to create a new tagtype")
							createdTagType, err := CreateNewTagType(context.Background(), client, "#cc3f3f", "EDR Detected in output of task", "EDR Detected")
							if err != nil {
								logging.LogError(err, "failed to create tag type")
							} else {
								logging.LogInfo("created a new tagtype", "tagtype", createdTagType)
								tagTypeID = createdTagType.Insert_tagtype_one.Id
							}
						} else {
							tagTypeID = tagTypes.Tagtype[0].Id
						}
					}
					if tagTypeID != 0 {
						logging.LogInfo("got a tagtypeid, so creating a new tag")
						_, err = CreateNewTag(context.Background(), client, tagTypeID, "HealthInspectorOutputEDRCheck", "",
							map[string]interface{}{
								"bad string": "/Applications/BlockBlock Helper.app",
							}, int(input.Inputs["SCRIPT_TASK_ID"].(float64)))
						if err != nil {
							logging.LogError(err, "failed to create tag")
						}
						updateCallbackResponse, err := UpdateCallback(context.Background(), client, int(input.Inputs["CALLBACK_ID"].(float64)), "EDR Detected in Callback")
						if err != nil {
							logging.LogError(err, "failed to update callback")
						}
						if updateCallbackResponse.UpdateCallback.Status == "error" {
							logging.LogError(nil, updateCallbackResponse.UpdateCallback.Error)
						}
					}
					return response
				},
			},
			{
				Name:        "alert_logons",
				Description: "process new alert messages for logon data and issue tasks",
				Function: func(input eventingstructs.NewCustomEventingMessage) eventingstructs.NewCustomEventingMessageResponse {
					response := eventingstructs.NewCustomEventingMessageResponse{
						EventStepInstanceID: input.EventStepInstanceID,
						Success:             true,
					}
					logging.LogInfo("called new alert", "data", input)
					return response
				},
			},
		},
		ConditionalChecks: []eventingstructs.ConditionalCheckDefinition{
			{
				Name:        "conditionalService",
				Description: "Just a test dummy func",
				Function: func(input eventingstructs.ConditionalCheckEventingMessage) eventingstructs.ConditionalCheckEventingMessageResponse {
					return eventingstructs.ConditionalCheckEventingMessageResponse{
						Success:  true,
						SkipStep: true,
					}
				},
			},
		},
		OnContainerStartFunction: func(message sharedStructs.ContainerOnStartMessage) sharedStructs.ContainerOnStartMessageResponse {
			logging.LogInfo("started", "inputMsg", message)
			return sharedStructs.ContainerOnStartMessageResponse{}
		},
		TaskInterceptFunction: func(input eventingstructs.TaskInterceptMessage) eventingstructs.TaskInterceptMessageResponse {
			command_name, ok := input.Environment["command_name"]
			if !ok {
				return eventingstructs.TaskInterceptMessageResponse{
					Success:       true,
					BlockTask:     false,
					BypassMessage: "failed to identify command, not blocking",
					BypassRole:    eventingstructs.OPSEC_ROLE_LEAD,
				}
			}
			if command_name.(string) == "shell" {
				return eventingstructs.TaskInterceptMessageResponse{
					Success:       true,
					BlockTask:     true,
					BypassMessage: "only blocking shell command, do better",
					BypassRole:    eventingstructs.OPSEC_ROLE_LEAD,
				}
			}
			response := eventingstructs.TaskInterceptMessageResponse{
				Success:       true,
				BlockTask:     false,
				BypassMessage: "i guess you can do this",
				BypassRole:    eventingstructs.OPSEC_ROLE_LEAD,
			}
			return response
		},
		ResponseInterceptFunction: func(input eventingstructs.ResponseInterceptMessage) eventingstructs.ResponseInterceptMessageResponse {
			return eventingstructs.ResponseInterceptMessageResponse{
				Success:  true,
				Response: "intercepted!\n" + input.Environment["user_output"].(string),
			}
		},
	}
	eventingstructs.AllEventingData.Get(myEventerName).AddEventingDefinition(myEventer)
}
