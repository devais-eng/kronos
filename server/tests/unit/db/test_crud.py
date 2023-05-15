from kronos.storage.entity.item import ItemEntity
from kronos.storage.entity.resource import ResourceEntity
from kronos.storage.entity.relation import RelationEntity
from kronos.storage import init_db
from kronos.storage import SessionFactory
from fastapi.encoders import jsonable_encoder
import yaml


class TestItem:
    db = SessionFactory()
    init_db(db)

    item_entity = ItemEntity()
    resource_entity = ResourceEntity()
    relation_entity = RelationEntity()

    with open('params_all_apis.yaml') as file:
        var = yaml.load(file, Loader=yaml.FullLoader)
    item1 = ItemEntity.view.create()(**{k: v for (k, v) in var['variables']['item1'].items()})
    item2 = ItemEntity.view.create()(**{k: v for (k, v) in var['variables']['item2'].items()})
    item3 = ItemEntity.view.create()(**{k: v for (k, v) in var['variables']['item3'].items()})
    updated_item1 = ItemEntity.view.create()(**{k: v for (k, v) in var['variables']['updated_item1'].items()})
    patched_item1 = ItemEntity.view.create()(**{k: v for (k, v) in var['variables']['patched_item1'].items()})

    relation1 = RelationEntity.view.create()(**{k: v for (k, v) in var['variables']['relation1'].items()})
    relation2 = RelationEntity.view.create()(**{k: v for (k, v) in var['variables']['relation2'].items()})
    updated_relation1 = RelationEntity.view.create()(
        **{k: v for (k, v) in var['variables']['updated_relation1'].items()})

    resource1 = ResourceEntity.view.create()(**{k: v for (k, v) in var['variables']['resource1'].items()})
    resource2 = ResourceEntity.view.create()(**{k: v for (k, v) in var['variables']['resource2'].items()})
    resource3 = ResourceEntity.view.create()(**{k: v for (k, v) in var['variables']['resource3'].items()})
    updated_resource1 = ResourceEntity.view.create()(
        **{k: v for (k, v) in var['variables']['updated_resource1'].items()})
    patched_resource1 = ResourceEntity.view.create()(
        **{k: v for (k, v) in var['variables']['patched_resource1'].items()})

    def test_post_item(self, item):
        entity = self.item_entity.crud.create(db=self.db, to_create=item)
        jsonable_encoder(entity)
        assert entity.name == item.name and \
               entity.type == item.type and \
               entity.customer_id == item.customer_id and \
               entity.sync_policy == item.sync_policy and \
               entity.edge_mac == item.edge_mac and \
               entity.id == item.id and \
               entity.created_by == item.created_by and \
               isinstance(entity.created_at, int) and \
               entity.modified_by == None and \
               entity.modified_at == None and \
               isinstance(entity.version, str) and \
               entity.active == True and \
               entity.resources == []

    def test_post_multiple_items(self, item_list):
        entities = [self.item_entity.crud.create(db=self.db, to_create=item) for item in item_list]
        for entity, item in zip(entities, item_list):
            jsonable_encoder(entity)
            assert entity.name == item.name and \
                   entity.type == item.type and \
                   entity.customer_id == item.customer_id and \
                   entity.sync_policy == item.sync_policy and \
                   entity.edge_mac == item.edge_mac and \
                   entity.id == item.id and \
                   entity.created_by == item.created_by and \
                   isinstance(entity.created_at, int) and \
                   entity.modified_by == None and \
                   entity.modified_at == None and \
                   isinstance(entity.version, str) and \
                   entity.resources == []

    def test_post_resource(self, resource):
        entity = self.resource_entity.crud.create(db=self.db, to_create=resource)
        jsonable_encoder(entity)
        assert entity.name == resource.name and \
               entity.id == resource.id and \
               entity.instance_id == resource.instance_id and \
               entity.item_id == resource.item_id and \
               entity.type == resource.type and \
               entity.source_timestamp == resource.source_timestamp and \
               entity.created_by == resource.created_by

    def test_post_multiple_resources(self, resource_list):
        entities = [self.resource_entity.crud.create(db=self.db, to_create=resource) for resource in resource_list]
        for entity, resource in zip(entities, resource_list):
            jsonable_encoder(entity)
            assert entity.instance_id == resource.instance_id and \
                   entity.type == resource.type and \
                   entity.id == resource.id and \
                   entity.source_timestamp == resource.source_timestamp and \
                   entity.created_by == resource.created_by and \
                   entity.id == resource.id and \
                   entity.parameters == resource.parameters and \
                   entity.item_id == resource.item_id

    def test_post_multiple_relations(self, all_relations):
        entities = [self.relation_entity.crud.create(db=self.db, to_create=relation) for relation in all_relations]
        for entity, relation in zip(entities, all_relations):
            jsonable_encoder(entity)
            assert entity.parent_id == relation.parent_id and \
                   entity.child_id == relation.child_id and \
                   entity.id == relation.id and \
                   entity.source_timestamp == relation.source_timestamp and \
                   entity.created_by == relation.created_by

    def test_get_item_by_id(self, item):
        previous_data = self.item_entity.crud.read(db=self.db, entity_id=item.id, raise_not_found=False)
        jsonable_encoder(previous_data)
        assert previous_data.name == item.name and \
               previous_data.type == item.type and \
               previous_data.customer_id == item.customer_id and \
               previous_data.sync_policy == item.sync_policy and \
               previous_data.edge_mac == item.edge_mac and \
               previous_data.id == item.id and \
               previous_data.created_by == item.created_by and \
               isinstance(previous_data.created_at, int) and \
               previous_data.modified_by is None and \
               previous_data.modified_at is None and \
               isinstance(previous_data.version, str) and \
               previous_data.active == True and \
               previous_data.resources == []

    def test_get_all_items(self, all_items):
        entities = [self.item_entity.crud.read(db=self.db, entity_id=item.id, raise_not_found=False) for item in
                    all_items]
        for entity, item in zip(entities, all_items):
            jsonable_encoder(entity)
            assert entity.name == item.name and \
                   entity.type == item.type and \
                   entity.customer_id == item.customer_id and \
                   entity.sync_policy == item.sync_policy and \
                   entity.edge_mac == item.edge_mac and \
                   entity.id == item.id and \
                   entity.created_by == item.created_by and \
                   isinstance(entity.created_at, int) and \
                   entity.modified_by is None and \
                   entity.modified_at is None and \
                   isinstance(entity.version, str) and \
                   entity.active is True and \
                   entity.resources == item.resources

    def test_get_relation_by_id(self, relation):
        entity = self.relation_entity.crud.read(db=self.db, entity_id=relation.id, raise_not_found=False)
        jsonable_encoder(entity)
        assert entity.parent_id == relation.parent_id and \
               entity.child_id == relation.child_id and \
               entity.id == relation.id and \
               entity.source_timestamp == relation.source_timestamp and \
               entity.created_by == relation.created_by

    def test_get_all_relations(self, all_relations):
        entities = [self.relation_entity.crud.read(db=self.db, entity_id=relation.id, raise_not_found=False) for
                    relation in all_relations]
        for entity, relation in zip(entities, all_relations):
            jsonable_encoder(entity)
            assert entity.parent_id == relation.parent_id and \
                   entity.child_id == relation.child_id and \
                   entity.id == relation.id and \
                   entity.source_timestamp == relation.source_timestamp and \
                   entity.created_by == relation.created_by

    def test_get_item_relation(self, item, item_relations):
        previous_data = self.item_entity.relatives(db=self.db, item_id=item.id)
        jsonable_encoder(previous_data)
        for relation, item_relation in zip(previous_data, item_relations):
            assert relation.parent_id == item_relation.parent_id and \
                   relation.child_id == item_relation.child_id and \
                   relation.id == item_relation.id and \
                   relation.source_timestamp == item_relation.source_timestamp and \
                   relation.created_by == item_relation.created_by

    def test_get_item_children(self, item, item_children):
        previous_data = self.item_entity.list_children(db=self.db, item_id=item.id)
        jsonable_encoder(previous_data)
        for child, item_child in zip(previous_data, item_children):
            assert child.id == item_child.id

    def test_get_item_parents(self, item, item_parents):
        previous_data = self.item_entity.list_parents(db=self.db, item_id=item.id)
        jsonable_encoder(previous_data)
        for parent, item_parent in zip(previous_data, item_parents):
            assert parent.id == item_parent.id

    def test_get_item_resources(self, item, item_resources):
        previous_data = self.item_entity.list_resources(db=self.db, item_id=item.id)
        jsonable_encoder(previous_data)
        for resource, item_resource in zip(previous_data, item_resources):
            assert resource.id == item_resource.id

    def test_get_resource_by_id(self, resource):
        entity = self.resource_entity.crud.read(db=self.db, entity_id=resource.id, raise_not_found=False)
        jsonable_encoder(entity)
        assert entity.instance_id == resource.instance_id and \
               entity.type == resource.type and \
               entity.id == resource.id and \
               entity.source_timestamp == resource.source_timestamp and \
               entity.created_by == resource.created_by and \
               entity.id == resource.id and \
               entity.item_id == resource.item_id

    def test_get_all_resources(self, all_resources):
        entities = [self.resource_entity.crud.read(db=self.db, entity_id=resource.id, raise_not_found=False) for
                    resource in all_resources]
        for entity, resource in zip(entities, all_resources):
            jsonable_encoder(entity)
            assert entity.instance_id == resource.instance_id and \
                   entity.type == resource.type and \
                   entity.id == resource.id and \
                   entity.source_timestamp == resource.source_timestamp and \
                   entity.created_by == resource.created_by and \
                   entity.id == resource.id and \
                   entity.item_id == resource.item_id

    def test_patch_item(self, item, patched_item):
        previous_data = self.item_entity.crud.read(db=self.db, entity_id=item.id, raise_not_found=False)
        self.item_entity.crud.update(db=self.db, previous=previous_data, new=jsonable_encoder(patched_item))
        updated_data = self.item_entity.crud.read(db=self.db, entity_id=item.id, raise_not_found=False)
        jsonable_encoder(updated_data)
        assert updated_data.id == patched_item.id and \
               updated_data.name == patched_item.name and \
               updated_data.type == patched_item.type and \
               updated_data.customer_id == patched_item.customer_id and \
               updated_data.sync_policy == patched_item.sync_policy and \
               updated_data.source_timestamp == patched_item.source_timestamp and \
               updated_data.created_by == patched_item.created_by and \
               updated_data.edge_mac == patched_item.edge_mac and \
               isinstance(updated_data.version, str) and \
               updated_data.resources == []

    def test_patch_relation(self, relation, patched_relation):
        previous_data = self.relation_entity.crud.read(db=self.db, entity_id=relation.id, raise_not_found=False)
        self.relation_entity.crud.update(db=self.db, previous=previous_data, new=jsonable_encoder(patched_relation))
        updated_data = self.relation_entity.crud.read(db=self.db, entity_id=relation.id, raise_not_found=False)
        jsonable_encoder(updated_data)
        assert updated_data.parent_id == patched_relation.parent_id and \
               updated_data.child_id == patched_relation.child_id and \
               updated_data.id == patched_relation.id and \
               updated_data.source_timestamp == patched_relation.source_timestamp and \
               updated_data.created_by == patched_relation.created_by

    def test_put_item(self, item, updated_item):
        previous_data = self.item_entity.crud.read(db=self.db, entity_id=item.id, raise_not_found=False)
        self.item_entity.crud.update(db=self.db, previous=previous_data, new=jsonable_encoder(updated_item))
        updated_data = self.item_entity.crud.read(db=self.db, entity_id=item.id, raise_not_found=False)
        jsonable_encoder(updated_data)
        assert updated_data.id == updated_item.id and \
               updated_data.name == updated_item.name and \
               updated_data.type == updated_item.type and \
               updated_data.customer_id == updated_item.customer_id and \
               updated_data.sync_policy == updated_item.sync_policy and \
               updated_data.source_timestamp == updated_item.source_timestamp and \
               updated_data.created_by == updated_item.created_by and \
               updated_data.edge_mac == updated_item.edge_mac and \
               isinstance(updated_data.version, str) and \
               updated_data.resources == []

    def test_put_resource(self, resource, updated_resource):
        previous_data = self.resource_entity.crud.read(db=self.db, entity_id=resource.id, raise_not_found=False)
        self.resource_entity.crud.update(db=self.db, previous=previous_data, new=jsonable_encoder(updated_resource))
        updated_data = self.resource_entity.crud.read(db=self.db, entity_id=resource.id, raise_not_found=False)
        jsonable_encoder(updated_data)
        assert updated_data.name == updated_resource.name and \
               updated_data.instance_id == updated_resource.instance_id and \
               updated_data.type == updated_resource.type and \
               updated_data.source_timestamp == updated_resource.source_timestamp and \
               updated_data.created_by == updated_resource.created_by and \
               updated_data.item_id == updated_resource.item_id

    def test_patch_resource(self, resource, patched_resource):
        previous_data = self.resource_entity.crud.read(db=self.db, entity_id=resource.id, raise_not_found=False)
        self.resource_entity.crud.update(db=self.db, previous=previous_data, new=jsonable_encoder(patched_resource))
        updated_data = self.resource_entity.crud.read(db=self.db, entity_id=resource.id, raise_not_found=False)
        assert updated_data.instance_id == patched_resource.instance_id and \
               updated_data.type == patched_resource.type and \
               updated_data.name == patched_resource.name and \
               updated_data.item_id == patched_resource.item_id and \
               updated_data.id == patched_resource.id and \
               updated_data.created_by == patched_resource.created_by

    def test_delete_item(self, item):
        self.item_entity.crud.remove(db=self.db, entity_id=item.id)
        deleted_data = self.item_entity.crud.read(db=self.db, entity_id=item.id, raise_not_found=False)
        assert deleted_data == None

    def test_delete_relation(self, relation):
        self.relation_entity.crud.remove(db=self.db, entity_id=relation.id)
        deleted_data = self.relation_entity.crud.read(db=self.db, entity_id=relation.id, raise_not_found=False)
        assert deleted_data == None

    def test_delete_resource(self, resource):
        self.relation_entity.crud.remove(db=self.db, entity_id=resource.id)
        deleted_data = self.resource_entity.crud.read(db=self.db, entity_id=resource.id, raise_not_found=False)
        assert deleted_data == None

    def test_is_resource_deleted(self, resource):
        deleted_data = self.resource_entity.crud.read(db=self.db, entity_id=resource.id, raise_not_found=False)
        assert deleted_data == None

    def test_is_relation_deleted(self, relation):
        deleted_data = self.relation_entity.crud.read(db=self.db, entity_id=relation.id, raise_not_found=False)
        assert deleted_data == None

    def __del__(self):
        self.db.close()


