from kronos.storage.entity import EntityType
from kronos.transaction import Transaction, TransactionType
import time


class CRUDTransaction(Transaction):
    """
        A CRUD Transaction is an operation over a storage.
    """

    _type = TransactionType.CRUD

    def __init__(self, body: dict, entity_type: EntityType, entity_id: str, *arg, **kwargs):
        self._entity_type = entity_type
        self._entity_id = entity_id
        super(CRUDTransaction, self).__init__(body, entity_type=entity_type, entity_id=entity_id, *arg,
                                              **kwargs)

    def apply(self):
        """
            Applies transaction logic over  a storage.
        """
        return super(CRUDTransaction, self).apply()

    def __eq__(self, other):
        return self._entity_type == other._entity_type and self._entity_id == other._entity_id and self._body == other._body

    def to_transaction(self):
        transaction_model = {
            "entity_type": self._entity_type,
            "entity_id": self._entity_id,
            "id": self._entity_id,
            "timestamp": time.time(),
            "triggered_by": None,
            "body": self._body,
            "tx_type": self._type.name,
            "tx_uuid": self.t_uuid,
            "tx_len": self.t_len,
            "tx_index": self.t_index,

        }
        return [transaction_model]
