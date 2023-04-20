from mythic_container.LoggingBase import *
import logging
from logging.handlers import RotatingFileHandler
from pathlib import Path


class MyLogger(Log):
    LogToFilePath = "."
    LogLevel = logging.DEBUG
    LogMaxSizeInMB = 10
    LogMaxBackups = 5

    def __init__(self):
        self.mylogger = logging.getLogger('mylogger')
        self.mylogger.setLevel(logging.DEBUG)
        myhandler = RotatingFileHandler(f"{Path(self.LogToFilePath) / 'mythic.log'}",
                                        maxBytes=self.LogMaxSizeInMB * 1024 * 1024,
                                        backupCount=self.LogMaxBackups)
        self.mylogger.addHandler(myhandler)

    async def new_task(self, msg: LoggingMessage) -> None:
        self.mylogger.info(msg)

    async def new_credential(self, msg: LoggingMessage) -> None:
        self.mylogger.info(msg)

    async def new_keylog(self, msg: LoggingMessage) -> None:
        self.mylogger.info(msg)

    async def new_file(self, msg: LoggingMessage) -> None:
        self.mylogger.info(msg)

    async def new_callback(self, msg: LoggingMessage) -> None:
        self.mylogger.info(msg)

    async def new_payload(self, msg: LoggingMessage) -> None:
        self.mylogger.info(msg)

    async def new_artifact(self, msg: LoggingMessage) -> None:
        self.mylogger.info(msg)
