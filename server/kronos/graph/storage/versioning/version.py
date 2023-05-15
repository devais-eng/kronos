from kronos.graph.storage.node import NodeEntity
from sqlalchemy import Column, Integer, ForeignKey, JSON
from time import time
from kronos.graph.versioning.version import Version
import json
from kronos.storage.utils import transactional_session
from sqlalchemy.orm import relationship


class VersionNodeEntity(NodeEntity):
    node_id = Column(ForeignKey(NodeEntity.id), primary_key=True, nullable=False)
    version = Column(JSON(), default=lambda: Version().to_dict())
    created_at = Column(Integer, default=time)
    node = relationship(NodeEntity,
                        cascade="all, delete-orphan",
                        single_parent=True,
                        post_update=True
                        )

    __mapper_args__ = {
        'polymorphic_identity': 'graph_node_version'
    }

    @classmethod
    def get_or_create(cls, graph_id, version):
        with transactional_session() as session:
            if isinstance(version, Version):
                destination_node = session.query(VersionNodeEntity).filter(
                    VersionNodeEntity.node_ix == version._node_ix).first()
                if destination_node is None:
                    destination_node_id = cls.add_node(graph_id, version=version, node_ix=version._node_ix)
                    destination_node = session.query(VersionNodeEntity).filter(
                        VersionNodeEntity.id == destination_node_id).first()
                    # session.refresh(destination_node)
            else:
                destination_node = session.query(VersionNodeEntity).filter(VersionNodeEntity.version == version).first()
                if destination_node is None:
                    destination_node_id = cls.add_node(graph_id)
                    destination_node = session.query(VersionNodeEntity).filter(
                        VersionNodeEntity.id == destination_node_id).first()
            return destination_node.node_id

    @classmethod
    def add_node(cls, graph_id, version=None, created_at=None, node_ix=None):
        version_data = {
            "graph_id": graph_id
        }
        if version is not None:
            version_data["version"] = version.to_dict()
        else:
            version_data["version"] = Version().to_dict()
        if created_at is not None:
            version_data["created_at"] = created_at
        if node_ix is not None:
            version_data["node_ix"] = node_ix
        with transactional_session() as session:
            from fastapi.encoders import jsonable_encoder
            node = cls(**version_data)
            session.add(node)
            # session.refresh(node)
            session.flush()
            return node.id

    def set_transition(self, changelog, destination_version=None):
        from kronos.graph.storage.versioning.changelog import ChangelogEntity
        destination_node_id = self.get_or_create(self.graph_id, destination_version)

        changelog_data = {
            "source_node_id": self.id,
            "destination_node_id": destination_node_id,
            "transaction": json.dumps(changelog.to_dict()),
            "graph_id": self.graph_id
        }

        with transactional_session() as session:
            new_changelog = ChangelogEntity(**changelog_data)
            session.add(new_changelog)
            session.flush()
            return new_changelog.edge_id

    def to_version(self):
        from kronos.graph.versioning.version import Version
        version_data = self.version
        version_data.update({"node_ix": self.node_ix})
        return Version.from_dict(version_data)
