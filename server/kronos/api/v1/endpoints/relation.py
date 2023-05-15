from fastapi import Depends
from kronos import deps
from kronos.api.v1.endpoints import Endpoint, RequestType, include_router
from kronos.storage.entity.relation import RelationEntity
from kronos.storage.entity import EntityType
from typing import Union, List
from kronos.storage.view import EntityCreateModel
from kronos.storage.view.relation import RelationModel


class RelationCreateApiModel(EntityCreateModel, RelationModel):
    pass


class RelationEndpoint(Endpoint):
    _entity = RelationEntity()
    _tags = ["Relations"]

    def _initialize_endpoints(self):  # noqa
        super(RelationEndpoint, self)._initialize_endpoints(query=False, patch=False, read_by_id=False, update=False,
                                                            delete=False, by_name=False, by_type=False)
        self.add_endpoint(self.read_by_id,
                          f"/{self.base_url}/" + "{parent_id}/{child_id}",
                          RequestType.GET,
                          self._entity.view.info())
        # self.add_endpoint(self.update,
        #                   f"/{self.base_url}/" + "{parent_id}/{child_id}",
        #                   RequestType.PUT,
        #                   self._entity.view.info())
        self.add_endpoint(self.delete_by_id,
                          f"/{self.base_url}/" + "{parent_id}/{child_id}",
                          RequestType.DELETE,
                          Union[self._entity.view.delete(), List[self._entity.view.delete()]])
        include_router(self._router, self._tags)

    async def create(self,  # noqa
                     *,
                     kafka=Depends(deps.get_kafka),
                     model_in: Union[RelationCreateApiModel, List[RelationCreateApiModel]]
                     ):
        """
            Creates a new relation between two items.
            You can also create multiple relations at once by using a list.
        """
        if not isinstance(model_in, List):
            model_in = [model_in]
        model_in = [self._entity.view.create()(**model.dict(), id=f"{model.parent_id}->{model.child_id}") for model in
                    model_in]
        return await super(RelationEndpoint, self).create(kafka=kafka,
                                                          model_in=model_in,
                                                          entity_type=EntityType.RELATION
                                                          )

    # async def update(self,  # noqa
    #                  *,
    #                  kafka=Depends(deps.get_kafka),
    #                  parent_id: str,
    #                  child_id: str,
    #                  model_in: _entity.view.update()
    #                  ):
    #     """
    #      Update relation by id.
    #     """
    #     internal_id = f"{parent_id}->{child_id}"
    #     return await super(RelationEndpoint, self).update(kafka=kafka,
    #                                                       model_in=model_in,
    #                                                       entity_id=internal_id,
    #                                                       entity_type=EntityType.RELATION)

    async def patch(self,  # noqa
                    *,
                    kafka=Depends(deps.get_kafka),
                    entity_id: str,
                    model_in: _entity.view.patch()
                    ):
        return await super(RelationEndpoint, self).patch(kafka=kafka,
                                                         model_in=model_in,
                                                         entity_id=entity_id,
                                                         entity_type=EntityType.RELATION)

    async def delete_by_id(self,  # noqa
                           parent_id: str,
                           child_id: str,
                           kafka=Depends(deps.get_kafka),
                           hard: bool = True):
        """
          Delete a relation by id
        """
        internal_id = f"{parent_id}->{child_id}"
        return await super(RelationEndpoint, self).delete_by_id(kafka=kafka,
                                                                entity_id=internal_id,
                                                                entity_type=EntityType.RELATION, hard=hard)

    async def read_by_id(self,  # noqa
                         parent_id,
                         child_id):  # noqa
        """
        Read relation by id
        """
        internal_id = f"{parent_id}->{child_id}"
        return self._entity.read(entity_id=internal_id, json=True, close=True,
                                 raise_soft_deleted=True)
