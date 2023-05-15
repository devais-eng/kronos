from sqlalchemy import UniqueConstraint, Column, Integer, ForeignKey, String
from sqlalchemy.orm import relationship
from kronos.storage.entity import BaseEntity
from kronos.graph.storage import GraphEntity


class NodeEntity(BaseEntity):
    graph_id = Column(ForeignKey(GraphEntity.id), nullable=True)
    graph = relationship(GraphEntity,
                         cascade="all, delete-orphan",
                         single_parent=True,
                         post_update=True
                         )  # Reference to resource

    node_ix = Column(Integer(), primary_key=True, autoincrement=True)

    out_edges = relationship("EdgeEntity", foreign_keys='EdgeEntity.source_node_id', back_populates="source_node",
                             cascade='save-update, merge, delete, delete-orphan')
    in_edges = relationship("EdgeEntity", foreign_keys='EdgeEntity.destination_node_id',
                            back_populates="destination_node",
                            cascade='save-update, merge, delete, delete-orphan')

    node_type = Column(String(50))

    # UniqueConstraint
    __table_args__ = (UniqueConstraint('graph_id', 'node_ix', name='graph_node_order'),)
    __mapper_args__ = {
        'polymorphic_identity': 'graph_node_base',
        'polymorphic_on': node_type
    }
