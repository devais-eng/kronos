from typing import AnyStr
from sqlalchemy import Column, String, ForeignKey
from sqlalchemy.orm import relationship
from kronos.storage.entity import Entity
from kronos.storage.entity.item import ItemEntity
from kronos.storage.view.attribute import AttributeView
from fastapi.encoders import jsonable_encoder
from kronos.storage.utils import read_session
from sqlalchemy import UniqueConstraint


class AttributeEntity(Entity):
    """
    An attribute is composed of multiple resources and can be assigned to one item.
    """
    view = AttributeView

    entity_id = Column(ForeignKey(Entity.id), primary_key=True)
    entity = relationship(Entity,
                          cascade="all, delete-orphan",
                          single_parent=True,
                          post_update=True)

    type = Column(String(255), nullable=False)  # Type of the attribute
    name = Column(String(255), nullable=False)  # Name of the attribute

    value = Column(String(45), nullable=True, default=None)  # Serialized payload of this attribute
    value_type = Column(String(45), default=None)  # Payload type (or deserialization method)

    sync_policy = Column(String(45), nullable=True, default=None)  # Synchronization Policy for this attribute

    item_id = Column(ForeignKey(ItemEntity.entity_id), nullable=False, index=True)  # ID of owner Item
    item = relationship(ItemEntity, foreign_keys=item_id, lazy="joined")  # Reference to item

    # UniqueConstraint
    __table_args__ = (UniqueConstraint('item_id', 'name', name='item_id_name_uk'),
                      )

    __mapper_args__ = {
        'polymorphic_identity': 'entity_attributes'
    }

    def get_type(self, resource_id: AnyStr):
        """
        Returns the type of this attribute.
        """
        with read_session(close=True) as db:
            return self.crud.read(db=db, entity_id=resource_id).type

    def get_value(self, attribute_id: AnyStr):
        """
        Returns the type of this attribute.
        """
        with read_session(close=True) as db:
            attribute = self.crud.read(db=db, entity_id=attribute_id)
            return {
                "value": attribute.value,
                "value_type": attribute.value_type
            }

    def get_item(self, resource_id: str) -> "ItemEntity":
        """
        Returns the item owning this attribute.
        """
        with read_session(close=True) as db:
            resource = self.crud.read(db=db, entity_id=resource_id)
            return jsonable_encoder(resource.item)
