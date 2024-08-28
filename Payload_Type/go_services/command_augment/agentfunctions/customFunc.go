package agentfunctions

import (
	"fmt"
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/mythicrpc"
)

func init() {
	agentstructs.AllPayloadData.Get("MyCustomCommands").AddCommand(agentstructs.Command{
		Name:                "test_path_parsing",
		Description:         "test_path_parsing tests a variety of paths to be parsed by Mythic",
		Version:             1,
		MitreAttackMappings: []string{},
		SupportedUIFeatures: []string{},
		Author:              "@its_a_feature_",
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			completed := true
			response.Completed = &completed
			pathTests := []string{
				"/root/.ssh/authorized_keys",
				"/home/bob/Desktop",
				"\\\\domain.com\\share_name\\test\\folder",
				"\\\\domain.com\\c$\\Users\\bob\\Desktop\\passwords.txt",
				"C:\\Users\\bob\\Desktop",
				"D:\\",
				"C:",
			}
			for _, pathTest := range pathTests {
				analyzedPath, err := mythicrpc.SendMythicRPCFileBrowserParsePath(mythicrpc.MythicRPCFileBrowserParsePathMessage{
					Path: pathTest,
				})
				if err != nil {
					mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
						TaskID:   taskData.Task.ID,
						Response: []byte(fmt.Sprintf("Path: %s, Result: %v\n", pathTest, err)),
					})
					continue
				}
				if analyzedPath.Success {
					mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
						TaskID:   taskData.Task.ID,
						Response: []byte(fmt.Sprintf("Path: %s, Result: %v\n", pathTest, analyzedPath.AnalyzedPath)),
					})
				} else {
					mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
						TaskID:   taskData.Task.ID,
						Response: []byte(fmt.Sprintf("Path: %s, Result: %v\n", pathTest, analyzedPath.Error)),
					})
				}
			}
			return response
		},
		TaskFunctionParseArgDictionary: func(args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
			// if we get a dictionary, it'll be from the file browser which will supply agentstructs.FileBrowserTask data
			return nil
		},
		TaskFunctionParseArgString: func(args *agentstructs.PTTaskMessageArgsData, input string) error {
			//return args.LoadArgsFromJSONString(input)
			return nil
		},
	})
}
