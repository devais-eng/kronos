import yaml
from typing import Dict, List, Union
from tests.settings import settings
from tests.utils import Action, Clients, ApiClient, HttpServerProtocol, Response

# Protocols and clients setup
actions_list = list(Action)
protocols_list = [HttpServerProtocol(settings.SERVER_APP_BASE_URL)]
api_clients = Clients(actions_list, protocols_list)

# Load test dataset
with open('./integration/test_dataset.yaml') as file:
    testDataset = yaml.load(file, Loader=yaml.FullLoader)
testItems = testDataset["test_items"]
testItemsUpd = testDataset["test_items_updated"]
testAttributes = testDataset["test_attributes"]
testAttributesUpd = testDataset["test_attributes_updated"]
testRelations = testDataset["test_relations"]


# Base class for pytest test classes
class Test:
    @staticmethod
    def items_gen(start_no=1, end_no=10) -> List[Dict]:
        """
        Generates a list of items with a standard schema
        """
        items_list = list()
        for i in range(start_no, end_no):
            item = {
                "id": f"test-item{i}-gen-id",
                "name": f"test-item{i}-gen-name",
                "type": "test-item-type",
                "customer_id": "test-customer-1",
                "sync_policy": "DEFAULT_SYNC",
                "edge_mac": "th:is:is:aa:te:st",
                "created_by": "test-routine",
                "source_timestamp": 0,
                "objects": []
            }
            items_list.append(item)
        return items_list

    @classmethod
    def multiple_items_set_150(cls) -> List[Dict]:
        return cls.items_gen(start_no=1, end_no=150)

    # Read methods
    @staticmethod
    def read_item_res(client: ApiClient, item_id: str) -> Response:
        """
        Performs a read operation to read an item.
        Returns the Response object.
        """
        endpoint = f'item/{item_id}'

        return client.read(endpoint=endpoint)

    @staticmethod
    def read_attribute_res(client: ApiClient, attribute_id: str) -> Response:
        """
        Performs a READ operation to read the attribute with the given id.
        Returns the Response object.
        """
        endpoint = f'attribute/{attribute_id}'

        return client.read(
            endpoint=endpoint
        )

    @staticmethod
    def read_attributes_res(client: ApiClient) -> Response:
        """
        Performs a read operation to read the items.
        Returns the list of items.
        """
        endpoint = "attributes"
        params = {
            "page": 1,
            "page_size": 200
        }

        return client.read(endpoint=endpoint, params=params)

    @staticmethod
    def read_relation_res(client: ApiClient, rel_parent_id: str, rel_child_id: str) -> Response:
        """
        Performs a READ operation to read the relation with the given parent and child ids.
        Returns the Response object.
        """
        endpoint = f'relation/{rel_parent_id}/{rel_child_id}'
        return client.read(endpoint=endpoint)

    @staticmethod
    def read_relations_res(client: ApiClient) -> Response:
        """
        Performs a read operation to read the relations.
        Returns the list of relations.
        """
        endpoint = "relations"
        params = {
            "page": 1,
            "page_size": 200
        }

        return client.read(endpoint=endpoint, params=params)

    # Existence assert methods
    @classmethod
    def item_exists(cls, client: ApiClient, item_id: str) -> bool:
        """
        Performs a read operation to read the item with the given id.
        Returns True if the item has been found.
        """

        def check(res: Response) -> bool:
            checks = list()
            checks.append(res.is_ok())
            checks.append(res.get_type() == dict)
            try:
                res_body = res.get_content()
                checks.append(res_body["id"] == item_id)
            except KeyError:
                checks.append(False)
            return False not in checks

        return check(
            cls.read_item_res(client=client, item_id=item_id)
        )

    @classmethod
    def attribute_exists(cls, client: ApiClient, attribute_id: str) -> bool:
        """
        Performs a READ operation to read the attribute with the given id.
        Returns True if the item has been found.
        """

        def check(res: Response) -> bool:
            checks = list()
            checks.append(res.is_ok())
            checks.append(res.get_type() == dict)
            try:
                res_body = res.get_content()
                checks.append(res_body["id"] == attribute_id)
            except KeyError:
                checks.append(False)
            return False not in checks

        return check(
            cls.read_attribute_res(client=client, attribute_id=attribute_id)
        )

    @classmethod
    def relation_exists(cls, client: ApiClient, rel_parent_id: str, rel_child_id: str) -> bool:
        """
        Returns true if the given client is able to find the given relation.
        """

        def check(res: Response) -> bool:
            checks = list()
            checks.append(res.is_ok())
            checks.append(res.get_type() == dict)
            if False not in checks:
                try:
                    res_body = res.get_content()
                    checks.append(res_body["parent_id"] == rel_parent_id)
                    checks.append(res_body["child_id"] == rel_child_id)
                except KeyError:
                    checks.append(False)
            return False not in checks

        return check(
            cls.read_relation_res(client=client, rel_child_id=rel_child_id, rel_parent_id=rel_parent_id)
        )

    # Creation methods
    @staticmethod
    def create_items_res(client: ApiClient, items: Union[Dict, List[Dict]]) -> Response:
        """
        Triggers the CREATE interface to create an item or more than one (given as a list of items).
        Returns true if success.
        """
        endpoint = "items"
        payload = items

        return client.create(endpoint=endpoint, payload=payload)

    @classmethod
    def create_items(cls, client: ApiClient, items: Union[Dict, List[Dict]]) -> bool:
        """
        Triggers the CREATE interface to create an item or more than one (given as a list of items).
        Returns true if success.
        """

        def check(res: Response) -> bool:
            checks = list()
            checks.append(res.is_ok())
            return False not in checks

        return check(
            cls.create_items_res(client=client, items=items)
        )

    @staticmethod
    def create_attributes_res(client: ApiClient, attributes: Union[Dict, List[Dict]]) -> Response:
        """
        Triggers the CREATE interface to create an attribute or more than one (given as a list of items).
        Returns the response object.
        """
        endpoint = "attributes"
        payload = attributes

        return client.create(endpoint=endpoint, payload=payload)

    @classmethod
    def create_attributes(cls, client: ApiClient, attribute: Union[Dict, List[Dict]]) -> bool:
        """
        Triggers the CREATE interface to create an attribute or more than one (given as a list of items).
        Returns true if success.
        """

        def check(res: Response) -> bool:
            checks = list()
            checks.append(res.is_ok())
            return False not in checks

        return check(
            cls.create_attributes_res(client=client, attributes=attribute)
        )

    @staticmethod
    def create_relation_res(client: ApiClient, relation: Dict) -> Response:
        """
        Triggers the CREATE interface to create a relation between two items.
        Returns the Response object.
        """
        endpoint = "relations"
        payload = {
            "parent_id": relation["parent_id"],
            "child_id": relation["child_id"],
            "created_by": relation["created_by"]
        }
        return client.create(endpoint=endpoint, payload=payload)

    @classmethod
    def create_relation(cls, client: ApiClient, relation: Dict) -> bool:
        """
        Triggers the CREATE interface to create a relation between two items.
        Returns True if success.
        """

        def check(res: Response) -> bool:
            checks = list()
            checks.append(res.is_ok())
            return False not in checks

        return check(
            cls.create_relation_res(client=client, relation=relation)
        )

    # Partial update methods
    @staticmethod
    def update_item_res(client: ApiClient, item: Dict, patched_item: Dict) -> Response:
        """
        Tests the UPDATE interface to partially update an item.
        Returns the Response object.
        """
        item_id = item["id"]
        endpoint = f'item/{item_id}'
        return client.update(
            endpoint=endpoint,
            payload=patched_item
        )

    @classmethod
    def update_item(cls, client: ApiClient, item: Dict, patched_item: Dict) -> bool:
        """
        Tests the UPDATE interface to partially update an item.
        Returns True if success.
        """

        def check(res: Response) -> bool:
            checks = list()
            checks.append(res.is_ok())
            return False not in checks

        return check(
            cls.update_item_res(client=client, item=item, patched_item=patched_item)
        )

    @staticmethod
    def update_attribute_res(client: ApiClient, attribute: Dict, updated_attribute: Dict) -> Response:
        """
        Tests the UPDATE interface to partially update an attribute.
        Returns the Response object.
        """
        attribute_id = attribute["id"]
        endpoint = f'attribute/{attribute_id}'
        return client.update(
            endpoint=endpoint,
            payload=updated_attribute
        )

    @classmethod
    def update_attribute(cls, client: ApiClient, attribute: Dict, updated_attribute: Dict) -> bool:
        """
        Tests the UPDATE interface to partially update an attribute.
        Returns True if success.
        """

        def check(res: Response) -> bool:
            checks = list()
            checks.append(res.is_ok())
            return False not in checks

        return check(
            cls.update_attribute_res(client=client, attribute=attribute, updated_attribute=updated_attribute)
        )

    # Deletion methods
    @staticmethod
    def delete_item_res(client: ApiClient, item_id: str) -> Response:
        """
        Triggers the DELETE interface to delete an item.
        Returns the Response object.
        """
        endpoint = f'item/{item_id}'
        return client.delete(endpoint=endpoint)

    @classmethod
    def delete_item(cls, client: ApiClient, item_id: str) -> bool:
        """
        Triggers the DELETE interface to delete an attribute.
        Returns true if success.
        """

        def check(res: Response) -> bool:
            res_body = res.get_content()

            checks = list()
            checks.append(res.is_ok())
            checks.append(res_body["id"] == item_id)
            checks.append(res.get_type() == dict)
            return False not in checks

        return check(
            cls.delete_item_res(client=client, item_id=item_id)
        )

    @staticmethod
    def delete_attribute_res(client: ApiClient, attribute_id: str) -> Response:
        """
        Triggers the DELETE interface to delete an attribute.
        Returns the Response object.
        """
        endpoint = f"attribute/{attribute_id}"
        return client.delete(endpoint=endpoint)

    @classmethod
    def delete_attribute(cls, client: ApiClient, attribute_id: str) -> bool:
        """
        Triggers the DELETE interface to delete an attribute.
        Returns true if success.
        """

        def check(res: Response) -> bool:
            res_body = res.get_content()

            checks = list()
            checks.append(res.is_ok())
            checks.append(res_body["id"] == attribute_id)
            checks.append(res.get_type() == dict)
            return False not in checks

        return check(
            cls.delete_attribute_res(client=client, attribute_id=attribute_id)
        )

    @staticmethod
    def delete_relation_res(client: ApiClient, relation: Dict) -> Response:
        """
        Triggers the DELETE interface to delete a relation between two items.
        Returns the Response object.
        """
        parent_id = relation["parent_id"]
        child_id = relation["child_id"]
        endpoint = f'relation/{parent_id}/{child_id}'
        return client.delete(endpoint=endpoint)

    @classmethod
    def delete_relation(cls, client: ApiClient, relation: Dict) -> bool:
        """
        Triggers the DELETE interface to delete a relation between two items.
        Returns True if success.
        """

        def check(res: Response) -> bool:
            checks = list()
            checks.append(res.is_ok())
            return False not in checks

        return check(
            cls.delete_relation_res(client=client, relation=relation)
        )
