from sqlalchemy import Column, ForeignKey, String
from kronos.graph.storage.node import NodeEntity
from sqlalchemy.orm import relationship
from kronos.storage.entity import BaseEntity
from kronos.graph.storage import GraphEntity


class EdgeEntity(BaseEntity):
    # TODO: Check nodes in same graph and not circulat (soruce_node.graph_id == destnation_node.graph_id && source_node.id!= destination_node.id) # noqa
    source_node_id = Column(ForeignKey(NodeEntity.id))
    destination_node_id = Column(ForeignKey(NodeEntity.id))

    source_node = relationship(NodeEntity, foreign_keys=source_node_id)
    destination_node = relationship(NodeEntity, foreign_keys=destination_node_id)

    graph_id = Column(ForeignKey(GraphEntity.id), nullable=True)
    graph = relationship(GraphEntity,cascade="all, delete-orphan",
                          single_parent=True,
                          post_update=True)  # Reference to resource
    # TODO: Multiple edges are allowed between the same two nodes (it is discouraged)

    edge_type = Column(String(50))

    __mapper_args__ = {
        'polymorphic_identity': 'graph_edge_base',
        'polymorphic_on': edge_type
    }
