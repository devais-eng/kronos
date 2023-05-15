import json

from starlette.requests import Request
from starlette.responses import JSONResponse
from starlette import status

from kronos.exception import EntityNotFoundException, EntityAlreadyExistException, IntegrityError, KronosIntegrityError


async def entity_not_found_exception_handler(request: Request, exc: EntityNotFoundException):
    return JSONResponse({'errors': [exc.message]}, status_code=status.HTTP_404_NOT_FOUND)


async def entity_already_exists_exception_handler(request: Request, exc: EntityAlreadyExistException):
    return JSONResponse({'errors': [exc.message]}, status_code=status.HTTP_409_CONFLICT)


async def sqlalchemy_integrity_exception_handler(request: Request, exc: IntegrityError):
    return JSONResponse({'errors': [str(exc)]}, status_code=status.HTTP_422_UNPROCESSABLE_ENTITY)


async def kronos_integrity_exception_handler(request: Request, exc: IntegrityError): # noqa
    return JSONResponse({'errors': [str(exc)]}, status_code=status.HTTP_422_UNPROCESSABLE_ENTITY)

async def kronos_redis_exception_handler(request: Request, exc: IntegrityError): # noqa
    return JSONResponse({'errors': [str(exc)]}, status_code=status.HTTP_504_GATEWAY_TIMEOUT)
