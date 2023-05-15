from kronos.transaction.crud import CRUDTransaction
from kronos.transaction import TransactionType
from kronos.storage.entity import EntityType, get_by_type
from kronos.sync.transaction.conflicts import ReadSyncEntityNotExists, ReadSyncMismatch
from kronos.exception import EntityNotFoundException
from fastapi.encoders import jsonable_encoder


class ReadTransaction(CRUDTransaction):
    _type = TransactionType.ENTITY_READ

    def __init__(self, entity_type: EntityType, entity_id: str, body={}, *arg, **kwargs):
        super(ReadTransaction, self).__init__(body, entity_type, entity_id, *arg, **kwargs)

    def get_inverse(self):
        return [self]

    def apply(self):
        entity = get_by_type(EntityType[self._entity_type])()  # noqa

        try:
            previous_data = jsonable_encoder(entity.read(entity_id=self._entity_id, raise_not_found=True))
        except EntityNotFoundException:
            raise ReadSyncEntityNotExists(self._entity_type, self._entity_id)

        if self._body is None or len(self._body) == 0 or "hard" in self._body:
            self._body = previous_data
            if self._revertable:
                self._revert = self.get_inverse()
        else:
            if previous_data != self._body:
                raise ReadSyncMismatch(self._entity_type, self._entity_id, self._body, previous_data)

        return previous_data
