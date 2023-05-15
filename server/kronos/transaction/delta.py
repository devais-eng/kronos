from kronos.transaction import TransactionType, Transaction
from typing import List


class DeltaTransaction(Transaction):
    _type = TransactionType.DELTA

    def __init__(self, body: List[Transaction], *arg, **kwargs):
        super(DeltaTransaction, self).__init__(body, *arg, **kwargs)

    def get_inverse(self):
        if self._revertable and len(self._revert) == 0:
            self._revert = []
            for t in self._body:
                self._revert.extend(t.get_inverse())
        return self._revert

    def apply(self):
        """
        Iterate and apply all transaction of this delta. Finally apply DeltaTransaction to validate.
        :param storage: Db-Session
        :return:
        """
        applied = []
        for t in self._body:
            try:
                t.apply()
                applied.append(t)
            except Exception as e:
                for t in applied[::-1]:
                    t.revert()
                raise e
        super(DeltaTransaction, self).apply()

    def to_dict(self, must_revert=True) -> dict:
        """
            Serialize a transaction.
        """
        data_dict = {
            "type": self._type.name,
            "ts_committed": int(self._ts_committed.timestamp()) if self._ts_committed is not None and not isinstance(
                self._ts_committed, int) else self._ts_committed,
            "ts_created": self._ts_created,
            "revert": [t.to_dict() for t in self.get_inverse()] if must_revert else [],
            "body": [t.to_dict() for t in self._body]
        }
        data_dict.update(self._kwargs)
        return data_dict

    @classmethod
    def from_dict(cls, data: dict):
        """
            Deserialize a transaction.
        """
        from kronos.transaction import get_by_type
        kwargs = [d for d in data if d not in ["body", "ts_created", "ts_committed", "type", "revert"]]
        transaction = get_by_type(data["type"])(body=[get_by_type(r["type"]).from_dict(r) for r in data["body"]],
                                                **{k: data[k] for k in kwargs})
        transaction._ts_created = data["ts_created"]
        transaction._ts_committed = data["ts_committed"]
        transaction._revert = [get_by_type(r["type"]).from_dict(r) for r in data["revert"]]
        return transaction

    def to_transaction(self):
        models = []

        for transaction in self._body:
            models.extend(transaction.to_transaction())

        new_models = []
        for ix, m in enumerate(models):
            m.update({
                "tx_uuid": self.t_uuid,
                "tx_len": len(models),
                "tx_index": ix})
            new_models.append(m)
        return new_models
