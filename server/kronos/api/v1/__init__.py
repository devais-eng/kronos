from fastapi import APIRouter

api_router = APIRouter()


def include_router(router, tags=["Others"]):
    api_router.include_router(router, tags=tags)


from kronos.api.v1.endpoints.relation import RelationEndpoint # noqa
from kronos.api.v1.endpoints.item import ItemEndpoint # noqa
from kronos.api.v1.endpoints.attribute import AttributeEndpoint # noqa

enabled_endpoints = [RelationEndpoint, ItemEndpoint, AttributeEndpoint]

for endpoint in enabled_endpoints:
    endpoint()

from .health import router as health_router # noqa

include_router(health_router, tags=["Health"])
