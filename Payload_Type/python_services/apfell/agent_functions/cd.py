from mythic_container.MythicCommandBase import *
from mythic_container.MythicRPC import *

class CdArguments(TaskArguments):
    def __init__(self, command_line, **kwargs):
        super().__init__(command_line, **kwargs)
        self.args = [
            CommandParameter(
                name="path",
                type=ParameterType.Array,
                choices=["test", "bob"],
                default_value=["test"],
                description="path to change directory to",
                parameter_group_info=[ParameterGroupInfo()]
            ),
        ]

    async def parse_arguments(self):
        if len(self.command_line) == 0:
            raise ValueError("Need to specify a path")
        self.add_arg("path", self.command_line)

    async def parse_dictionary(self, dictionary_arguments):
        if "path" in dictionary_arguments:
            self.add_arg("path", dictionary_arguments["path"])
        else:
            raise ValueError("Missing 'path' argument")

async def formulate_output( task: PTTaskCompletionFunctionMessage) -> PTTaskCompletionFunctionMessageResponse:
    # Check if the task is complete
    response = PTTaskCompletionFunctionMessageResponse(Success=True, TaskStatus="success")
    if task.TaskData.Task.Completed is True:
        # Check if the task was a success
        if task.TaskData.Task.Status == MythicStatus.Success:
            # Get the interval and jitter from the task information
            interval = task.TaskData.args.get_arg("interval")
            jitter = task.TaskData.args.get_arg("interval")

            # Format the output message
            output = "Set sleep interval to {} seconds with a jitter of {}%.".format(
                interval / 1000, jitter
            )
        else:
            output = "Failed to execute sleep"

        # Send the output to Mythic
        resp = await MythicRPC().execute(
            "create_output", task_id=task.TaskData.Task.ID, output=output.encode()
        )

        if resp != MythicStatus.Success:
            raise Exception("Failed to execute MythicRPC function.")
    return response

class CdCommand(CommandBase):
    cmd = "cd"
    needs_admin = False
    help_cmd = "cd [path]"
    description = "Change the current working directory to another directory. No quotes are necessary and relative paths are fine"
    version = 1
    author = "@its_a_feature_"
    argument_class = CdArguments
    attackmapping = ["T1083"]
    completion_functions = {"formulate_output": formulate_output}
    script_only = True

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        resp = await MythicRPC().execute("create_artifact", task_id=task.id,
            artifact="fileManager.changeCurrentDirectoryPath",
            artifact_type="API Called",
        )
        task.args.add_arg("path", 5, ParameterType.ChooseOne)
        task.completed_callback_function = "formulate_output"
        #task.status = "error, my custom error"
        return task

    async def process_response(self, response: AgentResponse):
        pass



