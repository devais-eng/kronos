# Integration tests
The current directory contains kronos-server integration tests.

## Test reports

The `tests/report` directory is intended to store the Pytest test reports. 
`conftest.py` can be modified to customize the reports content.

To make a test produce a report, simply add the _html_ argument in the 
pytest run command.

Example:
```Shell
python3 -m pytest integration/test_items_apis.py --html report/integration/test_items_apis_report.html
```

## Environment variables
Please make sure that the following environment variables are properly set
before starting any integration test:

```dotenv
# Server API settings (HTTP)
TEST_SERVER_APP_PROTOCOL="http"
TEST_SERVER_APP_HOSTNAME="localhost"
TEST_SERVER_APP_PORT=8081
TEST_SERVER_APP_API_PATH="api/v1"
# Gateway API settings (HTTP)
TEST_GATEWAY_APP_PROTOCOL="http"
TEST_GATEWAY_APP_HOSTNAME="localhost"
TEST_GATEWAY_APP_PORT=5000
```

## Test dataset
The data used to build requests during integration tests is based on 
`tests/integration/test_dataset.yml` content.

The provided dataset simulates a bike sharing hierarchy. The items can be
either bikes or bike stations. The relations bound bikes with bike stations.
The attributes contain properties such as bike stations location and bikes
tyre pressure.