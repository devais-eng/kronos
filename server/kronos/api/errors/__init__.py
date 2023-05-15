from starlette.exceptions import HTTPException
from fastapi import FastAPI
from fastapi.exceptions import RequestValidationError

from .http_error import http_error_handler
from .validation_error import http422_error_handler
from kronos.exception import EntityNotFoundException, EntityAlreadyExistException, IntegrityError, KronosIntegrityError, KronosRedisError
from .sync_error import entity_not_found_exception_handler, entity_already_exists_exception_handler, \
    kronos_integrity_exception_handler, sqlalchemy_integrity_exception_handler, kronos_redis_exception_handler


def register_api_error_handlers(app: FastAPI) -> None:
    app.add_exception_handler(HTTPException, http_error_handler)
    app.add_exception_handler(RequestValidationError, http422_error_handler)
    app.add_exception_handler(EntityNotFoundException, entity_not_found_exception_handler)
    app.add_exception_handler(EntityAlreadyExistException, entity_already_exists_exception_handler)
    app.add_exception_handler(IntegrityError, sqlalchemy_integrity_exception_handler)
    app.add_exception_handler(KronosIntegrityError, kronos_integrity_exception_handler)
    app.add_exception_handler(KronosRedisError, kronos_redis_exception_handler)