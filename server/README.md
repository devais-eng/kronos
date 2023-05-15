Kronos
=========

![image](./assets/logo.svg)

Overview
-------------------------------
This repository contains the code of the Kronos cloud application.

This application requires a MQTT broker to communicate with the Kronos edge service.

Entities can be created, updated and deleted locally through HTTP APIs.

Requirements
-------------------------------
This project requires [Python 3.7+](https://www.python.org/downloads/)

Build and launch
---
To build the application it is sufficient to run `docker-compose build`. To launch it run`docker-compose up`. Between
subsequents session for test purposes, it's suggested to
run `docker-compose down -v && docker-compose build kronos-backend && docker-compose up`.

Usage
---
When the application is up and running, it's possible to access an interface to try the APIs
at [localhost:8081/docs](http://localhost:8081/docs). It's even possible to check the models at bottom of the page.


HTTP APIs
-------------------------------
HTTP APIs are implemented using FastAPI framework.

Using built-in docker-compose configuration HTTP APIs are bound to `localhost:8080/docs`

Here you have a complete example:

<h5>Create 2 Items</h5>

```Bash
curl -X 'POST' \
  --cert 'path to cert'\
  --key 'path to key'
  'http://localhost:8081/api/v1/items' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '[
{
  "name": "test_item_1",
  "type": "item",
  "customer_id": "test_customer",
  "sync_policy": "test_sync_policy",
  "objects": [],
  "id": "test_item_1_id",
  "source_timestamp": 0,
  "created_by": "test"
},
{
  "name": "test_item_2",
  "type": "item",
  "customer_id": "test_customer",
  "sync_policy": "test_sync_policy",
  "objects": [],
  "id": "test_item_2_id",
  "source_timestamp": 0,
  "created_by": "test"
}
]'
```

Will give the response body:

```JSON
[
  {
    "name": "test_item_1",
    "type": "item",
    "customer_id": "test_customer",
    "sync_policy": "test_sync_policy",
    "edge_mac": null,
    "objects": [],
    "id": "test_item_1_id",
    "created_by": "test",
    "created_at": 1623418083,
    "modified_by": null,
    "modified_at": null,
    "version": "9834a5ae-37a6-4895-8a4e-0131f45ada47"
  },
  {
    "name": "test_item_2",
    "type": "item",
    "customer_id": "test_customer",
    "sync_policy": "test_sync_policy",
    "edge_mac": null,
    "objects": [],
    "id": "test_item_2_id",
    "created_by": "test",
    "created_at": 1623418084,
    "modified_by": null,
    "modified_at": null,
    "version": "08b624ad-dfde-4b1f-9dc2-7ec61011aec7"
  }
]
```

<h5>Get All Items</h5>

```Bash
curl -X 'GET' \
  'http://localhost:8081/api/v1/items?skip=0&limit=5' \
  -H 'accept: application/json'
```

Will give you the response body:

```JSON
[
  {
    "name": "test_item_1",
    "type": "item",
    "customer_id": "test_customer",
    "sync_policy": "test_sync_policy",
    "edge_mac": null,
    "objects": [],
    "id": "test_item_1_id",
    "created_by": "test",
    "created_at": 1623418083,
    "modified_by": null,
    "modified_at": null,
    "version": "9834a5ae-37a6-4895-8a4e-0131f45ada47"
  },
  {
    "name": "test_item_2",
    "type": "item",
    "customer_id": "test_customer",
    "sync_policy": "test_sync_policy",
    "edge_mac": null,
    "objects": [],
    "id": "test_item_2_id",
    "created_by": "test",
    "created_at": 1623418084,
    "modified_by": null,
    "modified_at": null,
    "version": "08b624ad-dfde-4b1f-9dc2-7ec61011aec7"
  }
]
```

<h5>Create a Relation between the 2 Items</h5>

```Bash
curl -X 'POST' \
  'http://localhost:8081/api/v1/relations' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "parent_id": "test_item_1_id",
  "child_id": "test_item_2_id",
  "source_timestamp": 0,
  "created_by": "test"
}'
```

Will give you the response body:

```JSON
[
  {
    "parent_id": "test_item_1_id",
    "child_id": "test_item_2_id",
    "id": "test_item_1_id -> test_item_2_id",
    "created_by": "test",
    "created_at": 1623517863,
    "modified_by": null,
    "modified_at": null,
    "version": "feedcd21-a5b6-427f-9305-97aa6e9b55f9"
  }
]
```

<h5>Get all Relations</h5>

```Bash
curl -X 'GET' \
  'http://localhost:8081/api/v1/relations?skip=0&limit=100' \
  -H 'accept: application/json'
```

Will give you the response body:

```JSON
[
  {
    "parent_id": "test_item_1_id",
    "child_id": "test_item_2_id",
    "id": "test_item_1_id -> test_item_2_id",
    "created_by": "test",
    "created_at": 1623517863,
    "modified_by": null,
    "modified_at": null,
    "version": "feedcd21-a5b6-427f-9305-97aa6e9b55f9"
  }
]
```

<h5>Create an Object attached to the first Item</h5>

```Bash
curl -X 'POST' \
  'http://localhost:8081/api/v1/objects' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "instance_id": "test_instance_object_1",
  "type": "object",
  "name": "test_object1",
  "resources": [],
  "id": "test_object_1_id",
  "source_timestamp": 0,
  "created_by": "test",
  "item_id": "test_item_1_id"
}'
```

Will give you the response body:

```JSON
[
  {
    "instance_id": "test_instance_object_1",
    "type": "object",
    "name": "test_object1",
    "resources": [],
    "id": "test_object_1_id",
    "created_by": "test",
    "created_at": 1623518021,
    "modified_by": null,
    "modified_at": null,
    "version": "0f027530-a8b6-4405-9a82-4bf4c3d78925",
    "item_id": "test_item_1_id"
  }
]
```

<h5>Create a Resource attached to the Object</h5>

```Bash
curl -X 'POST' \
  'http://localhost:8081/api/v1/resources' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "name": "test_resource_1",
  "type": "resource",
  "sync_policy": "test_sync_policy",
  "value": "",
  "value_type": "test_value_type",
  "source_timestamp": 0,
  "created_by": "test",
  "id": "test_resource_1_id",
  "object_id": "test_object_1_id"
}'
```

Will give you the response body:

```JSON
[
  {
    "name": "test_resource_1",
    "type": "resource",
    "sync_policy": "test_sync_policy",
    "value": "",
    "value_type": "test_value_type",
    "id": "test_resource_1_id",
    "created_by": "test",
    "created_at": 1623518118,
    "modified_by": null,
    "modified_at": null,
    "version": "7ea22746-665f-4910-85c3-91091c0b7bd0"
  }
]
```

<h5>Get all Resources</h5>

```Bash
curl -X 'GET' \
  'http://localhost:8081/api/v1/resources?skip=0&limit=2' \
  -H 'accept: application/json'
```

Will give you the response body:

```JSON
[
  {
    "name": "test_resource_1",
    "type": "resource",
    "sync_policy": "test_sync_policy",
    "value": "",
    "value_type": "test_value_type",
    "id": "test_resource_1_id",
    "created_by": "test",
    "created_at": 1623518118,
    "modified_by": null,
    "modified_at": null,
    "version": "7ea22746-665f-4910-85c3-91091c0b7bd0"
  }
]
```

Internal Kafka Topics
-------------------------------
Kronos internally use two topics:

* `ServiceRequest`: handle incoming message and redirect to the right application.
* `ServiceReply`: handle response from the service application

#### ServiceRequest's Data Structure

The topic's data structure for the ServiceRequest topic is:

```JSON
{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "type": "object",
  "properties": {
    "request_id": {
      "type": "string"
    },
    "command": {
      "type": "string"
    },
    "ts": {
      "type": "integer"
    },
    "request_type": {
      "type": "string"
    },
    "payload": {
      "type": "object"
    }
  },
  "required": [
    "request_id",
    "command",
    "ts",
    "request_type",
    "payload"
  ]
}
```

For example, the CRUD's json message can be:

```JSON
{
  "request_id": "a7ae56b1-67f5-4b1e-bce5-1030ed5d859b",
  "command": "CRUD",
  "ts": 1632233624,
  "request_type": "TransactionApp",
  "payload": [
    {
      "tx_type": "ENTITY_UPDATED",
      "entity_id": "string",
      "id": 1,
      "entity_type": "ITEM",
      "timestamp": 1632233624140,
      "triggered_by": "SYNC",
      "tx_uuid": null,
      "tx_len": 1,
      "tx_index": 0,
      "body": {
        "customer_id": "string",
        "edge_mac": "string",
        "id": "string",
        "modified_by": "SYNC",
        "name": "string",
        "sync_policy": "string",
        "sync_version": "e1461028-df8e-4461-bdeb-e926bc311214",
        "type": "string",
        "version": "e1461028-df8e-4461-bdeb-e926bc311214"
      }
    }
  ]
}
```

#### ServiceReply Data Structure

```JSON
{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "type": "object",
  "properties": {
    "caller_request_id": {
      "type": "string"
    },
    "response_id": {
      "type": "string"
    },
    "ts": {
      "type": "number"
    },
    "response_code": {
      "type": "integer"
    },
    "response_type": {
      "type": "string"
    },
    "response_body": {
      "type": "array",
      "items": [
        {
          "type": "object"
        }
      ]
    },
    "execution_time": {
      "type": "number"
    }
  }
}
```

For example, the CRUD's json response can be:

```JSON
{
  "caller_request_id": "1033e006-1ae6-11ec-ac0b-0242ac12000a",
  "response_id": "10d057ce-1ae6-11ec-9131-0242ac120008",
  "ts": 1632233596.8532665,
  "response_code": 0,
  "response_type": "TransactionApp",
  "response_body": [
    {
      "entity_type": "ITEM",
      "entity_id": "string",
      "version": "4ead15e7-0991-45ae-a146-7197797d245a",
      "action": "CREATE",
      "payload": {
        "id": "string",
        "name": "string",
        "type": "string",
        "customer_id": "string",
        "sync_policy": "string",
        "edge_mac": "string",
        "attributes": []
      }
    }
  ],
  "execution_time": 0.16830062866210938
}
```

Running
-------------------------------

The application can recover from panics, and report them to [Sentry](https://sentry.io/)


Sentry
-------------------------------
[Sentry](https://sentry.io/) can be enabled by provide `SENTRY_DSN` (uri of project registered in Sentry)


API Testing - Tavern
---
Tavern is a pytest plugin aimed in API testing, supporting HTTP, MQTT and other protocols. It's really easy to use and
extend in a scalable fashion.

## Installation

`pip install tavern`

## File structure

Tavern uses a superclass of `.yaml`. Every test file **must** begin with the word **test** and have `.tavern.yaml`
extension, es. `test_x.tavern.yaml`, where `x` is the description of what the test does. The anatomy of a common test is
well described in the [documentation](https://tavern.readthedocs.io/en/latest/basics.html#anatomy-of-a-test). In this
case, a `common.yaml` is used to give global configurations to all the tests. Both the **test files** and the **utils
functions** use **test cases configurations** depending on the tests' scope.

## Usage

To run a test on Kronos' Server, it's sufficient to run `docker-compose up` and run the
command `tavern-ci test_x.tavern.yaml -v`. The Verbose flag is really helpful in debugging purpose.

## Test the APIs

To test all the APIs it's sufficient to go into tests/unit/rest and run the following commands depending on the
respective entity to test:

- `tavern-ci test_item_api.tavern -v` to test item entities.
- `tavern-ci test_resource_api.tavern.yaml -v` to test resource entities.
- `tavern-ci test_relation_api.tavern.yaml -v` to test relation entities.
- `tavern-ci test_parameter_api.tavern.yaml -v` to test parameter entities.

End-to-end Tests
---
In order to ensure simplicity and maintainability, the end-to-end tests are structured by dividing them into separate files for each entity. This approach helps in organizing the tests and allows for a focused examination of each entity's functionality.

To support the testing process, the tests utilize utility functions provided by the `tests/utils` package. These utility functions likely provide common functionalities and helper methods that can be shared across multiple test files. By leveraging these utilities, the tests can be written more efficiently and avoid duplication of code.

This modularized approach to end-to-end testing enhances code organization, readability, and reusability. It promotes a clear separation of concerns, enabling developers to focus on specific entities and their associated test cases. Moreover, by centralizing common functionalities in the utility package, maintenance efforts are reduced, and updates can be applied consistently throughout the test suite.
## Usage

To run the tests is sufficient to go into the tests directory, run `docker-compose up` on kronos-server directory,
launch the edge and then run `pytest integration`.
`
