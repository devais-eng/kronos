import networkx as nx
from typing import Generic, TypeVar, Dict, Tuple, Type
from loguru import logger
from kronos.graph.utils import visualize_graph

from kronos.graph.edge import GraphEdge
from kronos.graph.node import GraphNode

NodeType = TypeVar("NodeType", bound=GraphNode)
EdgeType = TypeVar("EdgeType", bound=GraphEdge)


class Graph(Generic[NodeType, EdgeType]):
    """
        A Graph is a structure containing nodes and edges.
    """

    def __init__(self):
        """
        Initialize a new graph.
        """
        self._graph: nx.Graph = nx.Graph()
        self._nodes: Dict[int, Type[NodeType]] = {}
        self._edges: Dict[Tuple[int, int], Type[EdgeType]] = {}
        self._logger = logger

    @classmethod
    def new_node(cls):
        from kronos.graph.node import GraphNode
        return GraphNode()

    def _new_node(self, node: Type[NodeType] = None) -> Tuple[int, Type[NodeType]]:
        """
        Create a new node.

        :param node: Node data, if not assigned is a random UUID.
        :return: node.
        """
        if node is None:
            node = self.new_node()

        node_ix = node.node_index()

        if node_ix is None:
            node_ix = 0
            while True:
                if node_ix not in self._nodes:
                    break
                node_ix += 1
        elif node_ix in self._nodes:
            raise Exception(f"Already added node with index {node_ix} in current graph.")
        self._nodes[node_ix] = node
        self._nodes[node_ix].set_index(node_ix)
        return node_ix, node

    def _new_transition(self,
                        from_node_ix: int,
                        to_node_ix: int = None,
                        transition_data: Type[EdgeType] = None
                        ):
        """
        Apply new deltas and create transitions.

        :param deltas: A list of deltas.
        :return: the updated graph.
        """
        to_node_ix = to_node_ix or self._new_node()[0]
        self._edges[(from_node_ix, to_node_ix)] = transition_data
        #print(from_node_ix,to_node_ix, transition_data)
        self._graph.add_edges_from([(from_node_ix, to_node_ix)])

    def visualize(self, *args, **kwargs):
        """
        Display the graph content.

        :return: Display the version graph.
        """
        if len(self._nodes.keys()) > 1:
            node_labels = {ix: n.node_label() for ix, n in self._nodes.items()}
            edge_labels = {ix: e.edge_label() for ix, e in self._edges.items()}
            visualize_graph(self._graph, node_labels, edge_labels, *args, **kwargs)
        elif len(self._nodes.keys()) == 1:
            self._logger.info("Single node graph can't be displayed.")
        else:
            self._logger.info("Graph is empty.")

