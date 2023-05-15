import json
import time
from abc import ABC
from collections.abc import Iterable, Iterator
from enum import Enum
from typing import Any, Optional, List, Dict, Tuple

import requests
from requests.sessions import HTTPAdapter
from simplejson.errors import JSONDecodeError
from urllib3 import Retry
from tests.integration import settings


class ResponseStatusCode(Enum):
    """
    Protocol response codes.
    """
    OK = 0
    ENTITY_NOT_FOUND = 1
    BAD_REQUEST = 2
    OTHER_ERROR = 3


class Response:
    def __init__(self, status_code: ResponseStatusCode, content):
        """
        Used as client response object.

        :param status_code: The response status code.
        :type status_code: ResponseStatusCode
        :param content: The response content.
        :type content: Any
        """
        self._status_code = status_code
        self._content = content
        self._content_type = type(content)

    def get_content(self):
        """
        Returns the response body.
        """
        return self._content

    def get_type(self):
        """
        Returns the content type.
        """
        return self._content_type

    def get_status(self):
        """
        Returns the request status.
        """
        return self._status_code

    def is_ok(self):
        return self._status_code == ResponseStatusCode.OK

    def is_error(self):
        return self.is_ok() is False

    def is_bad_request(self):
        return self._status_code == ResponseStatusCode.BAD_REQUEST

    def is_not_found(self):
        return self._status_code == ResponseStatusCode.ENTITY_NOT_FOUND

    def is_other_error(self):
        return self._status_code == ResponseStatusCode.OTHER_ERROR


class Action(str, Enum):
    def __str__(self):
        return str(self.value)

    CREATE = 'CREATE'
    READ = 'READ'
    UPDATE = 'UPDATE'
    REPLACE = 'REPLACE'
    DELETE = 'DELETE'


class Protocol(ABC):
    def __init__(self) -> None:
        super().__init__()
        self.actions = {
            Action.CREATE: self.create,
            Action.READ: self.read,
            Action.UPDATE: self.update,
            Action.REPLACE: self.replace,
            Action.DELETE: self.delete,
        }

    def close(self):
        pass

    def create(self, endpoint, *, params=None, payload=None) -> Response:
        pass

    def read(self, endpoint, *, params=None, payload=None) -> Response:
        pass

    def update(self, endpoint, *, params=None, payload=None) -> Response:
        pass

    def replace(self, endpoint, *, params=None, payload=None) -> Response:
        pass

    def delete(self, endpoint, *, params=None, payload=None) -> Response:
        pass

    def wait_sync(self):
        pass

    def get_action_method(self, action: Action) -> Any:
        return self.actions[action]

    @staticmethod
    def _build_response(*args) -> Response:
        pass


class HttpServerProtocol(Protocol):
    def __init__(
            self,
            server_host: str,
            timeout: int = 3,
            *,
            verify_session: bool = False,
            retry: Optional[Retry] = None,
    ) -> None:
        super().__init__()
        self.server_host = server_host

        self._adapter = HTTPAdapter(max_retries=retry)
        self._session = requests.Session()
        self._session.mount('http://', self._adapter)
        self._session.mount('https://', self._adapter)
        self._session.verify = verify_session
        self._session.headers.update({
            'User-Agent': 'Kronos-Tests-0.1',
        })

        self._timeout = timeout

    def __str__(self):
        return "HttpServerProtocol->" + str(self.server_host)

    def close(self):
        self._session.close()

    def create(self, endpoint, *, params=None, payload=None):
        res = self._session.post(self.server_host + '/' + endpoint, params=params, json=payload, timeout=self._timeout)
        self._log_request('POST', endpoint, params, payload, res)
        return self._build_response(res)

    def read(self, endpoint, *, params=None, payload=None):
        res = self._session.get(self.server_host + '/' + endpoint, params=params, timeout=self._timeout)
        self._log_request('GET', endpoint, params, payload, res)
        return self._build_response(res)

    def update(self, endpoint, *, params=None, payload=None):
        res = self._session.put(self.server_host + '/' + endpoint, params=params, json=payload, timeout=self._timeout)
        self._log_request('PUT', endpoint, params, payload, res)
        return self._build_response(res)

    def replace(self, endpoint, *, params=None, payload=None):
        res = self._session.put(self.server_host + '/' + endpoint, params=params, json=payload, timeout=self._timeout)
        self._log_request('PUT', endpoint, params, payload, res)
        return self._build_response(res)

    def delete(self, endpoint, *, params=None, payload=None):
        res = self._session.delete(self.server_host + '/' + endpoint, params=params, timeout=self._timeout)
        self._log_request('DELETE', endpoint, params, payload, res)
        return self._build_response(res)

    def wait_sync(self):
        time.sleep(0.65)

    @staticmethod
    def _build_response(requests_response: requests.Response) -> Response:
        """ Converts a requests response object into a Response. """
        requests_status_code = requests_response.status_code
        if requests_status_code in (200, 201):
            status_code = ResponseStatusCode.OK
        elif requests_status_code in (404, 422):
            status_code = ResponseStatusCode.ENTITY_NOT_FOUND
        elif requests_status_code in (400, 409, 421):
            status_code = ResponseStatusCode.BAD_REQUEST
        else:
            status_code = ResponseStatusCode.OTHER_ERROR
            requests_response.raise_for_status()

        try:
            content = requests_response.json()
        except JSONDecodeError:
            content = requests_response.text

        return Response(
            status_code=status_code,
            content=content
        )

    def _log_request(self, verb, endpoint, params, payload, requests_response):
        """ Provides log output for Pytest reports. """
        try:
            _ = requests_response.json()
            res_content = json.dumps(_, indent=4)
        except JSONDecodeError:
            res_content = requests_response.text
        print(
            '*' * 30, "Request details:", '*' * 30, "\n",
            "Protocol name: %s\n" % self.__class__.__name__,
            "Request type: %s\n" % verb,
            "Server host: %s\n" % self.server_host,
            "Request endpoint: %s\n" % endpoint,
            "Request parameters: %s\n" % json.dumps(params, indent=2),
            "Request payload: %s\n" % json.dumps(payload, indent=2),
            "Response code: %i\n" % requests_response.status_code,
            "Response content: %s\n" % res_content,
            '*' * 77,
        )


