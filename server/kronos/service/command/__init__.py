import abc
from .resolver.mqtt import MqttResolver as Resolver
from enum import Enum
from loguru import logger
from tenacity import after_log, before_log, retry, stop_after_attempt, wait_fixed
from kronos.settings import settings
import logging


class Command(object):
    """
    A Command is used to process incoming message with their specific execution logic.
    """

    def __init__(self, device_id):
        self._logger = logger
        self._resolver = Resolver(device_id=device_id)

    @abc.abstractmethod
    def _get_payload(self, *args, **kwargs):
        pass

    @retry(
        stop=stop_after_attempt(settings.FIXED_MAX_TRIES),
        wait=wait_fixed(settings.FIXED_WAIT_SECONDS),
        before=before_log(logger, logging.INFO),
        after=after_log(logger, logging.WARN),
    )
    def execute(self, *args, **kwargs):
        """
        This function is called to actually execute the command, given any map of input parameters.

        :param kwargs: Eventual input parameters to be used in execution.
        :return: Some commands may return results (any type, they're lately converted to responses).
        """
        self._resolver.request(payload=self._get_payload(*args, **kwargs), **kwargs)
        response = self._resolver.response()
        self._logger.info(f"Response was {response}")
        return response


class CommandType(Enum):
    GET_TELEMETRY = "GetTelemetry"


class CommandBuilder(object):
    """
    CommandBuilder allows to obtain a command given its literal representation.
    """

    from .gateway import GetTelemetry

    _commands = {
        CommandType.GET_TELEMETRY.value: GetTelemetry
    }

    @staticmethod
    def build(command):
        """
        Return the actual command to be executed.

        :param command: A literal naming of the command.
        :return: The proper Command implementation.
        """
        return CommandBuilder._commands.get(command)
