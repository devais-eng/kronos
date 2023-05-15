from loguru import logger
import sys
from kronos.settings import settings


def init_logger():
    fmt = "{time} - {name} - {level} - {message}"
    logger.remove()
    logger.add(sys.stderr, level="INFO")
    logger.add(settings.LOG_NAME, level="DEBUG", format=fmt)
