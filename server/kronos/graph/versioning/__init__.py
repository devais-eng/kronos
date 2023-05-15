from networkx.algorithms.shortest_paths.generic import shortest_path
from kronos.graph.versioning.version import Version
from kronos.transaction import Transaction
from kronos.transaction.delta import DeltaTransaction
from kronos.graph import Graph
from typing import List, Tuple, Any
from kronos.graph.versioning.changelog import Changelog
from kronos.graph.storage.versioning import VersioningGraphEntity
from kronos.storage.utils import transactional_session
from kronos.sync.transaction.conflicts import VersionNotFound


class VersioningGraph(Graph[Version, Transaction]):
    """
    A Version Graph links versions and transactions in order to navigate through
    different states of tracked entities.
    """

    def __init__(self,
                 nodes: List[Version] = [Version("root")],
                 edges: List[Tuple[Version, Version, DeltaTransaction]] = [],
                 master: List[Version] = [],
                 entity=None,
                 sync=False,
                 ):
        """
        Initialize a new version graph.

        :param root: Optional, root name.
        """
        self._entity_id = entity
        super(VersioningGraph, self).__init__()
        self._master = master
        for n in nodes:
            self._new_node(n)
        for e in edges:
            self._new_transition(e[2], e[0], e[1], sync=sync)
        self._changelog = DeltaTransaction(body=[])

    @property
    def master(self) -> str:
        """
        The current (master) version.
        """
        return self._nodes[self._master[-1]]

    def visualize(self):
        super(VersioningGraph, self).visualize(master=self.master)

    def _new_node(self, node: Version = None) -> Tuple[int, Any]:
        """
        Creates a new node ahead master node.

        :return: new node.
        """
        node_ix, node = super(VersioningGraph, self)._new_node(node)
        self._master.append(node_ix)
        return node_ix, node

    def _new_transition(self, changes: Transaction, source=None, destination=None, sync=True):
        """
        Apply new deltas and create transitions.

        :param deltas: A list of deltas.
        :return: the updated graph.
        """
        if source is not None:
            source_ix = source.node_index()
        else:
            source_ix = self._master[-1]
        if destination is not None:
            destination_ix = destination.node_index()
        else:
            destination_ix = self._new_node()[0]
        edge = (source_ix, destination_ix)
        super(VersioningGraph, self)._new_transition(*edge, Changelog(source_ix, destination_ix, changes))
        self._graph.add_edges_from([edge])
        if sync:
            with transactional_session() as session:
                graph = session.query(VersioningGraphEntity).filter(
                    VersioningGraphEntity.graph_id == self._entity_id).first()
                if graph is not None:
                    graph.add_version(changes, self._nodes[source_ix], self._nodes[destination_ix])
                session.flush()
            # self._entity.add_version(changes, self._nodes[source_ix], self._nodes[destination_ix])

    def assign_version(self):
        """
        Commits applied transactions and assign version.

        :param storage: The storage where to apply transactions.
        :return: The applied changes.
        """
        from kronos.graph.versioning import VersioningGraphEntity
        deltas = self._changelog.get_template()
        deltas.apply()
        self._new_transition(deltas)
        self._changelog = DeltaTransaction(body=[])
        with transactional_session() as session:
            graph = session.query(VersioningGraphEntity).filter(
                VersioningGraphEntity.graph_id == self._entity_id).first()
            if graph is not None:
                graph.set_master(self._master)
                session.add(graph)
                session.flush()
                return deltas, VersioningGraphEntity.to_graph(graph_id=graph.id).master.__repr__()
            return deltas, None

    def apply(self, changes: Transaction):
        """
        Enqueue a transaction to this version graph. It must be committed explicitely by calling assign_version.

        :param changes:
        :return:
        """
        if not isinstance(changes, list):
            changes = [changes]
        self._changelog._body.extend(changes)
        return self

    def revert(self):
        """
        Reverts the reference to master. # TODO

        :return:
        """
        if len(self._master) > 1:
            self._logger.info(f"Reverting from {self._master}")
            self._master.pop(-1)
            with transactional_session() as session:
                graph = session.query(VersioningGraphEntity).filter(
                    VersioningGraphEntity.graph_id == self._entity_id).first()
                graph.set_master(self._master)

    def checkout(self, version: Version):
        """

        :param version:
        :return:
        """
        version_node = None
        for n, n_v in self._nodes.items():
            if n_v == version:
                version_node = n
        if version_node is None:
            raise VersionNotFound(version)
        path = shortest_path(self._graph, self._master[-1], version_node.node_index())
        transactions = []
        actual_node = self._master[-1]
        for n in path[1:]:
            to_apply = None
            if (actual_node, n) in self._edges:
                to_apply = self._edges[(actual_node, n)]
                transactions.append(to_apply)
                self._edges.append(n)
            else:
                to_apply = self._edges[(n, actual_node)].get_inverse()
                transactions.append(to_apply)
                self.revert()
            self.apply(to_apply)
            actual_node = n
        return transactions

    @classmethod
    def new_node(cls):
        from kronos.graph.versioning.version import Version
        return Version()
