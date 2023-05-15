from kronos.graph.edge import GraphEdge


class Changelog(GraphEdge):

    def __init__(self,
                 source_version,
                 destination_version,
                 deltas):
        super(Changelog, self).__init__(source_version, destination_version, deltas, str(deltas))

    def to_json(self):
        pass

    @classmethod
    def from_json(cls):
        pass
