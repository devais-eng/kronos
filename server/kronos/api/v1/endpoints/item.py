from fastapi import Depends
from typing import Union, List, Optional
from sqlalchemy.orm import Session
from kronos import deps
from kronos.api.v1.endpoints import Endpoint, RequestType, include_router
from kronos.storage.view.attribute import AttributeView, AttributeValueModel
from kronos.storage.view.relation import RelationView
from kronos.storage.entity.item import ItemEntity
from kronos.storage.entity import EntityType
from fastapi.encoders import jsonable_encoder


class ItemEndpoint(Endpoint):
    _entity = ItemEntity()
    _tags = ["Items"]

    def _initialize_endpoints(self):  # noqa
        super(ItemEndpoint, self)._initialize_endpoints(query=False, patch=False, update=True)

        self.add_endpoint(self.get_item_attributes,
                          f"/{self.base_url}/" + "{item_id}/attributes",
                          RequestType.GET,
                          List[AttributeView.info()])

        self.add_endpoint(self.get_item_attribute_by_name,
                          f"/{self.base_url}/" + "{item_id}/attribute/name/{attribute_name}",
                          RequestType.GET,
                          AttributeView.info())

        self.add_endpoint(self.get_item_attribute_id_by_name,
                          f"/{self.base_url}/" + "{item_id}/attribute/name/{attribute_name}/id",
                          RequestType.GET,
                          str)

        self.add_endpoint(self.get_item_attribute_value_by_name,
                          f"/{self.base_url}/" + "{item_id}/attribute/name/{attribute_name}/value",
                          RequestType.GET,
                          AttributeValueModel)

        self.add_endpoint(self.get_item_attributes_by_attribute_type,
                          f"/{self.base_url}/" + "{item_id}/attributes/type/{attribute_type}",
                          RequestType.GET,
                          List[AttributeView.info()])

        self.add_endpoint(self.get_item_children,
                          f"/{self.base_url}/" + "{item_id}/children",
                          RequestType.GET,
                          List[ItemEntity.view.info()])

        self.add_endpoint(self.get_item_parents,
                          f"/{self.base_url}/" + "{item_id}/parents",
                          RequestType.GET,
                          List[ItemEntity.view.info()])

        self.add_endpoint(self.get_item_relations,
                          f"/{self.base_url}/" + "{item_id}/relations",
                          RequestType.GET,
                          List[RelationView.info()])

        self.add_endpoint(self.get_item_version,
                          f"/{self.base_url}/" + "{item_id}/version",
                          RequestType.GET,
                          Optional[str])

        self.add_endpoint(self.get_item_customer_id,
                          f"/{self.base_url}/" + "{item_id}/customer",
                          RequestType.GET,
                          Optional[str])

        self.add_endpoint(self.get_item_modified_by,
                          f"/{self.base_url}/" + "{item_id}/modified_by",
                          RequestType.GET,
                          Optional[str])

        self.add_endpoint(self.get_item_mac,
                          f"/{self.base_url}/" + "{item_id}/mac",
                          RequestType.GET,
                          Optional[str])

        include_router(self._router, self._tags)

    async def create(self,  # noqa
                     *,
                     kafka=Depends(deps.get_kafka),
                     model_in: Union[_entity.view.create(), List[_entity.view.create()]]
                     ):
        """
        Creates a new item.
        You can do a bulk creation by setting a list of items as the body.
        A bulk creation is executed transactional.
        """
        return await super(ItemEndpoint, self).create(kafka=kafka,
                                                      model_in=model_in,
                                                      entity_type=EntityType.ITEM)

    async def update(self,  # noqa
                     *,
                     kafka=Depends(deps.get_kafka),
                     entity_id: str,
                     model_in: _entity.view.update()
                     ):
        """
        Makes a full update on an item
        """
        return await super(ItemEndpoint, self).update(kafka=kafka,
                                                      model_in=model_in,
                                                      entity_id=entity_id,
                                                      entity_type=EntityType.ITEM)

    async def patch(self,  # noqa
                    *,
                    kafka=Depends(deps.get_kafka),
                    entity_id: str,
                    model_in: _entity.view.patch()
                    ):
        """
            Partially updates an Item
        """
        return await super(ItemEndpoint, self).patch(kafka=kafka,
                                                     model_in=model_in,
                                                     entity_id=entity_id,
                                                     entity_type=EntityType.ITEM)

    async def delete_by_id(self,  # noqa
                           entity_id: str,
                           kafka=Depends(deps.get_kafka),
                           hard: bool = True):
        """
          Delete an Item by id
        """
        return await super(ItemEndpoint, self).delete_by_id(kafka=kafka, entity_id=entity_id,
                                                            entity_type=EntityType.ITEM, hard=hard)

    async def get_item_attributes(self, *, item_id: Union[int, str]):
        """
        Return all resources of an item.
        """
        return self._entity.list_attributes(item_id=item_id)

    async def add_object_to_item(self,
                                 *,
                                 kafka: Session = Depends(deps.get_kafka),
                                 resources: Union[
                                     AttributeView.create(), List[AttributeView.create()]]):
        """
        Adds resources to an item by ID
        """
        return super().forward_request(kafka=kafka, model_in=resources)

    async def get_item_children(self, *, item_id: Union[int, str]):
        """
        Returns all the children of an item by ID
        """
        return self._entity.list_children(item_id=item_id)

    async def get_item_parents(self, *, item_id: Union[int, str]):
        """
        Returns all parents of an item by ID
        """
        return self._entity.list_parents(item_id=item_id)

    async def get_item_relations(self, *, item_id: Union[int, str]):
        """
        Returns the list of all relations containing
        the given item ID
        """
        return self._entity.relatives(item_id=item_id)

    async def get_item_version(self, *, item_id: Union[int, str]):
        """
        Returns the version of an item by ID
        """
        return self._entity.last_version(entity_id=item_id)

    async def get_item_customer_id(self, *, item_id: Union[int, str]):
        """
        Returns the customer ID of an item by ID
        """
        return self._entity.customer(item_id=item_id)

    async def get_item_modified_by(self, *, item_id: Union[int, str]):
        """
        Returns the modified_by field of an item by ID
        """
        return self._entity.last_editor(entity_id=item_id)[0]

    async def get_item_mac(self, *, item_id: Union[int, str]):
        """
        Returns the modified_by field of an item by ID
        """
        return self._entity.get_edge(item_id=item_id)

    async def get_item_attribute_by_name(self, *, item_id: Union[int, str], attribute_name: Union[int, str]):
        """
        Returns the attribute with this name attached to item with id item_id
        """
        return jsonable_encoder(self._entity.get_attribute_by_item(item_id=item_id, attribute_name=attribute_name))

    async def get_item_attribute_id_by_name(self, *, item_id: Union[int, str], attribute_name: Union[int, str]):
        """
        Returns the id of the attribute with this name attached to item with id id
        """
        return self._entity.get_attribute_by_item(item_id=item_id, attribute_name=attribute_name).id

    async def get_item_attribute_value_by_name(self, *, item_id: Union[int, str], attribute_name: Union[int, str]):
        """
        Returns the value of the attribute with this name attached to item with id id
        """
        attribute = self._entity.get_attribute_by_item(item_id=item_id, attribute_name=attribute_name)
        return AttributeValueModel(id=attribute.id, value=attribute.value)

    async def get_item_attributes_by_attribute_type(self, *, item_id: Union[int, str], attribute_type: Union[int, str]):
        """
        Return all attributes of a given type on a given item
        """
        return self._entity.get_attributes_by_type(item_id=item_id, attribute_type=attribute_type)
