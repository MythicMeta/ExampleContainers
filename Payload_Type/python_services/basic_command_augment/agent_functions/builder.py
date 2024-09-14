from mythic_container.PayloadBuilder import *
from mythic_container.MythicCommandBase import *


class BasicCommandAugment(PayloadType):
    name = "basic_command_augment"
    author = "@its_a_feature_"
    supported_os = [SupportedOS.MacOS]
    description = """This adds custom jxa functions to all supported agents."""
    agent_path = pathlib.Path(".") / "basic_command_augment"
    agent_icon_path = agent_path / "agent_functions" / "basic_command_augment.svg"
    agent_code_path = agent_path / "agent_code"
    command_augment_supported_agents = [] # this means these commands are injected into all agent's callbacks on macOS
    agent_type = AgentType.CommandAugment
