from sqlalchemy import Column, String, ForeignKey
from sqlalchemy.orm import relationship
from kronos.storage.entity import BaseEntity


class GraphEntity(BaseEntity):
    nodes = relationship("NodeEntity", back_populates="graph", cascade='save-update, merge, delete, delete-orphan')
    edges = relationship("EdgeEntity", back_populates="graph", cascade='save-update, merge, delete, delete-orphan')
    graph_type = Column(String(50))

    __mapper_args__ = {
        'polymorphic_identity': 'graph_base',
        'polymorphic_on': graph_type
    }

    @classmethod
    def new_graph(cls):
        from kronos.storage.utils import transactional_session
        new_graph = cls()
        with transactional_session() as session:
            session.add(new_graph)
            session.refresh(new_graph)
        return new_graph.id
