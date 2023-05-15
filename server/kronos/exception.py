from kronos.sync.transaction.model import TransactionErrorCode
from sqlalchemy.exc import IntegrityError
from kronos.sync.transaction.conflicts import ReadSyncEntityNotExists, CreateOnExistingEntity, ReadSyncMismatch, \
    EntityAlreadyDeleted
from loguru import logger


class EntityNotFoundException(Exception):  # Used when queried entities are not found

    def __init__(self, message=""):
        self._message = message

    @property
    def message(self):
        return self._message


class EntityAlreadyExistException(Exception):  # Used when creating over a pre-existing entity

    def __init__(self, message=""):
        self._message = message

    @property
    def message(self):
        return self._message


class KronosIntegrityError(Exception):  # noqa Used with sqlalchemy integrity error

    def __init__(self, message=""):
        self._message = message

    @property
    def message(self):
        return self._message


class KronosRedisError(Exception):  # noqa Used with sqlalchemy integrity error

    def __init__(self, message=""):
        self._message = message

    @property
    def message(self):
        return self._message

class KronosCacheNotFoundError(Exception):  # noqa Used when a key not found in redis

    def __init__(self, message=""):
        self._message = message

    @property
    def message(self):
        return self._message


def handle_exception(value):
    def decorate(f):
        def applicator(*args, **kwargs):
            try:
                return f(*args, **kwargs)
            except Exception as e:  # noqa
                import traceback
                logger.error(f'An error occurred in {value}. {traceback.format_tb(e.__traceback__)}')

        return applicator

    return decorate


def get_error_code(exception: Exception):
    if isinstance(exception, (EntityAlreadyExistException, CreateOnExistingEntity)):
        return TransactionErrorCode.ENTITY_ALREADY_EXISTS
    elif isinstance(exception,
                    (EntityAlreadyDeleted, ReadSyncEntityNotExists, EntityNotFoundException, ReadSyncMismatch)):
        return TransactionErrorCode.ENTITY_NOT_FOUND
    elif isinstance(exception, IntegrityError):
        return TransactionErrorCode.INTEGRITY_ERROR
    return TransactionErrorCode.GENERAL


def get_from_error_code(error_code: int, message):
    error_code = TransactionErrorCode(error_code)
    if error_code == TransactionErrorCode.ENTITY_NOT_FOUND:
        raise EntityNotFoundException(message)
    elif error_code == TransactionErrorCode.ENTITY_ALREADY_EXISTS:
        raise EntityAlreadyExistException(message)
    elif error_code == TransactionErrorCode.INTEGRITY_ERROR:
        raise KronosIntegrityError(message)
    elif error_code == TransactionErrorCode.REDIS_ERROR:
        raise KronosRedisError(message)
    raise Exception(message)
