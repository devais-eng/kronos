import uuid
from kronos.transaction.crud import CRUDTransaction
from kronos.transaction import TransactionType
from kronos.storage.entity import EntityType, get_by_type
from kronos.sync.transaction.conflicts import CreateOnExistingEntity
from fastapi.encoders import jsonable_encoder


class CreateTransaction(CRUDTransaction):
    _type = TransactionType.ENTITY_CREATED

    def __init__(self, body: dict, entity_type: EntityType, entity_id: str, *arg, **kwargs):
        super(CreateTransaction, self).__init__(body, entity_type, entity_id, *arg, **kwargs)

    def get_inverse(self):
        if self._revertable and len(self._revert) == 0:
            from kronos.transaction.crud.delete import DeleteTransaction
            self._revert = [DeleteTransaction({}, self._entity_type, self._entity_id, revertable=False)]
        return self._revert

    def apply(self):
        from kronos.transaction.crud.read import ReadTransaction
        entity = get_by_type(EntityType[self._entity_type])()  # noqa

        if not self._entity_id:
            self._entity_id = str(uuid.uuid4())

        if isinstance(self._body, ReadTransaction):  # Reverse of delete (create from previous read)
            payload = self._body.apply()
        else:
            payload = self._body
            previous_data = entity.read(entity_id=self._entity_id, raise_not_found=False)
            self._logger.info(f"Entity founded is {previous_data}")
            if previous_data is not None:
                if not previous_data.active:
                    self._logger.info(f"Entity with id {self._entity_id} is soft deleted.")
                    entity.update(previous=previous_data, new={"active": True})
                    super(CreateTransaction, self).apply()
                    self.get_inverse()
                    return
                else:
                    self._logger.info(f"Exception ")
                    raise CreateOnExistingEntity(self._entity_type, self._entity_id, self._body,
                                                 jsonable_encoder(previous_data))

        payload['id'] = self._entity_id
        self._logger.debug(f"Executing transaction for {self._entity_id} of type {self._entity_type}")

        to_create = entity.view.create()(
            **payload
        )
        entity.create(to_create)
        super(CreateTransaction, self).apply()
        self.get_inverse()
