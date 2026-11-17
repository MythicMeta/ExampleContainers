import pathlib
import zipfile
import tarfile
import io
import base64
from mythic_container.PayloadBuilder import *
from mythic_container.MythicCommandBase import *
from mythic_container.MythicRPC import *


class BasicWrapper(PayloadType):
    name = "basic_wrapper"
    file_extension = "zip"
    author = "Lavender-exe"
    supported_os = [
        SupportedOS.Windows,
        SupportedOS.Linux,
        SupportedOS.MacOS,
    ]
    wrapper = True
    note = (
        "Basic wrapper example: downloads the inner agent payload "
        "and packages it into a zip, 7z, or tar.gz archive."
    )
    supports_dynamic_loading = False
    c2_profiles = []
    mythic_encrypts = False
    translation_container = None
    agent_type = AgentType.Wrapper
    agent_path = pathlib.Path(__file__).resolve().parent.parent
    agent_icon_path = None
    agent_code_path = agent_path / "agent_code"
    semver = "0.4.0"

    # Mythic v4: conditional wrapper compatibility
    wrapper_payload_requirements = [
        WrapperPayloadRequirement(
            requires=WrapperPayloadRequirementRequires(
                format=PayloadBuildMetadataFormat.Shellcode,
                architecture=PayloadBuildMetadataArchitecture.X64,
                os=SupportedOS.Windows,
            ),
        ),
        WrapperPayloadRequirement(
            requires=WrapperPayloadRequirementRequires(
                format=PayloadBuildMetadataFormat.Shellcode,
                architecture=PayloadBuildMetadataArchitecture.X86,
                os=SupportedOS.Windows,
            ),
        ),
        WrapperPayloadRequirement(
            requires=WrapperPayloadRequirementRequires(
                format=PayloadBuildMetadataFormat.Exe,
                architecture=PayloadBuildMetadataArchitecture.X64,
                os=SupportedOS.Windows,
            ),
        ),
        WrapperPayloadRequirement(
            requires=WrapperPayloadRequirementRequires(
                format=PayloadBuildMetadataFormat.Exe,
                architecture=PayloadBuildMetadataArchitecture.X86,
                os=SupportedOS.Windows,
            ),
        ),
    ]

    build_parameters = [
        BuildParameter(
            name="filename",
            parameter_type=BuildParameterType.String,
            default_value="payload.bin",
            description="Filename for the payload inside the archive",
        ),
        BuildParameter(
            name="archive_format",
            parameter_type=BuildParameterType.ChooseOne,
            choices=["zip", "7z", "tar.gz"],
            default_value="zip",
            description="Archive format for the output container",
        ),
    ]

    build_steps = [
        BuildStep(
            step_name="Wrapping",
            step_description="Packaging the inner agent into an archive",
        ),
    ]

    async def build(self) -> BuildResponse:
        resp = BuildResponse(status=BuildStatus.Success)
        try:
            agent_bytes = self.wrapped_payload
            if agent_bytes is None:
                resp.set_status(BuildStatus.Error)
                resp.build_stderr = "No wrapped payload provided"
                return resp
            if isinstance(agent_bytes, str):
                agent_bytes = base64.b64decode(agent_bytes)

            filename = "payload.bin"
            archive_format = "zip"
            for bp in self.build_parameters:
                if bp.name == "filename":
                    filename = bp.value if bp.value else "payload.bin"
                elif bp.name == "archive_format":
                    archive_format = bp.value if bp.value else "zip"

            buf = io.BytesIO()

            if archive_format == "zip":
                with zipfile.ZipFile(buf, "w", zipfile.ZIP_DEFLATED) as zf:
                    zf.writestr(filename, agent_bytes)
                resp.file_extension = "zip"

            elif archive_format == "tar.gz":
                with tarfile.open(fileobj=buf, mode="w:gz") as tf:
                    info = tarfile.TarInfo(name=filename)
                    info.size = len(agent_bytes)
                    tf.addfile(info, io.BytesIO(agent_bytes))
                resp.file_extension = "tar.gz"

            elif archive_format == "7z":
                import py7zr
                with py7zr.SevenZipFile(buf, "w") as sz:
                    sz.writestr(agent_bytes, filename)
                resp.file_extension = "7z"

            else:
                resp.set_status(BuildStatus.Error)
                resp.build_stderr = f"Unknown archive format: {archive_format}"
                return resp

            await SendMythicRPCPayloadUpdatebuildStep(
                MythicRPCPayloadUpdateBuildStepMessage(
                    PayloadUUID=self.uuid,
                    StepName="Wrapping",
                    StepStdout=(
                        f"Packaged \'{filename}\' "
                        f"({len(agent_bytes)} bytes) into {archive_format}"
                    ),
                    StepSuccess=True,
                )
            )

            resp.payload = buf.getvalue()
            resp.build_message = (
                f"Wrapped payload into {archive_format} "
                f"({len(resp.payload)} bytes)"
            )

        except Exception as e:
            resp.set_status(BuildStatus.Error)
            resp.build_stderr = f"Error wrapping payload: {str(e)}"

        return resp
