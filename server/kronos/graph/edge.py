from uuid import uuid4


class GraphEdge(object):

    def __init__(self, source_node, destination_node, weight, edge_label = None):
        self._source_node = source_node
        self._destination_ndoe = destination_node
        self._weight = weight
        self._edge_label = edge_label or str(uuid4())

    def edge_label(self):
        return self._edge_label

    def to_json(self):
        pass

    @classmethod
    def from_json(cls):
        pass