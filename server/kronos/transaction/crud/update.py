from kronos.transaction.crud import CRUDTransaction
from kronos.transaction import TransactionType
from kronos.storage.entity import EntityType, get_by_type
from kronos.sync.transaction.conflicts import NoChangeUpdate


class UpdateTransaction(CRUDTransaction):
    _type = TransactionType.ENTITY_UPDATED

    def __init__(self, body: dict, entity_type: EntityType, entity_id: str, *arg, **kwargs):
        super(UpdateTransaction, self).__init__(body, entity_type, entity_id, *arg, **kwargs)

    def get_inverse(self):
        if self._revertable and len(self._revert) == 0:
            from kronos.transaction.crud.read import ReadTransaction
            if self._body is None:
                previous_data = ReadTransaction(self._entity_type, self._entity_id, {}).apply()
            else:
                previous_data = self._body
            self._revert = [UpdateTransaction(previous_data, self._entity_type, self._entity_id, revertable=False)]
        return self._revert

    def apply(self):
        """
        Apply an update transaction. If entity_id is not provided we generate it.
        :param storage: SQLAlchemy engine
        :return: The committed transaction
        """
        from kronos.transaction.crud.read import ReadTransaction
        if not self._entity_id:
            self._entity_id = f"{self._entity_type}#{self._entity_id}"

        self.get_inverse()

        entity = get_by_type(EntityType[self._entity_type])()  # noqa

        self._body["id"] = self._entity_id

        previous = ReadTransaction(self._entity_type, self._entity_id, {}).apply()
        if previous == self._body:
            raise NoChangeUpdate(self._entity_type, self._entity_id, self._body)
        entity.update(previous=previous, new=self._body)
        super(UpdateTransaction, self).apply()
