import json
from typing import Any, Tuple

from fastapi import APIRouter
from fastapi.responses import JSONResponse
from healthcheck import HealthCheck, EnvironmentDump

from kronos.settings import settings

router = APIRouter()
health = HealthCheck()


@router.get('/health')
async def check_health() -> Any:
    """
    Performs application health checks
    """
    content, status_code, headers = health.run()
    return JSONResponse(
        json.loads(content),
        status_code,
        headers,
    )


if settings.ENVDUMP_ENABLED:
    env_dump = EnvironmentDump()


    @router.get('/envdump')
    async def environment_dump() -> Any:
        """
        Performs a dump of enviroment variables
        """
        content, status_code, headers = env_dump.run()
        return JSONResponse(
            json.loads(content),
            status_code,
            headers,
        )
