from uuid import uuid4
from time import time


def gen_random_item():
    return {
        "name": str(uuid4()),
        "type": "type",
        "customer_id": "string",
        "sync_policy": "string",
        "edge_mac": "string",
        "id": "item1",
        "source_timestamp": int(time()),
        "created_by": "string"
    }


def gen_random_res(item_id):
    return {

    }