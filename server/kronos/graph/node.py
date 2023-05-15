from uuid import uuid4


class GraphNode(object):

    def __init__(self, node_ix=None, node_label=None):
        self._node_ix = node_ix
        self._node_label = node_label or str(uuid4())

    def node_label(self):
        return self._node_label

    def node_index(self):
        return self._node_ix

    def set_index(self, ix):
        self._node_ix = ix

    def to_json(self):
        pass

    @classmethod
    def from_json(cls):
        pass