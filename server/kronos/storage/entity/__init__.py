from enum import Enum
from typing import Union, AnyStr, Tuple, Dict
from sqlalchemy import Column, Boolean, BigInteger, Text, String, ForeignKey, Float
from sqlalchemy.orm import Session, relationship
from sqlalchemy.ext.declarative import as_declarative, declared_attr
from time import time
from uuid import uuid4
from kronos.storage.view import View
from kronos.settings import settings
from kronos.storage.utils import transactional_session, read_session
from fastapi.encoders import jsonable_encoder


@as_declarative()
class BaseEntity:
    """
    An entity is any abstract representation which can be stored into the database.
    """

    __abstract__ = True

    id = Column(Text, primary_key=True, unique=True,
                default=lambda: str(uuid4()))  # Unique identifier for any entity

    @declared_attr
    def __tablename__(self) -> str:  # noqa
        """
        Automatically assign a table name according class.
        """
        return self.__name__.lower().replace("entity", "")


class EntityType(Enum):
    ITEM = "ITEM"
    RELATION = "RELATION"
    ATTRIBUTE = "ATTRIBUTE"


class Entity(BaseEntity):
    """
    An entity is any abstract representation which can be stored into the database.
    """

    @declared_attr
    def version_graph_id(cls):
        from kronos.graph.storage.versioning import VersioningGraphEntity
        return Column(ForeignKey(VersioningGraphEntity.id), nullable=True, default=None)

    @declared_attr
    def version_graph(cls):
        from kronos.graph.storage.versioning import VersioningGraphEntity
        return relationship(VersioningGraphEntity,
                            single_parent=True,
                            post_update=True)

    source_timestamp = Column(BigInteger, nullable=False,
                              default=lambda: time() * 1000)  # Timestamp when source generated this entity
    created_at = Column(BigInteger, nullable=False,
                        default=lambda: time() * 1000)  # Timestamp when entity was created in storage
    created_by = Column(String(20), nullable=False, default="CloudSync")  # Author of this entity

    modified_at = Column(BigInteger, nullable=True,
                         default=lambda: time() * 1000)  # Timestamp when entity was last updated
    modified_by = Column(String(20), nullable=True, default=None)  # Last editor of this entity

    version = Column(Text, nullable=False,
                     default=settings.INITIAL_VERSION)  # Last synchronized version for this entity
    active = Column(Boolean, nullable=False, default=True)  # Enable/Disable entities (i.e., for 'soft deletes')

    entity_type = Column(String(50), nullable=False)  # Enable/Disable entities (i.e., for 'soft deletes')

    __mapper_args__ = {
        'polymorphic_identity': 'entity_base',
        'polymorphic_on': entity_type
    }
    view = View

    @property
    def crud(self):
        return self._get_crud()

    @declared_attr
    def __tablename__(self) -> str:  # noqa
        """
        Automatically assign a table name according class.
        """
        return self.__name__.lower().replace("entity", "table")

    @classmethod
    def _get_crud(cls):  # -> CRUD:
        """
        Produces a CRUD operator for current class.
        """
        from kronos.storage.crud import CRUD
        return CRUD[cls,
                    cls.view.create(),
                    cls.view.info(),
                    cls.view.update(),
                    cls.view.delete(),
                    cls.view.query()](cls)

    def last_version(self, entity_id: AnyStr) -> str:
        """
        Return the last synchronized version of requested entity.
        """
        with read_session() as db:
            return self.crud.read(db=db, entity_id=entity_id).version

    def is_active(self, entity_id: AnyStr) -> bool:
        """
        Checks a requested entity is active.
        """
        with read_session() as db:
            return self.crud.read(db=db, entity_id=entity_id).active

    def author(self, entity_id: AnyStr) -> Tuple[str, int, int]:
        """
        Returns information about the author of requested entity.

        Expected Output: (author_name, creation_time, source_creation_time)
        """
        with read_session() as db:
            data = self.crud.read(db=db, entity_id=entity_id)
            return data.created_by, data.created_at, data.source_ts

    def find_by_name(self, entity_name, skip: int = 0, limit: int = 100):
        with read_session(close=True) as db:
            search = f"%{entity_name}%"
            return self.crud.like(search=search, attribute="name", db=db, skip=skip, limit=limit)

    def find_by_type(self, entity_type, skip: int = 0, limit: int = 100):
        with read_session(close=True) as db:
            search = f"%{entity_type}%"
            return self.crud.like(search=search, attribute="type", db=db, skip=skip, limit=limit)

    def last_editor(self, entity_id: AnyStr) -> Tuple[str, int]:
        """
        Returns information about the last editor of requested entity.

        Expected Output: (last_editor_name, last_edit_time)
        """
        with read_session() as db:
            data = self.crud.read(db=db, entity_id=entity_id)
            return data.modified_by, data.modified_at

    def read(self, entity_id, close=False, raise_not_found=True, raise_soft_deleted=False, json=False):
        """
        Returns information about this entity

        Expected Output: Entity attributes.
        """
        from fastapi.encoders import jsonable_encoder # noqa
        with read_session(close=close) as db:
            entity = self.crud.read(db=db, entity_id=entity_id, raise_not_found=raise_not_found,
                                    raise_soft_deleted=raise_soft_deleted)
            if json:
                return jsonable_encoder(entity)
            return entity

    def query(self, skip, limit, close=False):
        with read_session(close=close) as db:
            return jsonable_encoder(self.crud.query(db=db, skip=skip, limit=limit))

    def counts(self, close=False):
        with read_session(close=close) as db:
            return len(self.crud.query(db=db))

    def remove(self, entity_id):
        with transactional_session() as db:
            return self.crud.remove(db=db, entity_id=entity_id)

    def update(self, new: Dict, previous=None, close=False):
        """
        Update current entity with new value.

        Expected Output: upgraded entity
        """
        with transactional_session(close=close) as db:
            return self.crud.update(db=db, previous=previous, new=new)

    def create(self, to_create, close=False, init_graph=False):
        with transactional_session(close=close) as db:
            if init_graph:
                from kronos.graph.storage.versioning import VersioningGraphEntity
                graph = VersioningGraphEntity().new_graph()
                to_create.version_graph_id = graph
            entity = self.crud.create(db=db, to_create=self.view.db_create()(**to_create.dict()))
            return entity


def get_by_type(entity_type: Union[EntityType, str]) -> Entity:
    """
    Given a type, retrieves its entity implementation.
    """
    if isinstance(entity_type, str):
        entity_type = EntityType[entity_type]
    if entity_type == EntityType.ITEM:
        from kronos.storage.entity.item import ItemEntity
        return ItemEntity
    elif entity_type == EntityType.ATTRIBUTE:
        from kronos.storage.entity.attribute import AttributeEntity
        return AttributeEntity
    elif entity_type == EntityType.RELATION:
        from kronos.storage.entity.relation import RelationEntity
        return RelationEntity
    else:
        raise NotImplementedError(f"Cannot build entity of type '{entity_type}'.")
