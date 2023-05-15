from kronos.transaction.crud.create import CreateTransaction
from kronos.transaction.crud.delete import DeleteTransaction
from kronos.transaction.crud.read import ReadTransaction
from kronos.transaction.crud.update import UpdateTransaction
from kronos.transaction.delta import DeltaTransaction
from kronos.storage.entity import EntityType
from kronos.storage import init_db, SessionFactory as Session
import yaml
from fastapi.encoders import jsonable_encoder
from kronos.storage.entity.item import ItemEntity
from kronos.storage.entity.relation import RelationEntity
from kronos.storage.entity.resource import ResourceEntity
from kronos.deps import get_db

class TestTransactions:
    db = next(get_db())
    init_db(db)
    item_entity = ItemEntity()
    relation_entity = RelationEntity()
    resource_entity = ResourceEntity()
    with open('params_all_transactions.yaml') as file:
        var = yaml.load(file, Loader=yaml.FullLoader)

    item1 = {k: v for (k, v) in var['item1'].items()}
    item2 = {k: v for (k, v) in var['item2'].items()}
    item3 = {k: v for (k, v) in var['item3'].items()}
    updated_item1 = {k: v for (k, v) in var['updated_item1'].items()}
    relation1 = {k: v for (k, v) in var['relation1'].items()}
    updated_relation1 = {k: v for (k, v) in var['updated_relation1'].items()}
    resource1 = {k: v for (k, v) in var['resource1'].items()}
    updated_resource1 = {k: v for (k, v) in var['updated_resource1'].items()}

    hard = {"hard": True}

    def test_create_item(self, item):
        transaction = CreateTransaction(entity_type=EntityType.ITEM.name, entity_id=item['id'], body=item)
        transaction.apply(storage=self.db)
        stored_data = self.item_entity.crud.read(db=self.db, entity_id=item['id'], raise_not_found=False)
        assert stored_data.id == item['id']

    def test_create_relation(self, relation):
        transaction = CreateTransaction(entity_type=EntityType.RELATION.name, entity_id=relation['id'], body=relation)
        transaction.apply(storage=self.db)
        stored_data = self.relation_entity.crud.read(db=self.db, entity_id=relation['id'], raise_not_found=False)
        assert stored_data.id == relation['id']

    def test_create_resource(self, resource):
        transaction = CreateTransaction(entity_type=EntityType.RESOURCE.name, entity_id=resource['id'], body=resource)
        transaction.apply(storage=self.db)
        stored_data = self.resource_entity.crud.read(db=self.db, entity_id=resource['id'], raise_not_found=False)
        assert stored_data.id == resource['id']

    def test_read_item(self, item):
        transaction = ReadTransaction(entity_type=EntityType.ITEM.name, entity_id=item['id'])
        transaction.apply(storage=self.db)
        stored_data = self.item_entity.crud.read(db=self.db ,entity_id=item['id'], raise_not_found=False)
        assert stored_data.id == item['id']

    def test_read_relation(self, relation):
        transaction = ReadTransaction(entity_type=EntityType.RELATION.name, entity_id=relation['id'])
        transaction.apply(storage=self.db)
        stored_data = self.relation_entity.crud.read(db=self.db ,entity_id=relation['id'], raise_not_found=False)
        print(f"\n READ RELATION:\n{jsonable_encoder(stored_data)}\n")
        assert stored_data.id == relation['id']

    def test_read_resource(self, resource):
        transaction = ReadTransaction(entity_type=EntityType.RESOURCE.name, entity_id=resource['id'])
        transaction.apply(storage=self.db)
        stored_data = self.resource_entity.crud.read(db=self.db ,entity_id=resource['id'], raise_not_found=False)
        assert stored_data.id == resource['id']

    def test_update_item(self, item, updated_item):
        transaction = UpdateTransaction(entity_type=EntityType.ITEM.name, entity_id=item['id'], body=updated_item)
        transaction.apply(storage=self.db)
        stored_data = self.item_entity.crud.read(db=self.db ,entity_id=item['id'], raise_not_found=False)
        assert stored_data.name == updated_item['name']

    def test_update_relation(self, relation, updated_relation):
        pino = self.relation_entity.crud.read(db=self.db, entity_id=relation['id'], raise_not_found=False)
        print(f"\npino:{jsonable_encoder(pino)}")
        transaction = UpdateTransaction(entity_type=EntityType.RELATION.name, entity_id=relation['id'],
                                        body=updated_relation)
        transaction.apply(storage=self.db)
        stored_data = self.relation_entity.crud.read(db=self.db ,entity_id=relation['id'], raise_not_found=False)
        print(f"\n UPDATE RELATION:\n{jsonable_encoder(stored_data)}\n")
        assert stored_data.modified_by == updated_relation['modified_by']

    def test_update_resource(self, resource, updated_resource):
        transaction = UpdateTransaction(entity_type=EntityType.RESOURCE.name, entity_id=resource['id'],
                                        body=updated_resource)
        transaction.apply(storage=self.db)
        stored_data = self.resource_entity.crud.read(db=self.db ,entity_id=resource['id'], raise_not_found=False)
        assert stored_data.name == updated_resource['name']

    def test_delete_item(self, item):
        transaction = DeleteTransaction(entity_type=EntityType.ITEM.name, entity_id=item['id'], body=item)
        transaction.apply(storage=self.db)
        stored_data = self.item_entity.crud.read(db=self.db, entity_id=item['id'], raise_not_found=False)
        assert stored_data is None

    def test_delete_resource(self, resource):
        transaction = DeleteTransaction(entity_type=EntityType.RESOURCE.name, entity_id=resource['id'], body=resource)
        transaction.apply(storage=self.db)
        stored_data = self.resource_entity.crud.read(db=self.db, entity_id=resource['id'], raise_not_found=False)
        assert stored_data is None

    def test_delete_relation(self, relation):
        transaction = DeleteTransaction(entity_type=EntityType.RELATION.name, entity_id=relation['id'], body=relation)
        transaction.apply(storage=self.db)
        stored_data = self.relation_entity.crud.read(db=self.db, entity_id=relation['id'], raise_not_found=False)
        assert stored_data is None

    def test_create_then_revert_item(self, item):
        transaction = CreateTransaction(entity_type=EntityType.ITEM.name, entity_id=item['id'], body=item)
        transaction.apply(storage=self.db)
        stored_data = self.item_entity.crud.read(db=self.db, entity_id=item['id'], raise_not_found=False)
        assert stored_data.id == item['id']
        transaction.revert(storage=self.db)
        stored_data = self.item_entity.crud.read(db=self.db, entity_id=item['id'], raise_not_found=False)
        assert stored_data.active is False

    def test_create_then_revert_relation(self, relation):
        transaction = CreateTransaction(entity_type=EntityType.RELATION.name, entity_id=relation['id'], body=relation)
        transaction.apply(storage=self.db)
        stored_data = self.relation_entity.crud.read(db=self.db, entity_id=relation['id'], raise_not_found=False)
        assert stored_data.id == relation['id']
        transaction.revert(storage=self.db)
        stored_data = self.relation_entity.crud.read(db=self.db, entity_id=relation['id'], raise_not_found=False)
        assert stored_data.active is False

    def test_create_then_revert_resource(self, resource):
        transaction = CreateTransaction(entity_type=EntityType.RESOURCE.name, entity_id=resource['id'], body=resource)
        transaction.apply(storage=self.db)
        stored_data = self.resource_entity.crud.read(db=self.db, entity_id=resource['id'], raise_not_found=False)
        assert stored_data.id == resource['id']
        transaction.revert(storage=self.db)
        stored_data = self.resource_entity.crud.read(db=self.db, entity_id=resource['id'], raise_not_found=False)
        assert stored_data.active is False

    def test_update_then_revert_item(self, item, updated_item):
        print(jsonable_encoder(self.item_entity.crud.read(db=self.db, entity_id=item['id'], raise_not_found=False)))
        transaction = UpdateTransaction(entity_type=EntityType.ITEM.name, entity_id=item['id'], body=updated_item)
        transaction.apply(storage=self.db)
        stored_data = self.item_entity.crud.read(db=self.db, entity_id=item['id'], raise_not_found=False)
        assert stored_data.name == updated_item['name']
        transaction.revert(storage=self.db)
        stored_data = self.item_entity.crud.read(db=self.db, entity_id=item['id'], raise_not_found=False)
        print(f"\n---\n{jsonable_encoder(stored_data)}\n---\n")
        assert stored_data.name == item['name']

    def test_delete_then_revert_item(self, item):
        transaction = DeleteTransaction(entity_type=EntityType.ITEM.name, entity_id=item['id'], body=item)
        transaction.apply(storage=self.db)
        stored_data = self.item_entity.crud.read(db=self.db, entity_id=item['id'], raise_not_found=False)
        assert stored_data is None
        transaction.revert(storage=self.db)
        stored_data = self.item_entity.crud.read(db=self.db, entity_id=item['id'], raise_not_found=False)
        assert stored_data.id == item['id']

    def test_delete_then_revert_relation(self, relation):
        transaction = DeleteTransaction(entity_type=EntityType.RELATION.name, entity_id=relation['id'], body=relation)
        transaction.apply(storage=self.db)
        stored_data = self.relation_entity.crud.read(db=self.db, entity_id=relation['id'], raise_not_found=False)
        assert stored_data is None
        transaction.revert(storage=self.db)
        stored_data = self.relation_entity.crud.read(db=self.db, entity_id=relation['id'], raise_not_found=False)
        print(f"\n DELETE THEN REVER{jsonable_encoder(stored_data)}")
        assert stored_data.id == relation['id']

    def test_delete_then_revert_resource(self, resource):
        transaction = DeleteTransaction(entity_type=EntityType.RESOURCE.name, entity_id=resource['id'], body=resource)
        transaction.apply(storage=self.db)
        stored_data = self.resource_entity.crud.read(db=self.db, entity_id=resource['id'], raise_not_found=False)
        assert stored_data is None
        transaction.revert(storage=self.db)
        stored_data = self.resource_entity.crud.read(db=self.db, entity_id=resource['id'], raise_not_found=False)
        assert stored_data.id == resource['id']

    def test_create_then_revert_new_item(self, item):
        transaction = DeltaTransaction(body=[
            CreateTransaction(entity_type=EntityType.ITEM.name,
                              entity_id=item['id'], body=item)
        ])
        transaction.apply(storage=self.db)
        stored_data = self.item_entity.crud.read(db=self.db, entity_id=item['id'], raise_not_found=False)
        assert stored_data.id == item['id']
        transaction.revert(storage=self.db)
        stored_data = self.item_entity.crud.read(db=self.db, entity_id=item['id'], raise_not_found=False)
        assert stored_data.active is False

    def __del__(self) -> None:
        self.db.close()



