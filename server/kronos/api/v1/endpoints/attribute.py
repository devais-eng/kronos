from fastapi import Depends
from kronos import deps
from kronos.storage.entity.attribute import AttributeEntity
from kronos.storage.entity.item import ItemEntity
from kronos.api.v1.endpoints import Endpoint, include_router, RequestType
from kronos.storage.entity import EntityType
from typing import Union, List, Optional, Dict


class AttributeEndpoint(Endpoint):
    _entity = AttributeEntity()
    _tags = ["Attribute"]

    def _initialize_endpoints(self):  # noqa
        super(AttributeEndpoint, self)._initialize_endpoints(query=False,
                                                             patch=False,
                                                             update=True)
        self.add_endpoint(self.get_parent_item,
                          f"/{self.base_url}" + "/{entity_id}/item",
                          RequestType.GET,
                          Optional[ItemEntity.view.info()])
        self.add_endpoint(self.get_attribute_value,
                          f"/{self.base_url}" + "/{entity_id}/value",
                          RequestType.GET,
                          Optional[Dict])

        include_router(self._router, self._tags)

    async def create(self,  # noqa
                     *,
                     kafka=Depends(deps.get_kafka),
                     model_in: Union[_entity.view.create(), List[_entity.view.create()]]
                     ):
        """
        Adds resource to an item
        """
        return await super(AttributeEndpoint, self).create(kafka=kafka,
                                                           model_in=model_in,
                                                           entity_type=EntityType.ATTRIBUTE
                                                           )

    async def update(self,  # noqa
                     *,
                     kafka=Depends(deps.get_kafka),
                     model_in: _entity.view.update(),
                     item_id: str,
                     entity_id: str,
                     instance_id: str
                     ):
        """
            Updates a resource by resource ID, object ID or instance ID
        """
        return await super(AttributeEndpoint, self).update(kafka=kafka,
                                                           model_in=model_in,
                                                           entity_id=entity_id,
                                                           entity_type=EntityType.ATTRIBUTE
                                                           )

    async def patch(self,  # noqa
                    *,
                    kafka=Depends(deps.get_kafka),
                    entity_id: str,
                    model_in: _entity.view.patch()
                    ):
        """
            Partially updates an attribute by resource ID, object ID or instance ID
        """
        return await super(AttributeEndpoint, self).patch(kafka=kafka,
                                                          model_in=model_in,
                                                          entity_id=entity_id,
                                                          entity_type=EntityType.ATTRIBUTE
                                                          )

    async def delete_by_id(self,  # noqa
                           entity_id: str,
                           kafka=Depends(deps.get_kafka),
                           hard: bool = True):
        """
          Delete an attribute by id
        """
        return await super(AttributeEndpoint, self).delete_by_id(kafka=kafka,
                                                                 entity_id=entity_id,
                                                                 entity_type=EntityType.ATTRIBUTE,
                                                                 hard=hard)

    async def get_attribute_value(self,
                                  *,
                                  entity_id: str):
        return self._entity.get_value(attribute_id=entity_id)

    async def get_parent_item(self,
                              *,
                              entity_id: str):
        return self._entity.get_item(
            resource_id=entity_id
        )
