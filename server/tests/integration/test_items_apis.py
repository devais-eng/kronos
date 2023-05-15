import pytest

from tests.utils import ApiClient
from tests.integration import Test, api_clients
from tests.integration import testItems, testItemsUpd, testAttributes, testRelations


# Notes:
# _Scope = Literal["session", "package", "module", "class", "function"]
# - with scope "function" we'll have each test executed with the different clients consequently
# - with scope "class" we'll have the whole class executed each time with a different client
@pytest.fixture(scope="class", params=api_clients)
def client(request):
    yield request.param


class TestItems(Test):
    def test_items_get_count_empty(self, client: ApiClient):
        """
        Checks if the count counts 0.
        """
        endpoint = "items/count"
        r = client.read(endpoint=endpoint)
        assert r.is_ok()
        assert r.get_type() == dict

        r_dict = r.get_content()
        assert r_dict["count"] == 0

    @pytest.mark.parametrize("item", [testItems["item1"]])
    def test_item_read_nonexistent(self, client: ApiClient, item):
        """
        Tests the READ interface to read an item which does not exist. Assert that a not found error is raised.
        """
        res = self.read_item_res(client=client, item_id=item["id"])
        assert res.is_not_found()

    # After this test, the 3 test items described in the YAML file will be added to the database.
    def test_item_create_bulk(self, client: ApiClient, items=testItems):
        """
        Tests the creation of a set of items (bulk mode).
        """
        items_list = list(items.values())
        assert self.create_items(client=client, items=items_list)

    @pytest.mark.parametrize("item", [testItems["item1"]])
    def test_item_create_existent(self, client: ApiClient, item):
        """
        Tests the creation of an already existent item. Assert that the client raises a 'bad request' error.
        """
        res = self.create_items_res(client=client, items=item)
        assert res.is_error()

    def test_item_create_bulk_assert(self, client: ApiClient, items=testItems):
        """
        Asserts the creation of the set of items (bulk mode).
        """
        items_list = list(items.values())
        fed_ids = [x["id"] for x in items_list]
        for i in fed_ids:
            assert self.item_exists(client=client, item_id=i)

    def test_items_get_count(self, client: ApiClient):
        """
        Checks if the count counts 3.
        """
        endpoint = "items/count"
        r = client.read(endpoint=endpoint)
        assert r.is_ok()
        assert r.get_type() == dict

        r_dict = r.get_content()
        assert r_dict["count"] == 3

    # After this test, the 1st test item described in the YAML file will be updated
    # (it will still have the same ID but its content will be updated)
    def test_item_update(self, client: ApiClient, item=testItems["item1"], patched_item=testItemsUpd["item1_updated"]):
        """
        Tests the UPDATE interface to partially update an item.
        """
        assert self.update_item(client=client, item=item, patched_item=patched_item)

    def test_item_update_assert(self, client: ApiClient, patched_item=testItemsUpd["item1_updated"]):
        """
        Asserts that the UPDATE test has partially updated the item.
        """
        item_id = patched_item["id"]
        r = self.read_item_res(client=client, item_id=item_id)
        assert r.is_ok()

        patched_item_read = r.get_content()
        assert patched_item_read["name"] == patched_item["name"]
        assert patched_item_read["type"] == patched_item["type"]
        assert patched_item_read["customer_id"] == patched_item["customer_id"]

    def test_item_update_non_existent(self, client: ApiClient):
        """
        Tests the UPDATE interface to partially update an item which does not exist.
        """
        item = {
            "id": "non-existing-ID"
        }
        patched_item = dict()
        res = self.update_item_res(client=client, item=item, patched_item=patched_item)
        assert res.is_not_found()

    # # From this test we expect to find the 3 test items described in the YAML document
    # @pytest.mark.parametrize("keyword, expected_no", [(testItems["item2"]["name"], 1), ("foo_item_id", 0)])
    # def test_item_find_by_name(self, client: ApiClient, keyword, expected_no, skip_value=0, limit_value=100):
    #     """
    #     Tests the READ interface to find items by name.
    #     """
    #     endpoint = f'item/findByName/{keyword}'
    #     params = {
    #         "skip": skip_value,
    #         "limit": limit_value
    #     }
    #     r = client.read(
    #         endpoint=endpoint,
    #         params=params
    #     )
    #     assert r.is_ok()
    #
    #     # Response body check
    #     r_list = r.get_content()  # Will be a list
    #     assert len(r_list) == expected_no
    #
    # # From this test we expect to retrieve the 2 test items with unmodified type
    # @pytest.mark.parametrize("keyword, expected_no", [(testItems["item2"]["type"], 2), ("foo_item_id", 0)])
    # def test_item_find_by_type(self, client: ApiClient, keyword, expected_no, skip_value=0, limit_value=100):
    #     """
    #     Tests the READ interface to find items by type. Then, compares the found entities.
    #     """
    #     endpoint = f'item/findByType/{keyword}'
    #     params = {
    #         "skip": skip_value,
    #         "limit": limit_value
    #     }
    #     r = client.read(
    #         endpoint=endpoint,
    #         params=params
    #     )
    #     assert r.is_ok()
    #
    #     # Response body check
    #     r_list = r.get_content()
    #     assert isinstance(r_list, list)
    #     assert len(r_list) == expected_no

    @pytest.mark.parametrize("entity_id", [testItems["item2"]["id"]])
    @pytest.mark.parametrize("expected_mac", [testItems["item2"]["edge_mac"]])
    def test_item_get_mac(self, client: ApiClient, entity_id, expected_mac):
        """
        Tests the READ interface to read the MAC of an item.
        """
        endpoint = f'item/{entity_id}/mac'

        r = client.read(
            endpoint=endpoint
        )
        assert r.is_ok()
        assert r.get_type() == str

        r_str = r.get_content()
        assert r_str == expected_mac

    @pytest.mark.parametrize("entity_id", ["foo_item_id"])
    def test_item_get_mac_from_non_existent_item(self, client: ApiClient, entity_id):
        """
        Tests the READ interface to read the MAC of an item which doesn't exist.
        """
        endpoint = f'item/{entity_id}/mac'

        r = client.read(
            endpoint=endpoint
        )
        assert r.is_not_found()

    @pytest.mark.parametrize("entity_id", [testItems["item2"]["id"]])
    def test_item_get_version(self, client: ApiClient, entity_id):
        """
        Tests the READ interface to read the version of an item.
        """
        endpoint = f'item/{entity_id}/version'

        r = client.read(
            endpoint=endpoint
        )
        assert r.is_ok()
        assert r.get_type() == str

        # assert r.get_type() == dict
        # r_dict = r.get_content()
        # assert "version" in r_dict.keys()

    @pytest.mark.parametrize("entity_id", ["foo_item_id"])
    def test_item_get_version_from_non_existent_item(self, client: ApiClient, entity_id):
        """
        Tests the READ interface to read the version of an item which doesn't exist.
        """
        endpoint = f'item/{entity_id}/version'

        r = client.read(
            endpoint=endpoint
        )
        assert r.is_not_found()

    @pytest.mark.parametrize("entity_id", [testItems["item2"]["id"]])
    @pytest.mark.parametrize("expected_customer_id", [testItems["item2"]["customer_id"]])
    def test_item_get_customer(self, client: ApiClient, entity_id, expected_customer_id):
        """
        Tests the READ interface to read the customer of an item.
        """
        endpoint = f'item/{entity_id}/customer'
        r = client.read(
            endpoint=endpoint
        )
        assert r.is_ok()

        # Response body check: test if the returned string contains the correct customer id
        r_text = r.get_content()
        assert isinstance(r_text, str)
        r_text = r_text.replace("\"", "")  # Will be a string
        assert r_text == expected_customer_id

    @pytest.mark.parametrize("entity_id", ["foo_item_id"])
    def test_item_get_customer_from_non_existent_item(self, client: ApiClient, entity_id):
        """
        Tests the READ interface to read the customer of an item which does not exist.
        """
        endpoint = f'item/{entity_id}/customer'
        r = client.read(
            endpoint=endpoint
        )
        assert r.is_not_found()

    @pytest.mark.parametrize("entity_id", [testItemsUpd["item1_updated"]["id"]])
    @pytest.mark.parametrize("expected_modified_by", [testItemsUpd["item1_updated"]["modified_by"]])
    def test_item_get_modified_by(self, client: ApiClient, entity_id, expected_modified_by):
        """
        Tests the READ interface to read the id of the last entity that has modified an item.
        """
        # Request URL
        endpoint = f'item/{entity_id}/modified_by'
        r = client.read(
            endpoint=endpoint
        )
        assert r.is_ok()

        # Response body check: test if the returned object is a string
        r_text = r.get_content()
        assert isinstance(r_text, str)

    @pytest.mark.parametrize("entity_id", ["foo_item_id"])
    def test_item_get_modified_by_from_non_existent_item(self, client: ApiClient, entity_id):
        """
        Tests the READ interface to read the id of the last entity that has modified an item which does not exist.
        """
        # Request URL
        endpoint = f'item/{entity_id}/modified_by'
        r = client.read(
            endpoint=endpoint
        )
        assert r.is_not_found()

    # After this test, the relation #2 will be created
    def test_item_create_relation(self, client: ApiClient, relation=testRelations["relation2"]):
        """
        Creates a relation between two items.
        """
        assert self.create_relation(client, relation=relation)

    def test_item_create_relation_assert(self, client: ApiClient, rel=testRelations["relation2"]):
        """
        Asserts that the CREATE interface has created the relation.
        """
        assert self.relation_exists(client, rel["parent_id"], rel["child_id"])

    @pytest.mark.parametrize("test_relation", [testRelations["relation2"]])
    def test_item_get_children(self, client: ApiClient, test_relation):
        """
        Tests the READ interface to read the children of an item.
        """
        entity_id = test_relation["parent_id"]
        endpoint = f'item/{entity_id}/children'
        r = client.read(endpoint=endpoint)
        assert r.is_ok()

        # Response body check: test if the returned list contains the correct relations
        rel_list = r.get_content()
        assert isinstance(rel_list, list)
        rel_list_child_ids = [_["id"] for _ in rel_list]
        assert test_relation["child_id"] in rel_list_child_ids

    @pytest.mark.parametrize("parent_id", ["foo_parent"])
    def test_item_get_children_from_non_existent_item(self, client: ApiClient, parent_id):
        """
        Tests the READ interface to read the children of an item which doesn't exist.
        """
        entity_id = parent_id
        endpoint = f'item/{entity_id}/children'
        r = client.read(endpoint=endpoint)
        assert r.is_not_found()

    @pytest.mark.parametrize("parent_id", [testItems["item1"]["id"]])
    def test_item_get_children_unrelated(self, client: ApiClient, parent_id):
        """
        Tests the READ interface to read the children of an item which is not related with any other item.
        """
        entity_id = parent_id
        endpoint = f'item/{entity_id}/children'
        r = client.read(endpoint=endpoint)
        assert r.is_ok()
        assert r.get_content() == []

    @pytest.mark.parametrize("test_relation", [testRelations["relation2"]])
    def test_item_get_parents(self, client: ApiClient, test_relation):
        """
        Tests the READ interfaces to read the parents of an item.
        """
        entity_id = test_relation["child_id"]
        endpoint = f'item/{entity_id}/parents'
        r = client.read(endpoint=endpoint)
        assert r.is_ok()

        # Response body check: test if the returned list contains the correct relations
        rel_list = r.get_content()
        rel_list_parent_ids = [_["id"] for _ in rel_list]
        assert test_relation["parent_id"] in rel_list_parent_ids

    @pytest.mark.parametrize("parent_id", ["foo_parent"])
    def test_item_get_parents_from_non_existent_item(self, client: ApiClient, parent_id):
        """
        Tests the READ interfaces to read the parents of an item which does not exist.
        """
        entity_id = parent_id
        endpoint = f'item/{entity_id}/parents'
        r = client.read(endpoint=endpoint)
        assert r.is_not_found()

    @pytest.mark.parametrize("parent_id", [testItems["item1"]["id"]])
    def test_item_get_parents_unrelated(self, client: ApiClient, parent_id):
        """
        Tests the READ interfaces to read the parents of an item which is not related with any other items.
        """
        entity_id = parent_id
        endpoint = f'item/{entity_id}/parents'
        r = client.read(endpoint=endpoint)
        assert r.is_ok()
        assert r.get_content() == []

    @pytest.mark.parametrize("test_relation", [testRelations["relation2"]])
    def test_item_get_relations(self, client: ApiClient, test_relation):
        """
        Tests the READ interface to read the relations of an item.
        """
        entity_id = test_relation["child_id"]
        endpoint = f'item/{entity_id}/relations'

        r = client.read(endpoint=endpoint)
        assert r.is_ok()

        assert r.get_type() == list
        r_list = r.get_content()
        assert len(r_list) == 1
        assert type(r_list[0]) == dict
        assert r_list[0]["parent_id"] == test_relation["parent_id"]

    @pytest.mark.parametrize("parent_id", ["foo_parent"])
    def test_item_get_relations_from_non_existent_item(self, client: ApiClient, parent_id):
        """
        Tests the READ interface to read the relations of an item which doesn't exist.
        """
        entity_id = parent_id
        endpoint = f'item/{entity_id}/relations'

        r = client.read(endpoint=endpoint)
        assert r.is_not_found()

    @pytest.mark.parametrize("parent_id", [testItems["item1"]["id"]])
    def test_item_get_relations_unrelated(self, client: ApiClient, parent_id):
        """
        Tests the READ interface to read the relations of an item which is unrelated.
        """
        entity_id = parent_id
        endpoint = f'item/{entity_id}/relations'

        r = client.read(endpoint=endpoint)
        assert r.is_ok()
        assert r.get_content() == []

    # After this test, the relation #2 will no more exist
    def test_item_delete_relation(self, client: ApiClient, relation=testRelations["relation2"]):
        """
        Tests the DELETE interface to delete a relation between two items.
        """
        assert self.delete_relation(client=client, relation=relation)

    def test_item_delete_relation_assert(self, client: ApiClient, rel=testRelations["relation2"]):
        """
        Asserts that the DELETE interface has deleted the relation.
        """
        assert not self.relation_exists(client, rel["parent_id"], rel["child_id"])

    @pytest.mark.parametrize("test_item", [testItems["item1"]])
    def test_item_get_attributes(self, client: ApiClient, test_item):
        """
        Tests the READ interface to read the attributes of an item.
        """
        # ID of the item which owns the attribute
        item_id = test_item["id"]
        endpoint = f'item/{item_id}/attributes'
        r = client.read(endpoint=endpoint)
        assert r.is_ok()

        # Response body check: test if the returned list contains the
        # correct attributes
        expected_attr_ids = [_["id"] for _ in test_item["attributes"]]
        retrieved_attr_list = r.get_content()
        assert isinstance(retrieved_attr_list, list)
        assert len(test_item["attributes"]) == len(retrieved_attr_list)
        for attr in retrieved_attr_list:
            assert attr["id"] in expected_attr_ids

    @pytest.mark.parametrize("test_item", [testItems["item3"]])
    def test_item_get_attributes_from_empty_item(self, client: ApiClient, test_item):
        """
        Tests the READ interface to read the attributes of an item.
        """
        # ID of the item which owns the attribute
        item_id = test_item["id"]
        endpoint = f'item/{item_id}/attributes'
        r = client.read(endpoint=endpoint)
        assert r.is_ok()

        assert r.get_type() == list
        r_list = r.get_content()
        assert len(r_list) == 0

    @pytest.mark.parametrize("test_item", [{"id": "foo_item_id"}])
    def test_item_get_attributes_from_non_existent_item(self, client: ApiClient, test_item):
        """
        Tests the READ interface to read the attributes of an item.
        """
        # ID of the item which owns the attribute
        item_id = test_item["id"]
        endpoint = f'item/{item_id}/attributes'
        r = client.read(endpoint=endpoint)
        assert r.is_not_found()

    @pytest.mark.parametrize("test_attribute, test_item",
                             [(testAttributes["attribute2"], testItems["item1"])])
    def test_item_get_attribute_by_name(self, client: ApiClient, test_attribute, test_item):
        """
        Tests the READ interface to read a particular attribute by item ID and attribute name.
        """
        # ID of the item which owns the attribute
        item_id = test_item["id"]
        # Name of the attribute to look for
        attribute_name = test_attribute["name"]

        endpoint = f'item/{item_id}/attribute/name/{attribute_name}'
        r = client.read(endpoint=endpoint)
        assert r.is_ok()

        # Response body check: check the returned attribute
        retrieved_attr = r.get_content()  # Will be a dict
        assert isinstance(retrieved_attr, dict)
        assert retrieved_attr["id"] == test_attribute["id"]

    @pytest.mark.parametrize("test_attribute, test_item",
                             [({"name": "foo"}, testItems["item1"])])
    def test_item_get_attribute_by_name_non_existent_name(self, client: ApiClient, test_attribute, test_item):
        """
        Tests the READ interface to read a non existent attribute by item ID and attribute name.
        """
        # ID of the item which owns the attribute
        item_id = test_item["id"]
        # Name of the attribute to look for
        attribute_name = test_attribute["name"]

        endpoint = f'item/{item_id}/attribute/name/{attribute_name}'
        r = client.read(endpoint=endpoint)
        assert r.is_not_found()

    @pytest.mark.parametrize("test_attribute, test_item",
                             [(testAttributes["attribute2"], {"id": "foo_item_id"})])
    def test_item_get_attribute_by_name_non_existent_item(self, client: ApiClient, test_attribute, test_item):
        """
        Tests the READ interface to read a non existent attribute by item ID and attribute name.
        """
        # ID of the item which owns the attribute
        item_id = test_item["id"]
        # Name of the attribute to look for
        attribute_name = test_attribute["name"]

        endpoint = f'item/{item_id}/attribute/name/{attribute_name}'
        r = client.read(endpoint=endpoint)
        assert r.is_not_found()

    @pytest.mark.parametrize("test_attribute, test_item",
                             [(testAttributes["attribute1"], testItems["item1"])])
    def test_item_get_attribute_id_by_name(self, client: ApiClient, test_attribute, test_item):
        """
        Tests the READ interface to read the ID of a particular attribute attached to an item.
        """
        # ID of the item which owns the attribute
        item_id = test_item["id"]
        # Name of the attribute to look for
        attribute_name = test_attribute["name"]
        endpoint = f'item/{item_id}/attribute/name/{attribute_name}/id'
        r = client.read(endpoint=endpoint)
        assert r.is_ok()

        # Response body check: check the returned attribute
        retrieved_attr_id = r.get_content()
        assert retrieved_attr_id == test_attribute["id"]

    @pytest.mark.parametrize("test_attribute, test_item",
                             [({"name": "foo"}, testItems["item1"])])
    def test_item_get_attribute_id_by_name_non_existent_name(self, client: ApiClient, test_attribute, test_item):
        """
        Tests the READ interface to read the ID of a particular attribute attached to an item.
        """
        # ID of the item which owns the attribute
        item_id = test_item["id"]
        # Name of the attribute to look for
        attribute_name = test_attribute["name"]
        endpoint = f'item/{item_id}/attribute/name/{attribute_name}/id'
        r = client.read(endpoint=endpoint)
        assert r.is_not_found()

    @pytest.mark.parametrize("test_attribute, test_item",
                             [(testAttributes["attribute1"], {"id": "foo_item_id"})])
    def test_item_get_attribute_id_by_name_non_existent_item(self, client: ApiClient, test_attribute, test_item):
        """
        Tests the READ interface to read the ID of a particular attribute attached to an item.
        """
        # ID of the item which owns the attribute
        item_id = test_item["id"]
        # Name of the attribute to look for
        attribute_name = test_attribute["name"]
        endpoint = f'item/{item_id}/attribute/name/{attribute_name}/id'
        r = client.read(endpoint=endpoint)
        assert r.is_not_found()

    @pytest.mark.parametrize("test_attribute, test_item",
                             [(testAttributes["attribute1"], testItems["item1"])])
    def test_item_get_attribute_value_by_name(self, client: ApiClient, test_attribute, test_item):
        """
        Tests the READ interface to read the value of a particular attribute attached to an item.
        """
        # ID of the item which owns the attribute
        item_id = test_item["id"]
        # Name of the attribute to look for
        attribute_name = test_attribute["name"]
        endpoint = f'item/{item_id}/attribute/name/{attribute_name}/value'

        r = client.read(endpoint=endpoint, )
        assert r.is_ok()

        # Response body check: check the returned attribute
        assert r.get_type() == dict
        r_dict = r.get_content()
        retrieved_attr_value = r_dict["value"]
        assert retrieved_attr_value == test_attribute["value"]

        # retrieved_attr_value = r.get_content()
        # assert isinstance(retrieved_attr_value, str)
        # retrieved_attr_value = retrieved_attr_value.replace("\"", "")
        # assert retrieved_attr_value == test_attribute["value"]

    @pytest.mark.parametrize("test_attribute, test_item",
                             [({"name": "foo_attr_name"}, testItems["item1"])])
    def test_item_get_attribute_value_by_name_non_existent_name(self, client: ApiClient, test_attribute, test_item):
        """
        Tests the READ interface to read the value of a particular attribute attached to an item.
        """
        # ID of the item which owns the attribute
        item_id = test_item["id"]
        # Name of the attribute to look for
        attribute_name = test_attribute["name"]
        endpoint = f'item/{item_id}/attribute/name/{attribute_name}/value'

        r = client.read(endpoint=endpoint, )
        assert r.is_not_found()

    @pytest.mark.parametrize("test_attribute, test_item",
                             [(testAttributes["attribute1"], {"id": "foo_item_id"})])
    def test_item_get_attribute_value_by_name_non_existent_item(self, client: ApiClient, test_attribute, test_item):
        """
        Tests the READ interface to read the value of a particular attribute attached to an item.
        """
        # ID of the item which owns the attribute
        item_id = test_item["id"]
        # Name of the attribute to look for
        attribute_name = test_attribute["name"]
        endpoint = f'item/{item_id}/attribute/name/{attribute_name}/value'

        r = client.read(endpoint=endpoint, )
        assert r.is_not_found()

    @pytest.mark.parametrize("attribute_type, test_item",
                             [(testAttributes["attribute1"]["type"], testItems["item1"])])
    def test_item_get_attributes_by_type(self, client: ApiClient, attribute_type, test_item):
        """
        Tests the READ interface to read the attributes with a certain type attached to an item.
        """
        # ID of the item which owns the attribute
        item_id = test_item["id"]
        endpoint = f'item/{item_id}/attributes/type/{attribute_type}'

        r = client.read(endpoint=endpoint)
        assert r.is_ok()

        # Response body check: check the returned attribute
        expected_attr_list = test_item["attributes"]
        expected_attr_names = [_["name"] for _ in expected_attr_list]

        retrieved_attr_list = r.get_content()
        assert isinstance(retrieved_attr_list, list)
        retrieved_attr_names = [_["name"] for _ in retrieved_attr_list]
        for retrieved_attr_name in retrieved_attr_names:
            assert retrieved_attr_name in expected_attr_names

    @pytest.mark.parametrize("attribute_type, test_item",
                             [(testAttributes["attribute1"]["type"], {"id": "foo_item_id"})])
    def test_item_get_attributes_by_type_non_existent_item(self, client: ApiClient, attribute_type, test_item):
        """
        Tests the READ interface to read the attributes with a certain type attached to a non existent item.
        """
        # ID of the item which owns the attribute
        item_id = test_item["id"]
        endpoint = f'item/{item_id}/attributes/type/{attribute_type}'

        r = client.read(endpoint=endpoint)
        assert r.is_ok()

        assert r.get_type() == list
        assert r.get_content() == []

    @pytest.mark.parametrize("attribute_type, test_item",
                             [(testAttributes["attribute1"]["type"], {"id": "foo_item_id"})])
    def test_item_get_attributes_by_type_from_empty_item(self, client: ApiClient, attribute_type, test_item):
        """
        Tests the READ interface to read the attributes with a certain type attached to an empty item.
        """
        # ID of the item which owns the attribute
        item_id = test_item["id"]
        endpoint = f'item/{item_id}/attributes/type/{attribute_type}'

        r = client.read(endpoint=endpoint)
        assert r.is_ok()

        assert r.get_type() == list
        assert r.get_content() == []

    # After this test, the 2nd test item will be deleted.
    @pytest.mark.parametrize("item_id", [testItems["item2"]["id"]], scope="function")
    def test_item_delete(self, client: ApiClient, item_id):
        """
        Tests the DELETE interface to delete an item.
        """
        self.delete_item(client=client, item_id=item_id)

    def test_item_delete_cascade_assert(self, client: ApiClient, entity_id=testItems["item2"]["id"],
                                        entity_attributes=testItems["item2"]["attributes"]):
        """
        Asserts that the cascade deletion of attributes works correctly.
        """
        # Check that the item has actually been deleted
        assert not self.item_exists(client, entity_id)

        # Check that the item's attributes have actually been deleted
        entity_attributes_ids = [_["id"] for _ in entity_attributes]
        for attr_id in entity_attributes_ids:
            assert not self.attribute_exists(client, attr_id)

    @pytest.mark.parametrize("item_id", [testItems["item2"]["id"]], scope="function")
    def test_item_delete_non_existent(self, client: ApiClient, item_id):
        """
        Tests the DELETE interface to delete an item which does not exist.
        """
        res = self.delete_item_res(client=client, item_id=item_id)
        assert res.is_not_found()

    # Deletion of the test items will lead to the same database content as the beginning of the tests.
    @pytest.mark.parametrize("item_id", [testItems["item1"]["id"],
                                         testItems["item3"]["id"]], scope="function")
    def test_item_clear_remaining_items(self, client: ApiClient, item_id):
        """
        Deletes an item.
        """
        self.delete_item(client=client, item_id=item_id)

    @pytest.mark.parametrize("entity_id", [testItems["item1"]["id"],
                                           testItems["item3"]["id"]], scope="function")
    def test_item_clear_remaining_items_assert(self, client: ApiClient, entity_id):
        """
        Asserts that deletion of items works correctly.
        """
        # Check that the item has actually been deleted
        assert not self.item_exists(client, entity_id)

    # No more attributes should exist after the deletion of items.
    @pytest.mark.parametrize("attribute_id", [testAttributes["attribute1"]["id"], testAttributes["attribute2"]["id"],
                                              testAttributes["attribute3"]["id"], testAttributes["attribute4"]["id"],
                                              testAttributes["attribute5"]["id"], testAttributes["attribute6"]["id"]],
                             scope="function")
    def test_attributes_deletion_assert(self, client: ApiClient, attribute_id):
        """
        Checks that no more attributes exist after the deletion of items.
        """
        assert not self.attribute_exists(client=client, attribute_id=attribute_id)

    def test_items_get_count_after_deletion(self, client: ApiClient):
        """
        Checks if the count counts 0.
        """
        endpoint = "items/count"
        r = client.read(endpoint=endpoint)
        assert r.is_ok()
        assert r.get_type() == dict

        r_dict = r.get_content()
        assert r_dict["count"] == 0
