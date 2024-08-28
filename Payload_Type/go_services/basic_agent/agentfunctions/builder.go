package agentfunctions

import (
	"bytes"
	"encoding/json"
	"fmt"
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/mythicrpc"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var payloadDefinition = agentstructs.PayloadType{
	Name:                                   "basicAgent",
	FileExtension:                          "bin",
	Author:                                 "@xorrior, @djhohnstein, @Ne0nd0g, @its_a_feature_",
	SupportedOS:                            []string{agentstructs.SUPPORTED_OS_LINUX, agentstructs.SUPPORTED_OS_MACOS},
	Wrapper:                                false,
	CanBeWrappedByTheFollowingPayloadTypes: []string{},
	SupportsDynamicLoading:                 false,
	Description:                            "A fully featured macOS and Linux Golang agent",
	SupportedC2Profiles:                    []string{"http", "websocket", "poseidon_tcp"},
	MythicEncryptsData:                     true,
	MessageFormat:                          agentstructs.MessageFormatJSON,
	BuildParameters: []agentstructs.BuildParameter{
		{
			Name:          "mode",
			Description:   "Choose the build mode option. Select default for executables, c-shared for a .dylib or .so file, or c-archive for a .Zip containing C source code with an archive and header file",
			Required:      false,
			DefaultValue:  "default",
			Choices:       []string{"default", "c-archive", "c-shared"},
			ParameterType: agentstructs.BUILD_PARAMETER_TYPE_CHOOSE_ONE_CUSTOM,
		},
		{
			Name:          "architecture",
			Description:   "Choose the agent's architecture",
			Required:      false,
			DefaultValue:  "AMD_x64",
			Choices:       []string{"AMD_x64", "ARM_x64"},
			ParameterType: agentstructs.BUILD_PARAMETER_TYPE_CHOOSE_ONE,
		},
		{
			Name:          "proxy_bypass",
			Description:   "Ignore HTTP proxy environment settings configured on the target host?",
			Required:      false,
			DefaultValue:  false,
			ParameterType: agentstructs.BUILD_PARAMETER_TYPE_BOOLEAN,
		},
		{
			Name:          "garble",
			Description:   "Use Garble to obfuscate the output Go executable.\nWARNING - This significantly slows the agent build time.",
			Required:      false,
			DefaultValue:  false,
			ParameterType: agentstructs.BUILD_PARAMETER_TYPE_BOOLEAN,
		},
		{
			Name:          "supportFiles",
			Description:   "Uploading multiple support files.",
			Required:      false,
			DefaultValue:  false,
			ParameterType: agentstructs.BUILD_PARAMETER_TYPE_FILE_MULTIPLE,
		},
	},
	BuildSteps: []agentstructs.BuildStep{
		{
			Name:        "Configuring",
			Description: "Cleaning up configuration values and generating the golang build command",
		},

		{
			Name:        "Compiling",
			Description: "Compiling the golang agent (maybe with obfuscation via garble)",
		},
		{
			Name:        "Reporting back",
			Description: "Sending the payload back to Mythic",
		},
	},
}

