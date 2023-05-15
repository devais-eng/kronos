import os
import time

import pytest
from py._xmlgen import html  # noqa
from datetime import datetime
from .settings import settings


# Pytest settings
def pytest_addoption(parser):  # noqa
    parser.addoption(
        "--runslow", action="store_true", default=False, help="run slow tests",
    )


def pytest_configure(config):
    config.addinivalue_line("markers", "slow: mark test as slow to run")


def pytest_collection_modifyitems(config, items):  # noqa
    if config.getoption("--runslow"):
        # --runslow given in cli: do not skip slow tests
        return
    skip_slow = pytest.mark.skip(reason="need --runslow option to run")
    for item in items:
        if "slow" in item.keywords:
            item.add_marker(skip_slow)


# Add a "time" column to understand the tests execution order
@pytest.mark.optionalhook
def pytest_html_results_table_header(cells):
    cells.insert(3, html.th('Time', class_='sortable time', col='time'))
    cells.pop()


@pytest.mark.optionalhook
def pytest_html_results_table_row(report, cells):
    cells.insert(3, html.td(datetime.utcnow(), class_='col-time'))
    cells.pop()


# pytest-docker plugin configuration
@pytest.fixture(scope="session")
def docker_compose_file(pytestconfig):
    """
    This fixture overrides the pytest-docker built-in fixture to provide the docker-compose.yml file path
    """
    if settings.DOCKER_COMPOSE_FILE_PATH is None:
        project_path = pytestconfig.rootdir
        filepath = os.path.join(
            str(project_path),
            "docker-compose.yml"
        )

        return filepath
    return settings.DOCKER_COMPOSE_FILE_PATH


@pytest.fixture(scope="session")
def docker_compose_project_name():
    """
    Override the built-in fixture to provide a particular project name.
    """

    return settings.DOCKER_COMPOSE_PROJECT_NAME


@pytest.fixture(scope="session")
def docker_cleanup(pytestconfig):
    """
    This fixture overrides the docker_compose command to be executed for test clean-up actions.
    """
    return settings.DOCKER_COMPOSE_CLEAN_UP_POLICY


@pytest.fixture(scope="session", autouse=True)
def run_server_backend(docker_ip, docker_services):
    """
    Start the dockers using docker-compose and wait for its connection.
    """

    from tests.utils import is_ready
    docker_compose_up_timeout = 30.0
    docker_compose_up_check_interval = 5.0
    docker_compose_up_request_timeout = 5.0

    docker_services.wait_until_responsive(
        timeout=docker_compose_up_timeout,  # Allowed time to go up before interrupting the test setup
        pause=docker_compose_up_check_interval,  # Pause between checks
        check=lambda: is_ready(docker_compose_up_request_timeout)  # Method to check if the container is up and running
    )
    time.sleep(docker_compose_up_request_timeout * 2)
    return
