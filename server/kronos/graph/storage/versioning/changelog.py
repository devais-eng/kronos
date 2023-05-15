from kronos.graph.storage.edge import EdgeEntity, ForeignKey
from sqlalchemy.orm import relationship
from sqlalchemy import Column, JSON


class ChangelogEntity(EdgeEntity):
    edge_id = Column(ForeignKey(EdgeEntity.id), primary_key=True, nullable=False)
    edge = relationship(EdgeEntity,
                        cascade="all, delete-orphan",
                        single_parent=True,
                        post_update=True
                        )

    transaction = Column(JSON(), nullable=False)

    __mapper_args__ = {
        'polymorphic_identity': 'graph_edge_changelog'
    }

    def get_transaction(self, edge_id):
        from kronos.transaction.delta import DeltaTransaction
        from kronos.storage.utils import transactional_session
        import json
        with transactional_session() as session:
            changelog = session.query(ChangelogEntity).filter(ChangelogEntity.edge_id == edge_id).first()
            return DeltaTransaction.from_dict(json.loads(changelog.transaction))
