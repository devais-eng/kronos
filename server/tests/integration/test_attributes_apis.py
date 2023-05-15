import pytest

from tests.utils import ApiClient
from tests.integration import Test, api_clients, testAttributesUpd
from tests.integration import testItems, testAttributes


# Notes:
# _Scope = Literal["session", "package", "module", "class", "function"]
# - with scope "function" we'll have each test executed with the different clients consequently
# - with scope "class" we'll have the whole class executed each time with a different client
# Use api_clients.get_protocol_iterator() for issuing api tests on single clients
@pytest.fixture(scope="class", params=api_clients)
def client(request):
    yield request.param


class TestAttributes(Test):
    def test_attributes_read_all_empty(self, client: ApiClient):
        """
        Checks that the attributes list is empty.
        """
        r = self.read_attributes_res(client=client)
        attr_list = r.get_content()
        assert isinstance(attr_list, list)
        assert len(attr_list) == 0

    def test_attributes_get_count_empty(self, client: ApiClient):
        """
        Checks if the count counts 0.
        """
        endpoint = "attributes/count"
        r = client.read(endpoint=endpoint)
        assert r.is_ok()
        assert r.get_type() == dict

        r_dict = r.get_content()
        assert r_dict["count"] == 0

    # After this test, the 3 test items described in the YAML file will be added to the database.
    # As a consequence, attributes 1,2 (-> item 1) and 3,4 (-> item 2) will be created.
    def test_item_create_bulk(self, client: ApiClient, items=testItems):
        """
        Tests the creation of a set of items (bulk mode).
        """
        items_list = list(items.values())
        assert self.create_items(client=client, items=items_list)

    def test_item_create_bulk_assert(self, client: ApiClient, items=testItems):
        """
        Asserts the creation of the set of items (bulk mode).
        """
        items_list = list(items.values())
        fed_ids = [x["id"] for x in items_list]
        for i in fed_ids:
            assert self.item_exists(client=client, item_id=i)

    @pytest.mark.parametrize("attribute",
                             [testAttributes["attribute1"], testAttributes["attribute2"],
                              testAttributes["attribute3"], testAttributes["attribute4"]],
                             scope="function")
    def test_assert_attribute_creation_from_item(self, client: ApiClient, attribute):
        """
        Asserts the creation of the attributes attached to the set of items created in bulk mode.
        """
        assert self.attribute_exists(client=client, attribute_id=attribute["id"])

    # After this test, the test attributes #5,6 described in the YAML file will be added to the database
    @pytest.mark.parametrize("attribute",
                             [testAttributes["attribute5"], testAttributes["attribute6"]],
                             scope="function")
    def test_attribute_create(self, client: ApiClient, attribute):
        """
        Tests the creation of an attribute.
        """
        assert self.create_attributes(client=client, attribute=attribute)

    @pytest.mark.parametrize("attribute",
                             [testAttributes["attribute5"], testAttributes["attribute6"]],
                             scope="function")
    def test_attribute_create_assert(self, client: ApiClient, attribute):
        """
        Checks the creation of an attribute.
        """
        assert self.attribute_exists(client=client, attribute_id=attribute["id"])

    @pytest.mark.parametrize("attribute", [testAttributes["attribute5"]], scope="function")
    def test_attribute_create_existent(self, client: ApiClient, attribute):
        """
        Tests the creation of an attribute which already exists.
        """
        r = self.create_attributes_res(client=client, attributes=attribute)
        assert r.is_bad_request()

    # After this test, the test attributes #5,6 described in the YAML file will be deleted from the database
    @pytest.mark.parametrize("attribute",
                             [testAttributes["attribute5"], testAttributes["attribute6"]],
                             scope="function")
    def test_attribute_delete(self, client: ApiClient, attribute):
        """
        Tests the deletion of an attribute.
        """
        attribute_id = attribute["id"]
        assert self.delete_attribute(client=client, attribute_id=attribute_id)

    @pytest.mark.parametrize("attribute",
                             [testAttributes["attribute5"], testAttributes["attribute6"]],
                             scope="function")
    def test_attribute_delete_assert(self, client: ApiClient, attribute):
        """
        Verify the deletion of an attribute.
        """
        attribute_id = attribute["id"]
        assert not self.attribute_exists(client=client, attribute_id=attribute_id)

    @pytest.mark.parametrize("attribute",
                             [testAttributes["attribute5"], testAttributes["attribute6"]],
                             scope="function")
    def test_attribute_delete_nonexistent(self, client: ApiClient, attribute):
        """
        Tests the deletion of an attribute which does not exist.
        Checks that the client returns a 'not found' error.
        """
        attribute_id = attribute["id"]
        r = self.delete_attribute_res(client=client, attribute_id=attribute_id)
        assert r.is_not_found()

    # TODO: add a test for checking page and page_size parameters in "read all" operation

    def test_attributes_read_all(self, client: ApiClient):
        """
        Checks if the expected 4 attributes can be retrieved.
        """
        r = self.read_attributes_res(client=client)
        attr_list = r.get_content()
        assert isinstance(attr_list, list)
        assert len(attr_list) == 4

    def test_attributes_get_count(self, client: ApiClient):
        """
        Checks if the count counts the expected 4 attributes.
        """
        endpoint = "attributes/count"
        r = client.read(endpoint=endpoint)
        assert r.is_ok()
        assert r.get_type() == dict

        r_dict = r.get_content()
        assert r_dict["count"] == 4

    # @pytest.mark.parametrize("attribute_name, expected_attributes", [
    #     (testAttributes["attribute3"]["name"], [testAttributes["attribute1"], testAttributes["attribute3"]]),
    #     ("non_existent_attribute", [])],
    #                          scope="function")
    # def test_attributes_find_by_name(self, client: ApiClient, attribute_name, expected_attributes):
    #     """
    #     Check if it's possible to correctly retrieve the attributes by name through the READ interface.
    #     """
    #     endpoint = f"attribute/findByName/{attribute_name}"
    #     params = {
    #         "skip": 0,
    #         "limit": 100
    #     }
    #     r = client.read(endpoint=endpoint, params=params)
    #     assert r.is_ok()
    #
    #     attr_list = r.get_content()
    #     for attr in attr_list:
    #         assert attr["name"] == attribute_name
    #     assert len(attr_list) == len(expected_attributes)
    #
    #     read_ids = [_["id"] for _ in attr_list]
    #     expected_ids = [_["id"] for _ in expected_attributes]
    #     for expected_id in expected_ids:
    #         assert expected_id in read_ids
    #
    # @pytest.mark.parametrize("attribute_type, expected_attributes",
    #                          [(testAttributes["attribute1"]["type"], [testAttributes["attribute1"],
    #                                                                   testAttributes["attribute3"]]),
    #                           ("non_existent_attribute", [])],
    #                          scope="function")
    # def test_attributes_find_by_type(self, client: ApiClient, attribute_type, expected_attributes):
    #     """
    #     Check if it's possible to correctly retrieve the attributes by type through the READ interface.
    #     """
    #     endpoint = f"attribute/findByType/{attribute_type}"
    #     params = {
    #         "skip": 0,
    #         "limit": 100
    #     }
    #     r = client.read(endpoint=endpoint, params=params)
    #     assert r.is_ok()
    #
    #     attr_list = r.get_content()
    #     assert len(attr_list) == len(expected_attributes)
    #
    #     received_attributes_ids = [attr["id"] for attr in attr_list]
    #     expected_attributes_ids = [attr["id"] for attr in expected_attributes]
    #     for received_attr_id in received_attributes_ids:
    #         assert received_attr_id in expected_attributes_ids
    #
    # @pytest.mark.parametrize("attribute, expected_item",
    #                          [(testAttributes["attribute1"], testItems["item1"]),
    #                           (testAttributes["attribute2"], testItems["item1"])],
    #                          scope="function")
    # def test_attribute_get_owner_item(self, client: ApiClient, attribute, expected_item):
    #     """
    #     Tests if it's possible to correctly retrieve the item owning an attribute through the READ interface.
    #     """
    #     attribute_id = attribute["id"]
    #     endpoint = f'attribute/{attribute_id}/item'
    #
    #     r = client.read(endpoint=endpoint)
    #     assert r.is_ok()
    #     assert r.get_type() == dict
    #
    #     read_item = r.get_content()
    #     assert read_item["id"] == expected_item["id"]

    @pytest.mark.parametrize("attribute_id", ["foo_attribute_id"], scope="function")
    def test_attribute_get_owner_item_from_non_existent_attribute(self, client: ApiClient, attribute_id):
        """
        Tests if it's possible to correctly retrieve the item owning an attribute through the READ interface.
        """
        endpoint = f'attribute/{attribute_id}/item'

        r = client.read(endpoint=endpoint)
        assert r.is_not_found()

    @pytest.mark.parametrize("attribute",
                             [testAttributes["attribute3"], testAttributes["attribute4"]],
                             scope="function")
    def test_attribute_read_by_id(self, client: ApiClient, attribute):
        """
        Tests if it's possible to correctly retrieve the attributes by ID through the READ interface.
        """
        attribute_id = attribute["id"]
        attribute_name = attribute["name"]
        attribute_value = attribute["value"]

        r = self.read_attribute_res(client=client, attribute_id=attribute_id)
        assert r.is_ok()
        assert r.get_type() == dict

        read_attr = r.get_content()
        assert read_attr["id"] == attribute_id
        assert read_attr["name"] == attribute_name
        assert read_attr["value"] == attribute_value

    @pytest.mark.parametrize("attribute_id", ["foo_attribute_id"], scope="function")
    def test_attribute_read_by_id_non_existent(self, client: ApiClient, attribute_id):
        """
        Test the attribute read by ID through the READ interface against a non-existing attribute.
        """
        r = self.read_attribute_res(client=client, attribute_id=attribute_id)
        assert r.is_not_found()

    @pytest.mark.parametrize("attribute, updated_attribute",
                             [(testAttributes["attribute1"], testAttributesUpd["attribute1_updated"])],
                             scope="function")
    def test_attribute_partial_update(self, client: ApiClient, attribute, updated_attribute):
        """
        Tests the partial update of an attribute through the UPDATE interface.
        """
        r = self.update_attribute_res(client=client, attribute=attribute, updated_attribute=updated_attribute)
        assert r.is_ok()
        assert r.get_type() == dict

        r_dict = r.get_content()
        for key in ("name", "value_type", "value", "modified_by"):
            assert r_dict[key] == updated_attribute[key]

    @pytest.mark.parametrize("attribute_id", ["non_existent_attr"], scope="function")
    def test_attribute_partial_update_nonexistent(self, client: ApiClient, attribute_id):
        """
        Checks that a partial update of a non existent attribute through the UPDATE interface raises the proper error.
        """
        endpoint = f'attribute/{attribute_id}'

        updated_attribute = dict()
        r = client.update(
            endpoint=endpoint,
            payload=updated_attribute
        )
        assert r.get_type() == dict
        assert r.is_not_found()

    @pytest.mark.parametrize("attribute", [testAttributes["attribute2"]], scope="function")
    def test_attribute_read_value(self, client: ApiClient, attribute):
        """
        Tests the value reading from an attribute through the READ interface.
        """
        attribute_id = attribute["id"]
        endpoint = f"attribute/{attribute_id}/value"

        r = client.read(endpoint=endpoint)
        assert r.is_ok()
        assert r.get_type() == dict

        r_dict = r.get_content()
        for key in ("value_type", "value"):
            assert r_dict[key] == attribute[key]

    @pytest.mark.parametrize("attribute", [{"id": "foo_attribute_id"}], scope="function")
    def test_attribute_read_value_from_non_existent_attribute(self, client: ApiClient, attribute):
        """
        Tests the value reading from an attribute through the READ interface, against a non existing attribute.
        """
        attribute_id = attribute["id"]
        endpoint = f"attribute/{attribute_id}/value"

        r = client.read(endpoint=endpoint)
        assert r.is_not_found()

    # After this test, the 2nd test item will be deleted.
    def test_item_delete(self, client: ApiClient, item_id=testItems["item2"]["id"]):
        """
        Tests the DELETE interface to delete an item.
        """
        assert self.delete_item(client, item_id)

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
    def test_attributes_delete_nonexistent(self, client: ApiClient, attribute_id):
        """
        Tests the DELETE interface to delete an attribute. Since no more attributes should exist after the deletion of
        items, checks that the client raises a 'not found' error.
        """
        r = self.delete_attribute_res(client=client, attribute_id=attribute_id)
        assert r.is_not_found()

    def test_attributes_get_count_after_deletion(self, client: ApiClient):
        """
        Checks if the count counts 0.
        """
        endpoint = "attributes/count"
        r = client.read(endpoint=endpoint)
        assert r.is_ok()
        assert r.get_type() == dict

        r_dict = r.get_content()
        assert r_dict["count"] == 0
