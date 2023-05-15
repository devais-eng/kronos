from sqlalchemy import Column, String, ForeignKey, JSON, Boolean, Text
from sqlalchemy.orm import relationship
from kronos.storage.entity import BaseEntity
from kronos.graph.storage.versioning import VersioningGraphEntity
from fastapi.encoders import jsonable_encoder
from kronos.transaction import Transaction


class ConflictEntity(BaseEntity):
    graph_id = Column(ForeignKey(VersioningGraphEntity.graph_id), primary_key=True, nullable=False)
    graph = relationship(
        VersioningGraphEntity,
        cascade="all, delete-orphan",
        single_parent=True,
        post_update=True
    )
    transaction = Column(JSON(), nullable=False)
    role = Column(String(45), nullable=True)
    version = Column(String(255), nullable=True)
    conflict = Column(Text, nullable=False)
    solved = Column(Boolean, default=False)

    @classmethod
    def new_conflict(cls, graph_id, transaction: Transaction, role, version, conflict):
        from kronos.storage.utils import transactional_session
        with transactional_session() as session:
            new_conflict = ConflictEntity(graph_id=graph_id, transaction=transaction.to_dict(must_revert=False),
                                          role=str(role),
                                          version=version, conflict=conflict)
            session.add(new_conflict)
            session.flush()
            session.refresh(new_conflict)
            return new_conflict.id

    @classmethod
    def get_unsolved(cls):
        from kronos.storage.utils import transactional_session
        with transactional_session() as session:
            unsolved = session.query(cls).filter(
                cls.solved == False).all()
            return jsonable_encoder(unsolved)
