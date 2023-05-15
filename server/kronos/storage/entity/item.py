from typing import AnyStr, List
from sqlalchemy.orm import relationship
from sqlalchemy import Column, String, ForeignKey
from kronos.storage.entity import Entity
from kronos.storage.view.item import ItemView
from kronos.storage.utils import read_session
from fastapi.encoders import jsonable_encoder
from kronos.exception import EntityNotFoundException


class ItemEntity(Entity):
    """
    An Item is a 'lead' entity that can own objects and share hierarchical relationships with other items.
    """
    view = ItemView

    entity_id = Column(ForeignKey(Entity.id), primary_key=True)
    entity = relationship(Entity,
                          cascade="all, delete-orphan",
                          single_parent=True,
                          post_update=True)

    name = Column(String(255), unique=True, nullable=False)  # Name of the item
    type = Column(String(45))  # Type of the Item
    customer_id = Column(String(45), nullable=True, default=None)  # Customer ID
    sync_policy = Column(String(255), nullable=True)  # Synchronization Policy
    edge_mac = Column(String(45), nullable=True, default=None)  # Edge MAC, if gateway
    # Relations
    attributes = relationship("AttributeEntity",
                              back_populates="item",
                              cascade="all, delete-orphan",
                              lazy="joined",
                              primaryjoin='AttributeEntity.item_id == ItemEntity.entity_id')

    parents = relationship("RelationEntity",
                           back_populates="child",
                           cascade="all, delete-orphan",
                           primaryjoin='RelationEntity.child_id == ItemEntity.entity_id')

    children = relationship("RelationEntity",
                            back_populates="parent",
                            cascade="all, delete-orphan",
                            primaryjoin='RelationEntity.parent_id == ItemEntity.entity_id')

    __mapper_args__ = {
        'polymorphic_identity': 'entity_item'
    }

    def list_parents(self, item_id: str) -> List["ItemEntity"]:
        """
        Returns the set of parents for this entity.
        """
        with read_session(close=True) as db:
            item = self.crud.read(db=db, entity_id=item_id)
            result_set = []
            for rel in item.parents:
                result_set.append(jsonable_encoder(rel.parent))
            return result_set

    def list_children(self, item_id: str) -> List["ItemEntity"]:
        """
        Returns the set of children for this entity.
        """
        with read_session(close=True) as db:
            item = self.crud.read(db=db, entity_id=item_id)
            result_set = []
            for rel in item.children:
                result_set.append(jsonable_encoder(rel.child))
            return result_set

    def relatives(self, item_id: AnyStr):
        """
        Returns the set of all relations for this entity.
        """
        with read_session(close=True) as db:
            item = self.crud.read(db=db, entity_id=item_id)
            return jsonable_encoder(item.parents + item.children)

    def customer(self, item_id: AnyStr):
        """
        Returns the customer of this entity.
        """
        with read_session(close=True) as db:
            return jsonable_encoder(self.crud.read(db=db, entity_id=item_id).customer_id)

    def list_attributes(self, item_id: AnyStr):
        """
        Returns all attributes of this entity.
        """
        with read_session(close=True) as db:
            return jsonable_encoder(self.crud.read(db=db, entity_id=item_id).attributes)

    def get_sync_policy(self, item_id: AnyStr):
        """
        Returns the sync policy of this entity.
        """
        with read_session(close=True) as db:
            return self.crud.read(db=db, entity_id=item_id).sync_policy

    def get_edge(self, item_id: AnyStr):
        """
        Returns the edge mac of this entity.
        """
        with read_session(close=True) as db:
            return self.crud.read(db=db, entity_id=item_id).edge_mac

    def describe(self, item_id: AnyStr):
        """
        Returns the tuple (name,type) of this entity.
        """
        with read_session(close=True) as db:
            item = self.crud.read(db=db, entity_id=item_id)
            return item.name, item.type

    def get_attribute_by_item(self, item_id: AnyStr, attribute_name) -> "AttributeEntity":  # noqa
        """
        Returns attribute by item and name
        """
        from kronos.storage.entity.attribute import AttributeEntity
        with read_session(close=True) as db:
            self.crud.read(db=db, entity_id=item_id, raise_not_found=True)
            attribute = db.query(AttributeEntity) \
                .filter(AttributeEntity.item_id == item_id) \
                .filter(AttributeEntity.name == attribute_name) \
                .first()
            if attribute is not None:
                return attribute
            raise EntityNotFoundException(f"Attribute with name {attribute_name} not found on item {item_id}")

    def get_attributes_by_type(self, item_id: AnyStr, attribute_type: AnyStr) -> "List[AttributeEntity]":  # noqa
        """
        Returns all attributes by type
        """
        from kronos.storage.entity.attribute import AttributeEntity
        with read_session(close=True) as db:
            attributes = db.query(AttributeEntity) \
                .filter(AttributeEntity.item_id == item_id) \
                .filter(AttributeEntity.type.like(attribute_type)) \
                .all()
            if attributes is not None:
                return attributes
            raise EntityNotFoundException(f"Attribute with name {attribute_type} not found on item {item_id}")
