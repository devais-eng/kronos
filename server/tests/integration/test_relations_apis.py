import pytest

from tests.utils import ApiClient
from tests.integration import Test, api_clients
from tests.integration import testItems, testRelations


# Notes:
# _Scope = Literal["session", "package", "module", "class", "function"]
# - with scope "function" we'll have each test executed with the different clients consequently
# - with scope "class" we'll have the whole class executed each time with a different client
@pytest.fixture(scope="class", params=api_clients)
def client(request):
    yield request.param


class TestRelations(Test):
    def test_relations_read_all_empty(self, client: ApiClient):
        """
        Checks that the attributes list is empty.
        """
        r = self.read_relations_res(client=client)
        attr_list = r.get_content()
        assert isinstance(attr_list, list)
        assert len(attr_list) == 0

    def test_relations_get_count_empty(self, client: ApiClient):
        """
        Checks if the count counts 0.
        """
        endpoint = "relations/count"
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

    # After this test, the test relations #1,2 described in the YAML file will be added to the database
    @pytest.mark.parametrize("relation",
                             [testRelations["relation1"], testRelations["relation2"]],
                             scope="function")
    def test_relation_create(self, client: ApiClient, relation):
        """
        Tests the creation of a relation.
        """
        assert self.create_relation(client=client, relation=relation)

    @pytest.mark.parametrize("relation",
                             [testRelations["relation1"], testRelations["relation2"]],
                             scope="function")
    def test_read_relation(self, client: ApiClient, relation):
        """
        Checks the creation of a relation.
        """
        assert self.relation_exists(client=client,
                                    rel_child_id=relation["child_id"],
                                    rel_parent_id=relation["parent_id"])

    @pytest.mark.parametrize("relation",
                             [{"child_id": "foo", "parent_id": "bar"}],
                             scope="function")
    def test_read_non_existent_relation(self, client: ApiClient, relation):
        """
        Tests the reading of a relation which does not exist.
        """
        res = self.read_relation_res(client=client,
                                     rel_child_id=relation["child_id"],
                                     rel_parent_id=relation["parent_id"])
        assert res.is_not_found()

    @pytest.mark.parametrize("relation", [testRelations["relation1"]], scope="function")
    def test_relation_create_existent(self, client: ApiClient, relation):
        """
        Tests the creation of a relation which already exists.
        """
        r = self.create_relation_res(client=client, relation=relation)
        assert r.is_bad_request()

    # After this test, the test relation #1 described in the YAML file will be deleted from the database
    @pytest.mark.parametrize("relation", [testRelations["relation1"]], scope="function")
    def test_relation_delete(self, client: ApiClient, relation):
        """
        Tests the deletion of a relation.
        """
        assert self.delete_relation(client=client, relation=relation)

    @pytest.mark.parametrize("relation", [testRelations["relation1"]], scope="function")
    def test_relation_delete_assert(self, client: ApiClient, relation):
        """
        Verify the deletion of a relation.
        """
        assert not self.relation_exists(client=client,
                                        rel_parent_id=relation["parent_id"],
                                        rel_child_id=relation["child_id"])

    @pytest.mark.parametrize("relation", [testRelations["relation1"]], scope="function")
    def test_relation_delete_nonexistent(self, client: ApiClient, relation):
        """
        Tests the deletion of a relation which does not exist.
        Checks that the client returns a 'not found' error.
        """
        r = self.delete_relation_res(client=client, relation=relation)
        assert r.is_not_found()

    def test_relations_read_all(self, client: ApiClient):
        """
        Checks if the expected 1 relations can be retrieved.
        """
        r = self.read_relations_res(client=client)
        attr_list = r.get_content()
        assert isinstance(attr_list, list)
        assert len(attr_list) == 1

    # @pytest.mark.skip("Skipping because the server's response must be changed from Int to Dict.")
    def test_relations_get_count(self, client: ApiClient):
        """
        Checks if the count counts 1.
        """
        endpoint = "relations/count"
        r = client.read(endpoint=endpoint)
        assert r.is_ok()
        assert r.get_type() == dict

        r_dict = r.get_content()
        assert r_dict["count"] == 1

    # After this test, the 2nd test item will be deleted.
    def test_item_delete(self, client: ApiClient, item_id=testItems["item2"]["id"]):
        """
        Tests the DELETE interface to delete an item.
        """
        assert self.delete_item(client, item_id)

    @pytest.mark.parametrize("relation", [testRelations["relation1"], testRelations["relation2"]], scope="function")
    def test_item_delete_cascade_assert(self, client: ApiClient, relation):
        """
        Asserts that the cascade deletion of relations works correctly.
        """
        # Check that the relation has been deleted
        assert not self.relation_exists(client, rel_parent_id=relation["parent_id"], rel_child_id=relation["child_id"])

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
    @pytest.mark.parametrize("relation", [testRelations["relation1"], testRelations["relation2"]], scope="function")
    def test_relation_delete_nonexistent(self, client: ApiClient, relation):
        """
        Tests the DELETE interface to delete a relation. Since no more relations should exist after the deletion of
        items, checks that the client raises a 'not found' error.
        """
        r = self.delete_relation_res(client=client, relation=relation)
        assert r.is_not_found()

    def test_relations_get_count_after_deletion(self, client: ApiClient):
        """
        Checks if the count counts 0.
        """
        endpoint = "relations/count"
        r = client.read(endpoint=endpoint)
        assert r.is_ok()
        assert r.get_type() == dict

        r_dict = r.get_content()
        assert r_dict["count"] == 0
