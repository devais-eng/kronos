from faust import Record
from typing import AnyStr, Dict
from kronos.storage.entity import EntityType


class VersionGraphResponse(Record):
    def __abstract_init__(self) -> None:
        pass

    entity_type: EntityType
    entity_id: AnyStr
    version: AnyStr
    action: AnyStr
    timestamp: int
    # change_log: Dict
    payload: Dict
