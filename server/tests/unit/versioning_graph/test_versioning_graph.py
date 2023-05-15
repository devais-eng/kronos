from kronos.deps import get_db
from kronos.transaction.crud.create import CreateTransaction
from kronos.transaction.delta import DeltaTransaction
from kronos.storage.entity import EntityType
from uuid import uuid1
from kronos.storage import init_db, SessionFactory as Session
from kronos.graph.versioning import VersioningGraph

class TestGraph():
    db = next(get_db())
    init_db(db)
    graph = VersioningGraph()
    graph.visualize()
    transaction = CreateTransaction(entity_type=EntityType.ITEM.name, entity_id=str(uuid1()),
                              body={"name": str(uuid1()),
                                    "type": "str",
                                    "sync_policy": None,
                                    "customer_id": None,
                                    "created_by": "Test"})