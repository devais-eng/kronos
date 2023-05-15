from sqlalchemy import Column, ForeignKey
from sqlalchemy.orm import relationship
from kronos.storage.entity import Entity
from kronos.storage.entity.item import ItemEntity
from kronos.storage.view.relation import RelationView
from typing import Dict


def default_id(context):
    return f"{context.current_parameters.get('parent_id')}->{context.current_parameters.get('child_id')}"


class RelationEntity(Entity):
    """
    A relation maps hierarchies for items.
    """
    view = RelationView

    entity_id = Column(ForeignKey(Entity.id), primary_key=True, default=default_id)
    entity = relationship(Entity,
                          cascade="all, delete-orphan",
                          single_parent=True,
                          post_update=True)

    parent_id = Column(ForeignKey(ItemEntity.entity_id), nullable=False)
    child_id = Column(ForeignKey(ItemEntity.entity_id), nullable=False)

    parent = relationship(ItemEntity, primaryjoin='RelationEntity.parent_id == ItemEntity.entity_id')
    child = relationship(ItemEntity, primaryjoin='RelationEntity.child_id == ItemEntity.entity_id')

    __mapper_args__ = {
        'polymorphic_identity': 'entity_relation'
    }

    def create(self, to_create, close=False): # noqa
        """
        Override create function for custom relation's id.
        """
        to_create.id = f"{to_create.parent_id}->{to_create.child_id}"
        return super().create(
            to_create=to_create,
            close=close
        )

    def update(self, new: Dict, previous=None, close=False):
        """
        Override update function for custom relation's id.
        """
        parent_id = new.get("parent_id", None)
        child_id = new.get("child_id", None)
        if parent_id is not None and child_id is not None:
            new["id"] = f"{parent_id}->{child_id}"
        return super().update(
            new=new, previous=previous, close=close
        )
