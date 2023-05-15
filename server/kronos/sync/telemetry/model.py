from faust_utils.model import Request, ErrorResponse
from faust import Record
from typing import AnyStr, Optional, Dict, Union
from enum import Enum


class TelemetryCommand(Enum):
    CONNECTION = "CONNECTION"
    DISCONNECTION = "DISCONNECTION"


class TelemetryPayload(Record):

    def __abstract_init__(self) -> None:
        pass

    hostname: AnyStr
    applicationUptime: int
    systemUptime: AnyStr
    is_in_docker: Optional[AnyStr]
    battery: Optional[Dict[AnyStr, AnyStr]]
    timestamp_utc: int
    timestamp_local: int
    last_sync_ts: int
    last_connection_ts: int
    last_received_message_ts: int
    db_file_size: int
    items_count: int
    attributes_count: int
    relations_count: int


class Telemetry(Record):

    def __abstract_init__(self) -> None:
        pass

    device_id: AnyStr
    timestamp: int
    telemetry: Optional[Dict]


class TelemetryRequest(Request):

    def __abstract_init__(self) -> None:
        pass

    payload: Telemetry