if __name__ == '__main__':
    t = TestTransactions()
    t.test_create_item(t.item1)
    t.test_create_then_revert_new_item(t.item3)
    t.test_update_then_revert_item(t.item1, t.updated_item1)
    t.test_delete_then_revert_item(t.item1)
    t.test_read_item(t.item1)
    t.test_update_item(t.item1, t.updated_item1)
    t.test_create_item(t.item2)

    t.test_create_relation(t.relation1)
    #TODO t.test_update_then_revert_relation(t.relation1, t.updated_relation1) post fix
    t.test_delete_then_revert_relation(t.relation1)
    t.test_read_relation(t.relation1)
    t.test_update_relation(t.relation1, t.updated_relation1)

    t.test_create_resource(t.resource1)
    #TODO t.test_update_then_revert_resource(t.resource1, t.updated_resource1)
    t.test_delete_then_revert_resource(t.resource1)
    t.test_read_resource(t.resource1)
    t.test_update_resource(t.resource1, t.updated_resource1)

    t.test_delete_resource(t.resource1)
    #t.test_create_then_revert_resource(t.resource1)
    t.test_delete_relation(t.relation1)
    #t.test_create_then_revert_relation(t.relation1)
    t.test_delete_item(t.item1)
    t.test_delete_item(t.item2)
    #t.test_create_then_revert_item(t.item1)