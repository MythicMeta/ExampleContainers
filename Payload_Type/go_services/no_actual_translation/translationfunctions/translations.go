package translationfunctions

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"github.com/MythicMeta/MythicContainer/logging"
	"github.com/MythicMeta/MythicContainer/translationstructs"
)

var myTranslationService = translationstructs.TranslationContainer{
	Name:        "myTranslationService",
	Description: "My test translation service that doesn't actually translate anything",
	Author:      "@its_a_feature_",
	GenerateEncryptionKeys: func(input translationstructs.TrGenerateEncryptionKeysMessage) translationstructs.TrGenerateEncryptionKeysMessageResponse {
		response := translationstructs.TrGenerateEncryptionKeysMessageResponse{
			Success: false,
		}
		if keys, err := GenerateKeysForPayload(input.CryptoParamValue); err != nil {
			logging.LogError(err, "Failed to determine valid crypto type")
			response.Error = err.Error()
			return response
		} else {
			response.Success = true
			response.EncryptionKey = keys.EncKey
			response.DecryptionKey = keys.DecKey
			return response
		}
	},
	TranslateMythicToCustomFormat: func(input translationstructs.TrMythicC2ToCustomMessageFormatMessage) translationstructs.TrMythicC2ToCustomMessageFormatMessageResponse {
		response := translationstructs.TrMythicC2ToCustomMessageFormatMessageResponse{}
		if input.MythicEncrypts {
			// mythic will take our resulting bytes and encrypt them, so just convert to bytes and return
			if jsonBytes, err := json.Marshal(input.Message); err != nil {
				response.Success = false
				response.Error = err.Error()
			} else {
				response.Success = true
				response.Message = jsonBytes
			}
		} else {

		}
		return response
	},
	TranslateCustomToMythicFormat: func(input translationstructs.TrCustomMessageToMythicC2FormatMessage) translationstructs.TrCustomMessageToMythicC2FormatMessageResponse {
		response := translationstructs.TrCustomMessageToMythicC2FormatMessageResponse{}
		outputMap := map[string]interface{}{}
		if input.MythicEncrypts {
			// mythic already decrypted these bytes, so just convert to map and return
			if err := json.Unmarshal(input.Message, &outputMap); err != nil {
				response.Success = false
				response.Error = err.Error()
			} else {
				response.Success = true
				response.Message = outputMap
			}
		} else {
			// we're expected to decrypt the bytes first, then convert them
		}

		return response
	},
}

func GenerateKeysForPayload(cryptoType string) (translationstructs.CryptoKeys, error) {
	switch cryptoType {
	case "aes256_hmac":
		bytes := make([]byte, 32)
		if _, err := rand.Read(bytes); err != nil {
			logging.LogError(err, "Failed to generate new random 32 bytes for aes256 key")
			return translationstructs.CryptoKeys{
				EncKey: nil,
				DecKey: nil,
				Value:  cryptoType,
			}, err
		}
		return translationstructs.CryptoKeys{
			EncKey: &bytes,
			DecKey: &bytes,
			Value:  cryptoType,
		}, nil
	case "none":
		return translationstructs.CryptoKeys{
			EncKey: nil,
			DecKey: nil,
			Value:  cryptoType,
		}, nil
	default:
		return translationstructs.CryptoKeys{
			EncKey: nil,
			DecKey: nil,
			Value:  cryptoType,
		}, errors.New("Unknown crypto type")
	}
}

func Initialize() {
	translationstructs.AllTranslationData.Get("myTranslationService").AddPayloadDefinition(myTranslationService)
}
