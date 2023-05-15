from kronos.graph.storage import GraphEntity
from sqlalchemy import Column, String, ForeignKey
from sqlalchemy.orm import relationship


class VersioningGraphEntity(GraphEntity):
    graph_id = Column(ForeignKey(GraphEntity.id), primary_key=True, nullable=False)
    graph = relationship(
        GraphEntity,
        cascade="all, delete-orphan",
        single_parent=True,
        post_update=True
    )
    master = Column(String(255), nullable=True, default="")

    __mapper_args__ = {
        'polymorphic_identity': 'graph_versioning'
    }

    @classmethod
    def new_graph(cls):
        from kronos.storage.utils import transactional_session
        new_graph = cls()
        with transactional_session() as session:
            session.add(new_graph)
            session.flush()
            session.refresh(new_graph)
            from kronos.graph.storage.versioning.version import VersionNodeEntity
            new_node_id = VersionNodeEntity.add_node(new_graph.graph_id)
            destination_node = session.query(VersionNodeEntity).filter(
                VersionNodeEntity.id == new_node_id).first()
            new_graph.set_master([destination_node.node_ix])
            # session.refresh(new_graph)
            return new_graph.graph_id

    def add_version(self, changelog, origin_version, new_version):
        from kronos.graph.storage.versioning.version import VersionNodeEntity
        start_node_id = VersionNodeEntity.get_or_create(self.graph_id, origin_version)
        destination_node_id = VersionNodeEntity.get_or_create(self.graph_id, new_version)
        from kronos.storage.utils import transactional_session
        with transactional_session() as session:
            start_node = session.query(VersionNodeEntity).filter(VersionNodeEntity.node_id == start_node_id).first()
            destination_node = session.query(VersionNodeEntity).filter(
                VersionNodeEntity.node_id == destination_node_id).first()
            start_node.set_transition(changelog, destination_node.to_version())
            session.flush()

    def set_master(self, new_master):
        self.master = "/".join(list(map(str, new_master)))

    @classmethod
    def to_graph(cls, graph_id):
        from kronos.graph.versioning import VersioningGraph
        nodes = []
        edges = []
        from kronos.storage.utils import transactional_session
        with transactional_session(close=False) as session:
            graph = session.query(VersioningGraphEntity).filter(VersioningGraphEntity.id == graph_id).first()
            master = graph.master.split("/")
            for n in graph.nodes:
                nodes.append(n.to_version())
            for e in graph.edges:
                edges.append(
                    (e.source_node.to_version(), e.destination_node.to_version(), e.get_transaction(edge_id=e.id)))
            return VersioningGraph(nodes, edges, master, graph_id)