class ApiClient:
    def __init__(self, protocols: Dict[Action, Protocol]) -> None:
        self.protocols = protocols

    # def __init__(self, *, protocol: Protocol) -> None:
    #     self.protocols = {action: protocol for action in Action}

    def _get_action_method(self, action: Action):
        protocol = self.protocols[action]
        method = protocol.get_action_method(action)
        if action == Action.READ:
            return method

        def sync_method(endpoint, *, params=None, payload=None):
            res = method(endpoint, params=params, payload=payload)
            protocol.wait_sync()
            return res

        return sync_method

    def close(self):
        protocols_set = {p for p in self.protocols}
        for p in protocols_set:
            p.close()

    def create(self, endpoint, *, params=None, payload=None) -> Response:
        return self._get_action_method(Action.CREATE)(
            endpoint,
            params=params,
            payload=payload,
        )

    def read(self, endpoint, *, params=None, payload=None) -> Response:
        return self._get_action_method(Action.READ)(
            endpoint,
            params=params,
            payload=payload,
        )

    def update(self, endpoint, *, params=None, payload=None) -> Response:
        return self._get_action_method(Action.UPDATE)(
            endpoint,
            params=params,
            payload=payload,
        )

    def replace(self, endpoint, *, params=None, payload=None) -> Response:
        return self._get_action_method(Action.REPLACE)(
            endpoint,
            params=params,
            payload=payload,
        )

    def delete(self, endpoint, *, params=None, payload=None) -> Response:
        return self._get_action_method(Action.DELETE)(
            endpoint,
            params=params,
            payload=payload,
        )


