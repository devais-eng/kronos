from enum import Enum


class SyncPolicy(Enum):
    GATEWAY_TO_SERVER = "GATEWAY_TO_SERVER"
    SERVER_TO_GATEWAY = "SERVER_TO_GATEWAY"
    ALWAYS_SYNC = "ALWAYS_SYNC"
    NO_SYNC = "NO_SYNC"
