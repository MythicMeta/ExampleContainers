from mythic_container.EventingBase import Eventing
from mythic_container.EventingBase import CustomFunctionDefinition, NewCustomEventingMessage, NewCustomEventingMessageResponse


class EventingInstance(Eventing):
    def __init__(self, **kwargs):
        self.name = "my_eventing"
        self.description = "Checks payloads on build against bad strings"

        self.custom_functions = [
            CustomFunctionDefinition(
                Name="check_payload",
                Description="check for bad strings in the payload",
                Function=self.check_payload
            )
        ]

    async def check_payload(self, msg: NewCustomEventingMessage) -> NewCustomEventingMessageResponse:
        funcResponse = NewCustomEventingMessageResponse(Success=False)
        funcResponse.Success = True
        funcResponse.StdOut = "We are running from a custom event container"
        funcResponse.StdErr = "Testing"
        funcResponse.EventStepInstanceID = msg.EventStepInstanceID
        return funcResponse