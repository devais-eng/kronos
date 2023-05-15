from typing import Any, Dict, Generic, Type, TypeVar, Union
from fastapi.encoders import jsonable_encoder
from sqlalchemy.orm import Session
from loguru import logger

from kronos.storage.entity import Entity
from kronos.exception import EntityNotFoundException
from typing import List

from kronos.storage.view import EntityCreateModel, EntityUpdateModel, EntityInfoModel, EntityQueryModel, \
    EntityDeleteModel

EntityType = TypeVar("EntityType", bound=Entity)
CreateModelType = TypeVar("CreateModelType", bound=EntityCreateModel)
UpdateModelType = TypeVar("UpdateModelType", bound=EntityUpdateModel)
DeleteModelType = TypeVar("DeleteModelType", bound=EntityDeleteModel)
InfoModelType = TypeVar("InfoModelType", bound=EntityInfoModel)
QueryModelType = TypeVar("QueryModelType", bound=EntityQueryModel)


class CRUD(Generic[EntityType, CreateModelType, InfoModelType, UpdateModelType, DeleteModelType, QueryModelType]):
    """
    CRUDs are used to operate with create, read, update and delete over the storage.
    """

    def __init__(self,
                 entity: Type[EntityType]):
        """
        Instantiate a CRUD operator for a given entity.

        :param entity: The entity operated by this instance.
        """
        self._entity = entity
        self._logger = logger

    def read(self,
             db: Session,
             entity_id: Any,
             raise_not_found=True,
             raise_soft_deleted=False) -> Type[EntityType]:
        """
        Loads an entity with the given ID. If it is not present, it can raise an exception.

        :param db: Used storage client.
        :param entity_id: ID of the requested entity.
        :param raise_not_found: If true an exception is raised when entity is not found, otherwise it will return None.
        :param raise_soft_deleted: If true an exception is raised when entity is soft deleted
        """
        if type(entity_id) in [bytes]:
            entity_id = entity_id.decode()
        if not raise_soft_deleted:
            self._logger.debug(f"Query all entities.")
            found = db.query(self._entity).filter(self._entity.id == entity_id).first()
        else:
            self._logger.debug(f"Query all active entities.")
            found = db.query(self._entity).filter(self._entity.id == entity_id, self._entity.active == True).first()
        if found is None and raise_not_found:
            self._logger.debug(
                f"Fail to find entity with id {entity_id}. Raising exception because raise_not_found is enabled.")
            raise EntityNotFoundException(f"Cannot read: entity with id {entity_id} not exists.")
        return found

    def query(self,
              db: Session,
              filters: QueryModelType = None,
              skip: int = 0,
              limit: int = 100) -> Type[EntityType]:
        """
        Apply filters to query from storage.

        :param filters: Filters to be applied.
        :param skip: Amount of records to skip from results.
        :param limit: Maximum size of the result set.
        :param db: Used storage client.
        """
        result_set = db.query(self._entity)
        if filters is not None:
            self._logger.debug(f"Perform query using {filters}")
            result_set = result_set.filter(
                *filters if (isinstance(filters, list) or isinstance(filters, tuple)) else filters)
        if skip:
            self._logger.debug(f"Skipping {skip} rows")
            result_set = result_set.offset(skip)
        if limit:
            self._logger.debug(f"Limit output to {limit} rows")
            result_set = result_set.limit(limit)
        results = result_set.all()
        return [r for r in results if r.active]  # TODO: Resolve filters.

    def create(self,
               db: Session,
               *,
               to_create: CreateModelType) -> EntityType:
        """
        Apply filters to query from storage.

        :param to_create: Object to be created.
        :param db: Used storage client.
        """
        obj_in_data = to_create.dict(exclude_unset=True)
        self._logger.debug(f"Creation of {to_create}")
        parent_obj = {k: v for k, v in obj_in_data.items() if not isinstance(v, List) and v is not None}
        child_obj = {k: v for k, v in obj_in_data.items() if isinstance(v, List) and v is not None}
        db_obj = self._entity(**parent_obj)  # noqa
        for child_key, child_value in child_obj.items():
            setattr(db_obj, child_key, child_value)
        db.add(db_obj)
        db.flush()
        db.refresh(db_obj)
        return db_obj

    def update(self,
               db: Session,
               previous: Type[EntityType] = None,
               *,
               new: Union[UpdateModelType, Dict[str, Any]],
               ) -> EntityType:
        """
        Update an entity.

        :param db: Reference to storage client
        :param previous: This is the entity to be updated
        :param new: Updates to be applied.
        :return: Updated entity
        """
        if isinstance(new, dict):
            update_data = new
        else:
            update_data = new.dict(exclude_unset=True)
        if previous is not None and isinstance(previous, dict):
            try:
                persisted = self.read(db=db, entity_id=update_data.get("id"))
                obj_data = jsonable_encoder(persisted)
                obj_data.update(previous)
            except EntityNotFoundException:
                persisted = self._entity(**previous)
                obj_data = jsonable_encoder(persisted)
        elif isinstance(previous, self._entity):
            persisted = previous
            obj_data = jsonable_encoder(persisted)
        else:
            persisted = self.read(db=db, entity_id=update_data.get("id"))
            obj_data = jsonable_encoder(persisted)
        for field in obj_data:
            if field in update_data and update_data[field] is not None:
                setattr(persisted, field, update_data[field])
        if "active" not in update_data:
            setattr(persisted, "active", True)
        db.add(persisted)
        db.flush()
        db.refresh(persisted)
        return persisted

    def remove(self,
               db: Session,
               *,
               entity_id: int) -> EntityType:
        """
        Remove object using unique identifier from database.

        :param db: Reference to storage client.
        :param entity_id: The unique identifier of the entity.
        """
        obj = self.read(db=db, entity_id=entity_id)
        db.delete(obj)
        return obj

    def like(self,
             db: Session,
             *,
             search: str,
             attribute: str = "name",
             skip: int = 0,
             limit: int = 100
             ) -> Type[EntityType]:
        attribute = getattr(self._entity, attribute)
        return db.query(self._entity).filter(attribute.like(search)).offset(skip).limit(limit).all()
