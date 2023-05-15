from typing import Generator
from kronos.service.broker import KafkaBroker
from aiocache import Cache
from kronos.settings import settings


def get_kafka() -> Generator:
    """
    Get an instance of kafka
    :return:
    """
    kafka = KafkaBroker()
    yield kafka


def get_redis() -> Generator:
    cache = Cache(Cache.REDIS, endpoint=settings.REDIS_HOSTNAME, port=settings.REDIS_PORT,
                  namespace=settings.REDIS_NAMESPACE)
    yield cache