func build(payloadBuildMsg agentstructs.PayloadBuildMessage) agentstructs.PayloadBuildResponse {
	payloadBuildResponse := agentstructs.PayloadBuildResponse{
		PayloadUUID:        payloadBuildMsg.PayloadUUID,
		Success:            true,
		UpdatedCommandList: &payloadBuildMsg.CommandList,
	}

	if len(payloadBuildMsg.C2Profiles) > 1 || len(payloadBuildMsg.C2Profiles) == 0 {
		payloadBuildResponse.Success = false
		payloadBuildResponse.BuildStdErr = "Failed to build - must select only one C2 Profile at a time"
		return payloadBuildResponse
	}
	macOSVersion := "10.12"
	targetOs := "linux"
	if payloadBuildMsg.SelectedOS == "macOS" {
		targetOs = "darwin"
	} else if payloadBuildMsg.SelectedOS == "Windows" {
		targetOs = "windows"
	}
	// This package path is used with Go's "-X" link flag to set the value string variables in code at compile
	// time. This is how each profile's configurable options are passed in.
	poseidon_repo_profile := "github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles"

	// Build Go link flags that are passed in at compile time through the "-ldflags=" argument
	// https://golang.org/cmd/link/
	ldflags := fmt.Sprintf("-s -w -X '%s.UUID=%s'", poseidon_repo_profile, payloadBuildMsg.PayloadUUID)
	// Iterate over the C2 profile parameters and associated variable through Go's "-X" link flag
	for _, key := range payloadBuildMsg.C2Profiles[0].GetArgNames() {
		if key == "AESPSK" {
			//cryptoVal := val.(map[string]interface{})
			cryptoVal, err := payloadBuildMsg.C2Profiles[0].GetCryptoArg(key)
			if err != nil {
				payloadBuildResponse.Success = false
				payloadBuildResponse.BuildStdErr = err.Error()
				return payloadBuildResponse
			}
			ldflags += fmt.Sprintf(" -X '%s.%s=%s'", poseidon_repo_profile, key, cryptoVal.EncKey)
		} else if key == "headers" {
			headers, err := payloadBuildMsg.C2Profiles[0].GetDictionaryArg(key)
			if err != nil {
				payloadBuildResponse.Success = false
				payloadBuildResponse.BuildStdErr = err.Error()
				return payloadBuildResponse
			}
			if jsonBytes, err := json.Marshal(headers); err != nil {
				payloadBuildResponse.Success = false
				payloadBuildResponse.BuildStdErr = err.Error()
				return payloadBuildResponse
			} else {
				stringBytes := string(jsonBytes)
				stringBytes = strings.ReplaceAll(stringBytes, "\"", "\\\"")
				ldflags += fmt.Sprintf(" -X '%s.%s=%s'", poseidon_repo_profile, key, stringBytes)
			}
		} else {
			val, err := payloadBuildMsg.C2Profiles[0].GetArg(key)
			if err != nil {
				payloadBuildResponse.Success = false
				payloadBuildResponse.BuildStdErr = err.Error()
				return payloadBuildResponse
			}
			ldflags += fmt.Sprintf(" -X '%s.%s=%v'", poseidon_repo_profile, key, val)
		}
	}
	proxyBypass, err := payloadBuildMsg.BuildParameters.GetBooleanArg("proxy_bypass")
	if err != nil {
		payloadBuildResponse.Success = false
		payloadBuildResponse.BuildStdErr = err.Error()
		return payloadBuildResponse
	}
	architecture, err := payloadBuildMsg.BuildParameters.GetStringArg("architecture")
	if err != nil {
		payloadBuildResponse.Success = false
		payloadBuildResponse.BuildStdErr = err.Error()
		return payloadBuildResponse
	}
	mode, err := payloadBuildMsg.BuildParameters.GetStringArg("mode")
	if err != nil {
		payloadBuildResponse.Success = false
		payloadBuildResponse.BuildStdErr = err.Error()
		return payloadBuildResponse
	}
	garble, err := payloadBuildMsg.BuildParameters.GetBooleanArg("garble")
	if err != nil {
		payloadBuildResponse.Success = false
		payloadBuildResponse.BuildStdErr = err.Error()
		return payloadBuildResponse
	}
	ldflags += fmt.Sprintf(" -X '%s.proxy_bypass=%v'", poseidon_repo_profile, proxyBypass)
	ldflags += " -buildid="
	goarch := "amd64"
	if architecture == "ARM_x64" {
		goarch = "arm64"
	}
	tags := payloadBuildMsg.C2Profiles[0].Name
	command := fmt.Sprintf("rm -rf /deps; CGO_ENABLED=1 GOOS=%s GOARCH=%s ", targetOs, goarch)
	goCmd := fmt.Sprintf("-tags %s -buildmode %s -ldflags \"%s\"", tags, mode, ldflags)
	if targetOs == "darwin" {
		command += "CC=o64-clang CXX=o64-clang++ "
	} else if targetOs == "windows" {
		command += "CC=x86_64-w64-mingw32-gcc "
	} else {
		if goarch == "arm64" {
			command += "CC=aarch64-linux-gnu-gcc "
		}
	}
	command += "GOGARBLE=* "
	if garble {
		command += "/go/bin/garble -tiny -literals -debug -seed random build "
	} else {
		command += "go build "
	}
	payloadName := fmt.Sprintf("%s-%s", payloadBuildMsg.PayloadUUID, targetOs)
	command += fmt.Sprintf("%s -o /build/%s", goCmd, payloadName)
	if targetOs == "darwin" {
		command += fmt.Sprintf("-%s", macOSVersion)
		payloadName += fmt.Sprintf("-%s", macOSVersion)
	}
	command += fmt.Sprintf("-%s", goarch)
	payloadName += fmt.Sprintf("-%s", goarch)
	if mode == "c-shared" {
		if targetOs == "windows" {
			command += ".dll"
			payloadName += ".dll"
		} else if targetOs == "darwin" {
			command += ".dylib"
			payloadName += ".dylib"
		} else {
			command += ".so"
			payloadName += ".so"
		}
	} else if mode == "c-archive" {
		command += ".a"
		payloadName += ".a"
	}

	mythicrpc.SendMythicRPCPayloadUpdateBuildStep(mythicrpc.MythicRPCPayloadUpdateBuildStepMessage{
		PayloadUUID: payloadBuildMsg.PayloadUUID,
		StepName:    "Configuring",
		StepSuccess: true,
		StepStdout:  fmt.Sprintf("Successfully configured\n%s", command),
	})
	cmd := exec.Command("/bin/bash")
	cmd.Stdin = strings.NewReader(command)
	cmd.Dir = "./poseidon/agent_code/"
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		payloadBuildResponse.Success = false
		payloadBuildResponse.BuildMessage = "Compilation failed with errors"
		payloadBuildResponse.BuildStdErr = stderr.String() + "\n" + err.Error()
		payloadBuildResponse.BuildStdOut = stdout.String()
		mythicrpc.SendMythicRPCPayloadUpdateBuildStep(mythicrpc.MythicRPCPayloadUpdateBuildStepMessage{
			PayloadUUID: payloadBuildMsg.PayloadUUID,
			StepName:    "Compiling",
			StepSuccess: false,
			StepStdout:  fmt.Sprintf("failed to compile\n%s\n%s\n%s", stderr.String(), stdout.String(), err.Error()),
		})
		return payloadBuildResponse
	} else {
		outputString := stdout.String()
		if !garble {
			// only adding stderr if garble is false, otherwise it's too much data
			outputString += "\n" + stderr.String()
		}

		mythicrpc.SendMythicRPCPayloadUpdateBuildStep(mythicrpc.MythicRPCPayloadUpdateBuildStepMessage{
			PayloadUUID: payloadBuildMsg.PayloadUUID,
			StepName:    "Compiling",
			StepSuccess: true,
			StepStdout:  fmt.Sprintf("Successfully executed\n%s", outputString),
		})
	}
	if !garble {
		payloadBuildResponse.BuildStdErr = stderr.String()
	}
	payloadBuildResponse.BuildStdOut = stdout.String()
	if payloadBytes, err := os.ReadFile(fmt.Sprintf("/build/%s", payloadName)); err != nil {
		payloadBuildResponse.Success = false
		payloadBuildResponse.BuildMessage = "Failed to find final payload"
	} else {
		payloadBuildResponse.Payload = &payloadBytes
		payloadBuildResponse.Success = true
		payloadBuildResponse.BuildMessage = "Successfully built payload!"
	}

	//payloadBuildResponse.Status = agentstructs.PAYLOAD_BUILD_STATUS_ERROR
	return payloadBuildResponse
}

func Initialize() {
	agentstructs.AllPayloadData.Get("basicAgent").AddPayloadDefinition(payloadDefinition)
	agentstructs.AllPayloadData.Get("basicAgent").AddBuildFunction(build)
	agentstructs.AllPayloadData.Get("basicAgent").AddIcon(filepath.Join(".", "basic_agent", "agentfunctions", "basicAgent.svg"))
}
