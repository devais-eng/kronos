from enum import Enum
from fastapi import APIRouter, Depends, status
from loguru import logger
from kronos import deps
from kronos.storage.entity import Entity
from pydantic import BaseModel
from typing import List, AnyStr, Union
from kronos.api.v1 import include_router
from kronos.service.broker import KafkaBroker
from kronos.settings import settings
from uuid import uuid1
import time
from kronos.sync.transaction.model import TransactionRequest, to_request, TransactionErrorCode
from kronos.transaction import TransactionType
from aiocache import Cache
from kronos.storage.entity import EntityType
from kronos.exception import get_from_error_code, KronosCacheNotFoundError
from fastapi.encoders import jsonable_encoder
from tenacity import after_log, before_log, stop_after_attempt, wait_fixed
from tenacity import retry as retry_ten
import logging


class RequestType(Enum):
    GET = "GET"
    POST = "POST"
    DELETE = "DELETE"
    PATCH = "PATCH"
    PUT = "PUT"


class Endpoint:
    _entity = Entity()
    _tags = ["Base"]

    def __init__(self):
        self._logger = logger
        self._router = APIRouter()
        self._initialize_endpoints()

    @property
    def base_url(self):
        return self.__class__.__name__.lower().replace('endpoint', '')

    def _initialize_endpoints(self, read_all=True, counts=True, read_by_id=True, create=True,
                              update=True, patch=True, delete=True, query=True, by_name=True, by_type=True):
        """
        Initialize base CRUD endpoints for given entity.
        """
        if read_all:
            self.add_endpoint(self.read_all,
                              f"/{self.base_url}s",
                              RequestType.GET,
                              List[self._entity.view.info()])
        if counts:
            self.add_endpoint(
                self.counts,
                f"/{self.base_url}s/count",
                RequestType.GET,
                Union[int, AnyStr, self._entity.view.count()]
            )
        if read_by_id:
            self.add_endpoint(self.read_by_id,
                              f"/{self.base_url}/" + "{entity_id}",
                              RequestType.GET,
                              self._entity.view.info())

        if create:
            self.add_endpoint(self.create,
                              f"/{self.base_url}s",
                              RequestType.POST,
                              List[self._entity.view.info()])

        if update:
            self.add_endpoint(self.update,
                              f"/{self.base_url}/" + "{entity_id}",
                              RequestType.PUT,
                              self._entity.view.info())

        if patch:
            self.add_endpoint(self.patch, f"/{self.base_url}/" + "{entity_id}",
                              RequestType.PATCH,
                              self._entity.view.info())

        if delete:
            self.add_endpoint(self.delete_by_id,
                              f"/{self.base_url}/" + "{entity_id}",
                              RequestType.DELETE,
                              Union[self._entity.view.delete(), List[self._entity.view.delete()]])

        if query and self._entity.view.query() is not None:
            self.add_endpoint(self.query,
                              f"/{self.base_url}/query",
                              RequestType.POST,
                              self._entity.view.info())

        if by_name:
            self.add_endpoint(
                self.find_by_name,
                f"/{self.base_url}/" + "findByName/{entity_name}",
                RequestType.GET,
                Union[List[self._entity.view.info()], self._entity.view.info()]
            )
        if by_type:
            self.add_endpoint(
                self.find_by_type,
                f"/{self.base_url}/" + "findByType/{entity_type}",
                RequestType.GET,
                Union[List[self._entity.view.info()], self._entity.view.info()]
            )

        include_router(self._router, self._tags)

    def add_endpoint(self, func, path, request_type: RequestType, response_model, status_code=status.HTTP_200_OK):
        api_router = self._router.api_route(path,
                                            response_model=response_model,
                                            methods=[request_type.value],
                                            status_code=status_code)
        api_router(func)

    async def read_by_id(self, entity_id: AnyStr):
        return self._entity.read(entity_id=entity_id, json=True, close=True, raise_soft_deleted=True)

    async def delete_by_id(self,
                           entity_id: AnyStr,
                           entity_type: EntityType,
                           kafka: KafkaBroker = Depends(deps.get_kafka),
                           hard: bool = True
                           ):
        await self.forward_request(kafka=kafka,
                                   model_in=self._entity.view.delete()(id=entity_id, hard=hard),
                                   entity_type=entity_type,
                                   entity_id=entity_id)
        return self._entity.view.delete()(
            id=entity_id,
            hard=hard
        )

    async def read_all(self,
                       page: int = 0,
                       page_size: int = 100):
        return self._entity.query(skip=page, limit=page_size, close=True)

    @retry_ten(
        stop=stop_after_attempt(settings.FIXED_MAX_TRIES),
        wait=wait_fixed(settings.FIXED_WAIT_SECONDS),
        before=before_log(logger, logging.INFO),
        after=after_log(logger, logging.WARN),
    )
    async def get_response(self, entity_id, entity_model=None):
        if entity_model is None:
            result = self._entity.read(entity_id=entity_id, raise_not_found=True, json=True,
                                       close=True)
        else:
            result = entity_model.read(entity_id=entity_id, raise_not_found=True, json=True,
                                       close=True)
        return result

    @retry_ten(
        stop=stop_after_attempt(settings.FIXED_MAX_TRIES),
        wait=wait_fixed(settings.FIXED_WAIT_SECONDS)
    )
    async def load_from_cache(self, redis, request_id: AnyStr):
        self._logger.info(f"Getting response of request with uuid: {request_id} from backend")
        cached_response = await redis.get(request_id)
        if not cached_response:
            self._logger.info("Response unavailable")
            raise KronosCacheNotFoundError()
        return cached_response

    async def forward_request(self,
                              *,
                              kafka: KafkaBroker = Depends(deps.get_kafka),
                              redis: Cache.REDIS = next(deps.get_redis()),
                              model_in: Union[BaseModel, List[BaseModel]],
                              entity_id: str = None,
                              entity_type: EntityType,
                              entity_model=None):

        if not isinstance(model_in, List):
            model_in = [model_in]
        request = TransactionRequest(
            ts=time.time(),
            command=TransactionType.CRUD,
            request_id=str(uuid1()),
            request_type="TransactionApp",
            payload=[to_request(model, ix, entity_type, entity_id if entity_id else model.id, ix, len(model_in)) for
                     ix, model in enumerate(model_in)]
        )
        kafka.send(settings.SERVICE_REQUESTS, request.dumps(serializer="json"))
        redis_response = None
        try:
            redis_response = await self.load_from_cache(redis=redis, request_id=request.request_id)
        except Exception as e:
            get_from_error_code(TransactionErrorCode.REDIS_ERROR.value, "Unable to get response from redis")

        if redis_response and redis_response.get("response_code", None) and redis_response.get("response_code",
                                                                                               None) != 0:
            get_from_error_code(redis_response.get("error_code", 0), redis_response.get("error_message", ""))
        else:
            result_set = []
            for entity in redis_response.get("response_body", []):
                action = entity.get('action', None)
                self._logger.debug(f"Got action {entity}")
                if action and action != "DELETE":
                    entity_id = entity.get("entity_id", None)
                    self._logger.debug(
                        f"Processing entity with id {entity_id}"
                    )
                    result_set.append(await self.get_response(entity_id=entity_id, entity_model=entity_model))
            return result_set

    async def create(self,
                     *,
                     kafka: KafkaBroker = Depends(deps.get_kafka),
                     model_in: Union[BaseModel, List[BaseModel]],
                     entity_type: EntityType
                     ):
        return await self.forward_request(kafka=kafka, model_in=model_in,
                                          entity_type=entity_type)

    async def update(self,
                     *,
                     kafka: KafkaBroker = Depends(deps.get_kafka),
                     model_in: BaseModel,
                     entity_id: str,
                     entity_type: EntityType,
                     entity_model=None
                     ):
        result = await self.forward_request(kafka=kafka, model_in=model_in, entity_id=entity_id,
                                            entity_type=entity_type, entity_model=entity_model)
        if len(result) > 0:
            return result[0]

    async def counts(self):
        return self._entity.view.count()(count=self._entity.counts())

    async def patch(self,
                    *,
                    kafka: KafkaBroker = Depends(deps.get_kafka),
                    model_in: BaseModel,
                    entity_id: str,
                    entity_type: EntityType
                    ):
        result = await self.forward_request(kafka=kafka, model_in=model_in, entity_id=entity_id,
                                            entity_type=entity_type)
        if len(result) > 0:
            return result[0]

    async def query(self,
                    *,
                    query_in: BaseModel
                    ):
        return self._entity.query(filters=query_in)

    async def find_by_name(self,
                           *,
                           entity_name: str,
                           skip: int = 0,
                           limit: int = 100):
        """
        Find entity by name
        """
        return [self._get_new_entity(jsonable_encoder(entity)) for entity in self._entity.find_by_name(
            entity_name=entity_name, skip=skip, limit=limit
        )]

    async def find_by_type(self,
                           *,
                           entity_type: str,
                           skip: int = 0,
                           limit: int = 100):
        """
        Find entity by type
        """
        return [self._get_new_entity(jsonable_encoder(entity)) for entity in self._entity.find_by_type(
            entity_type=entity_type, skip=skip, limit=limit
        )]

    def _get_new_entity(self, entity):  # noqa
        return entity
