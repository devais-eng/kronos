from kronos.transaction.crud import CRUDTransaction
from kronos.transaction import TransactionType
from kronos.storage.entity import EntityType, get_by_type
from kronos.sync.transaction.conflicts import EntityAlreadyDeleted


class DeleteTransaction(CRUDTransaction):
    _type = TransactionType.ENTITY_DELETED

    def __init__(self, body: dict, entity_type: EntityType, entity_id: str, *arg, **kwargs):
        super(DeleteTransaction, self).__init__(body, entity_type, entity_id, *arg, **kwargs)

    def get_inverse(self, expected=None):
        if self._revertable and len(self._revert) == 0:
            from kronos.transaction.crud.create import CreateTransaction
            from kronos.transaction.crud.read import ReadTransaction
            if expected is None:
                previous_data = ReadTransaction(self._entity_type, self._entity_id, self._body).apply()
            else:
                previous_data = expected
            self._revert = [CreateTransaction(previous_data, self._entity_type, self._entity_id, revertable=False)]
        return self._revert

    def apply(self):
        from kronos.transaction.crud.read import ReadTransaction
        if not self._entity_id:
            self._entity_id = f"{self._entity_type}#{self._entity_id}"

        if isinstance(self._body, ReadTransaction):  # Reverse of delete (create from previous read)
            expected = self._body.apply()
        else:
            expected = ReadTransaction(self._entity_type, self._entity_id, {}).apply()

        self.get_inverse(expected)
        hard_delete = True if self._body is None else self._body.get("hard", True)
        self._body = {}  # noqa
        self._body["id"] = self._entity_id
        entity = get_by_type(EntityType[self._entity_type])()  # noqa

        if hard_delete:
            entity.remove(entity_id=self._entity_id)
        else:
            if not expected['active']:
                raise EntityAlreadyDeleted(self._entity_type, self._entity_id)
            entity.update(new={"active": False, "id": self._entity_id})
        super(DeleteTransaction, self).apply()
