#from mywebhook.webhook import *
import mythic_container
import asyncio
#import basic_python_agent
#import websocket.mythic.c2_functions.websocket
#from translator.translator import *
#from my_logger import logger
from basic_command_augment import *
from basic_eventer.my_eventing import *

mythic_container.mythic_service.start_and_run_forever()