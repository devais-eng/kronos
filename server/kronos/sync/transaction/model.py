import uuid
from faust_utils.model import Request, ErrorResponse
from faust import Record
from typing import AnyStr, Optional, Dict, List
from kronos.transaction import TransactionType
from fastapi.encoders import jsonable_encoder
import time
from enum import Enum


class TransactionErrorCode(Enum):
    """
    Specialize a code error.
    """
    ENTITY_NOT_FOUND = 0
    ENTITY_ALREADY_EXISTS = 1
    GENERAL = 2
    INTEGRITY_ERROR = 3
    REDIS_ERROR = 4


class TransactionErrorResponse(ErrorResponse):
    """
    Specialized response for errored requests.
    """

    def __abstract_init__(self) -> None:
        pass

    error_code: TransactionErrorCode


class TransactionCommand(Enum):
    CRUD = "CRUD"


class TransactionModel(Record):

    def __abstract_init__(self) -> None:
        pass

    tx_type: AnyStr
    entity_id: AnyStr
    id: int
    entity_type: AnyStr
    timestamp: int
    triggered_by: Optional[AnyStr]
    tx_uuid: Optional[AnyStr]
    tx_len: Optional[int]
    tx_index: Optional[int]

    body: Optional[Dict] = None


class TransactionRequest(Request):
    def __abstract_init__(self) -> None:
        pass

    payload: List[TransactionModel]


def to_request(model, transaction_id, entity_type, entity_id, tx_index, tx_len):
    from kronos.storage.view import EntityCreateModel, EntityDeleteModel, EntityUpdateModel, EntityPatchModel

    if isinstance(model, EntityCreateModel):
        transaction_type = TransactionType.ENTITY_CREATED
    elif isinstance(model, EntityUpdateModel) or isinstance(model, EntityPatchModel):
        transaction_type = TransactionType.ENTITY_UPDATED
    elif isinstance(model, EntityDeleteModel):
        transaction_type = TransactionType.ENTITY_DELETED
    else:
        raise Exception(f"Can't convert model {model} of type {type(model)} to a CRUD request.")

    return TransactionModel(
        tx_type=transaction_type,
        entity_id=entity_id,
        id=transaction_id,
        entity_type=entity_type,
        tx_uuid=str(uuid.uuid1()),
        tx_len=tx_len,
        tx_index=tx_index,
        timestamp=time.time(),
        body=jsonable_encoder(model)
    )
