from starlette.middleware.cors import CORSMiddleware
from raven import Client
from kronos.settings import settings
from kronos.api.v1 import api_router
from kronos.deps import get_redis
from sentry_sdk.integrations.asgi import SentryAsgiMiddleware
from fastapi import FastAPI
from starlette_exporter import PrometheusMiddleware, handle_metrics
from kronos.logger import init_logger
from kronos.api.errors import register_api_error_handlers


def get_app() -> FastAPI:
    # Create a FastAPI application
    app = FastAPI(
        title=settings.PROJECT_NAME, openapi_url=f"{settings.API_V1_STR}/openapi.json"
    )

    # Instantiate a Sentry Client
    if settings.SENTRY_DSN is not None:
        client_sentry = Client(settings.SENTRY_DSN)
        app.add_middleware(SentryAsgiMiddleware)

    init_logger()

    # Set all CORS enabled origins
    if settings.BACKEND_CORS_ORIGINS:
        app.add_middleware(
            CORSMiddleware,
            allow_origins=[str(origin) for origin in settings.BACKEND_CORS_ORIGINS],
            allow_credentials=True,
            allow_methods=["*"],
            allow_headers=["*"],
        )

    # Enable prometheus
    if settings.PROMETHEUS_ENABLED:
        app.add_middleware(PrometheusMiddleware, app_name=settings.PROJECT_NAME, group_paths=True,
                           buckets=[0.1, 0.25, 0.5])
        app.add_route("/metrics", handle_metrics)
    # Include the main API router
    app.include_router(api_router, prefix=settings.API_V1_STR)
    register_api_error_handlers(app)

    # Start cache
    @app.on_event("startup")
    async def startup():
        from kronos.service.broker import ResponseResolver, KafkaBroker
        ResponseResolver(broker=KafkaBroker(), cache=next(get_redis())).start()

    return app


app = get_app()
