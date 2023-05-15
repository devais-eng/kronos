from datetime import datetime
from enum import Enum
from loguru import logger
import json
from time import time
from uuid import uuid4


class TransactionType(Enum):
    BASE = "BASE"
    CRUD = "CRUD"
    ENTITY_CREATED = "ENTITY_CREATED"
    ENTITY_UPDATED = "ENTITY_UPDATED"
    ENTITY_DELETED = "ENTITY_DELETED"
    ENTITY_READ = "ENTITY_READ"
    DELTA = "DELTA"


def get_by_type(type: TransactionType):
    if isinstance(type, str):
        type = TransactionType[type]
    if type == TransactionType.DELTA:
        from kronos.transaction.delta import DeltaTransaction
        return DeltaTransaction
    elif type == TransactionType.ENTITY_CREATED:
        from kronos.transaction.crud.create import CreateTransaction
        return CreateTransaction
    elif type == TransactionType.ENTITY_UPDATED:
        from kronos.transaction.crud.update import UpdateTransaction
        return UpdateTransaction
    elif type == TransactionType.ENTITY_DELETED:
        from kronos.transaction.crud.delete import DeleteTransaction
        return DeleteTransaction


class Transaction(object):
    """
        A transaction is a trackable operation that can be applied or reverted.
        Transactions may be stored in a log to recreate different versions of the environment.
    """

    _type = TransactionType.BASE

    def __init__(self, body, t_uuid=None, t_index=0, t_len=1, revertable=True, *arg, **kwargs):
        self.t_uuid = t_uuid or str(uuid4())
        self.t_index = t_index
        self.t_len = t_len
        self._ts_created = int(time())  # Time transaction was created
        self._ts_committed = None  # Time transaction was committed
        self._body = body  # Transaction body
        self._kwargs = kwargs
        self._logger = logger
        self._revert = []
        self._revertable = revertable

    def get_template(self, *args, **kwargs):
        return self.__class__(self._body, *args, **kwargs)

    def get_inverse(self, *args, **kwargs):
        raise NotImplemented()

    def commit(self, *args, **kwargs):
        """
            Commits operation, thus locking the transaction.
        """
        if self._ts_committed is not None:
            self.revert(*args, **kwargs)
            raise Exception("Cannot apply an already committed transaction.")
        self._ts_committed = datetime.now()
        self._logger.debug(f"In commit of {__name__} : {self._ts_committed}")

    def apply(self, *args, **kwargs):
        """
            Applies a transaction. (Usually defines revert logic too)
        """
        self.commit(*args, **kwargs)

    def revert(self, *args, **kwargs):
        """
            Reverts a transaction. Since not predictable, revert logic may be defined at commit time.
        """
        if self._ts_committed is None:
            raise Exception("Cannot revert an uncommitted transaction.")
        for r in self._revert[::-1]:
            r.apply(*args, **kwargs)

    def to_dict(self, must_revert=True) -> dict:
        """
            Serialize a transaction.
        """

        body_dict = self._body.copy() if not isinstance(self._body, Transaction) else self._body.to_dict()["body"]
        if "revertable" in body_dict:  # noqa
            del body_dict["revertable"]  # noqa
        data_dict = {
            "type": self._type.name,
            "ts_committed": int(self._ts_committed.timestamp()) if self._ts_committed is not None and not isinstance(
                self._ts_committed, int) else self._ts_committed,
            "ts_created": self._ts_created,
            "revert": [t.to_dict() for t in self._revert[::-1]] if must_revert else [],
            "body": body_dict
        }

        data_dict.update(self._kwargs)
        return data_dict

    def to_transaction(self):
        raise NotImplemented()

    @classmethod
    def from_dict(cls, data: dict):
        """
            Deserialize a transaction.
        """
        kwargs = [d for d in data if d not in ["body", "ts_created", "ts_committed", "type", "revert"]]
        transaction = get_by_type(data["type"])(data["body"], **{k: data[k] for k in kwargs})
        transaction._ts_created = data["ts_created"]
        transaction._ts_committed = data["ts_committed"]
        transaction._revert = [get_by_type(r["type"]).from_dict(r) for r in data["revert"]]
        return transaction

    def __repr__(self):
        return f"{self._type}@{self._ts_committed}"

    def __eq__(self, other):
        return self._body == other._body

    def __hash__(self):
        return hash(json.dumps(self._body))
