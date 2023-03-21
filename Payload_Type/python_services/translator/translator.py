import json
import base64

from mythic_container.TranslationBase import *


class myPythonTranslation(TranslationContainer):
    name = "myPythonTranslation"
    description = "python translation service that doesn't change anything"
    author = "@its_a_feature_"

    async def translate_to_c2_format(self, inputMsg: TrMythicC2ToCustomMessageFormatMessage) -> TrMythicC2ToCustomMessageFormatMessageResponse:
        response = TrMythicC2ToCustomMessageFormatMessageResponse(Success=True)
        response.Message = inputMsg.Message
        return response

    async def translate_from_c2_format(self, inputMsg: TrCustomMessageToMythicC2FormatMessage) -> TrCustomMessageToMythicC2FormatMessageResponse:
        response = TrCustomMessageToMythicC2FormatMessageResponse(Success=True)
        response.Message = inputMsg.Message
        return response