class ProtocolIterator(Iterator):
    def __init__(self, actions, protocols) -> None:
        """
        This iterator will return clients for API testing iterating the available protocols over the available actions.

        :param actions: list of supported actions.
        :param protocols: list of supported protocols.
        """
        self.protocols_combinations = self._protocols_combinations(actions,
                                                                   protocols)
        self.action_protocols_combinations = self._action_protocols_combinations(actions,
                                                                                 self.protocols_combinations)
        self._index = 0

    def __next__(self):
        try:
            client = ApiClient(
                protocols=self.action_protocols_combinations[self._index]
            )
            client.__name__ = self._client_name()
            self._index += 1
        except IndexError:
            raise StopIteration()
        return client

    @staticmethod
    def _protocols_combinations(actions, protocols) -> List[Tuple[Protocol, ...]]:
        """
        Generates a list of tuples. Each tuple contains a different protocol repeated for the number of actions.

        :param actions: list of supported actions.
        :param protocols: list of supported protocols.
        :return: list of protocols tuples.

        :example:
        _protocols_combinations(actions=[a,b,c,d,e], protocols=[Dbus, Http])
            [('Dbus', 'Dbus', 'Dbus', 'Dbus', 'Dbus'),
            ('Http', 'Http', 'Http', 'Http', 'Http')]
        """
        n_actions = len(actions)
        comb = list()
        for p in protocols:
            p_list = []
            for i in range(n_actions):
                p_list.append(p)
            comb.append(tuple(p_list))
        return comb

    @staticmethod
    def _action_protocols_combination(actions, protocols) -> Dict[Action, Protocol]:
        """
        Associates a list of actions to a list of protocols. The result is a dict: {Action: Protocol}.

        :param actions: list of supported actions.
        :param protocols: list of supported protocols.
        :return: dict of form {Action: Protocol}.

        :example:
        _action_protocols_combination(actions=[a,b,c,d,e], protocols=[p1,p2,p3,p4,p5])
            {
                a: p1,
                b: p2,
                c: p3,
                d: p4,
                e: p5
            }
        """
        combination = dict()
        protocol = iter(protocols)
        for action in actions:
            combination[action] = next(protocol)
        return combination

    @classmethod
    def _action_protocols_combinations(cls, actions, protocols_combinations) -> List[Dict[Action, Protocol]]:
        """
        Generates a list of dictionaries of the form {Action: Protocol}, using the given
        actions and protocols combinations.

        :param actions: list of supported actions.
        :type actions: list containing Action objects.
        :param protocols_combinations: list of combinations of protocols.
        :type protocols_combinations: list containing tuples of Protocol objects.
        :return: list containing dicts of form {Action: Protocol}.

        :example:
        cls._action_protocols_combinations(actions=[a,b,c,d,e],
                                           protocols_combinations=[(p1,p2,p3,p4,p5), (p2,p4,p5,p3,p1)])
            [{a: p1, b: p2, c: p3, d: p4, e: p5},
             {a: p2, b: p4, c: p5, d: p3, e: p1}]
        """
        action_protocols_combinations = list()
        for protocol_combination in protocols_combinations:
            action_protocols_combinations.append(
                cls._action_protocols_combination(actions, protocol_combination)
            )
        return action_protocols_combinations

    def _client_name(self):
        # Shape of the action:protocols dictionary:
        # [{a: p1, b: p2, c: p3, d: p4, e: p5},
        #  {a: p2, b: p4, c: p5, d: p3, e: p1}]
        d = self.action_protocols_combinations[self._index]
        client_name = str()
        for a, p in d.items():
            client_name += str(a) + "::" + str(p) + ","
        return client_name[:-1]


class CartesianProductIterator(ProtocolIterator):
    @staticmethod
    def _protocols_combinations(actions, protocols) -> List[Tuple[Protocol, ...]]:
        """
        Generates a list of tuples which contains mixed protocols. Each tuple contains a number of protocols
        equal to the number of actions.

        :param actions: list of supported actions.
        :param protocols: list of supported protocols.
        :return: list of mixed protocols tuples.

        :example:
        _protocols_combinations(actions=[a,b,c,d,e], protocols=[Dbus, Http])
            [('Dbus', 'Dbus', 'Dbus', 'Dbus', 'Dbus'),
            ('Dbus', 'Dbus', 'Dbus', 'Dbus', 'Http'),
            ('Dbus', 'Dbus', 'Dbus', 'Http', 'Dbus'),
            ('Dbus', 'Dbus', 'Dbus', 'Http', 'Http'),
            ...]
        """
        from itertools import product
        n_actions = len(actions)
        return list(product(protocols, repeat=n_actions))


class Clients(Iterable):
    def __init__(self, actions: List[Action], protocols: List[Protocol]) -> None:
        """
        Collection of clients for issuing test operations.

        :param actions: List of actions that the protocols can perform.
        :param protocols: List of Protocol objects, already set up.
        """
        self.actions = actions
        self.protocols = protocols
        self._default_iterator = self.get_protocol_iterator

    def __iter__(self) -> CartesianProductIterator:
        """
        Returns a default iterator.
        """
        return self._default_iterator()

    def get_cartesian_product_iterator(self):
        """
        Returns the CartesianProductIterator iterator.
        """
        return CartesianProductIterator(self.actions, self.protocols)

    def get_protocol_iterator(self):
        return ProtocolIterator(self.actions, self.protocols)


def is_ready(req_timeout):
    ready_endpoint = "health"
    url = settings.SERVER_APP_BASE_URL + '/' + ready_endpoint
    try:
        res = requests.get(url=url, timeout=req_timeout)
        return res.status_code == 200
    except requests.exceptions.ConnectionError:
        return False
