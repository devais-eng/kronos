from time import time
from kronos.graph.node import GraphNode


class Version(GraphNode):
    """
    A Version reference to a specific state on a versioning graph.
    """

    def __init__(self, node_ix=None, version_id: str = None):
        super(Version, self).__init__(node_ix, version_id)
        self._created_at = time()

    def __repr__(self):
        return self._node_label

    def __eq__(self, other):
        return self.__repr__() == other.__repr__()

    def to_dict(self):
        return {
            "node_ix": self._node_ix,
            "version_id" : self._node_label
        }

    @classmethod
    def from_dict(cls, dict_data):
        return Version(**dict_data)
