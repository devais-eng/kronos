from loguru import logger
from abc import abstractmethod
from uuid import uuid4


class Resolver:

    def __init__(self, device_id):
        self._logger = logger
        self._device_id = device_id
        self._client_id = str(uuid4())
        self._request_id = str(uuid4())

    @property
    def request_id(self):
        return self._request_id

    @abstractmethod
    def request(self, *args, **kwargs):
        pass

    @abstractmethod
    def response(self, *args, **kwargs):
        pass