# TODO: use pytest
if __name__ == '__main__':
    t = TestItem()
    t.test_post_item(t.item1)
    t.test_post_multiple_items([t.item2, t.item3])
    t.test_get_item_by_id(t.item1)
    t.test_get_all_items([t.item1, t.item2, t.item3])
    t.test_put_item(t.item1, t.updated_item1)
    t.test_patch_item(t.item1, t.patched_item1)
    t.test_post_multiple_relations([t.relation1, t.relation2])
    t.test_get_all_relations([t.relation1, t.relation2])
    t.test_get_item_relation(t.item2, [t.relation1, t.relation2])
    t.test_get_item_children(t.item2, [t.item3])
    t.test_get_item_parents(t.item2, [t.item1])
    t.test_post_resource(t.resource1)
    t.test_get_item_resources(t.item1, [t.resource1])
    t.test_get_relation_by_id(t.relation1)
    t.test_patch_relation(t.relation1, t.updated_relation1)
    t.test_post_multiple_resources([t.resource2, t.resource3])
    t.test_get_all_resources([t.resource1, t.resource2, t.resource3])
    t.test_put_resource(t.resource1, t.updated_resource1)
    t.test_get_resource_by_id(t.resource2)
    t.test_patch_resource(t.resource1, t.patched_resource1)
    t.test_delete_relation(t.relation1)
    # t.test_delete_resource(t.resource1) #TODO: Test after patch fixed
    t.test_delete_item(t.item1)
    t.test_delete_item(t.item2)
    t.test_delete_item(t.item3)
    t.test_is_resource_deleted(t.resource2)
    t.test_is_resource_deleted(t.resource3)
    t.test_is_relation_deleted(t.relation2)
