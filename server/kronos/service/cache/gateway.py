from faust_utils.cache import Cache
from .tables.gateway import GatewayConnection
from loguru import logger
from typing import AnyStr, Dict
from kronos.exception import handle_exception
import json


class GatewayCache(Cache):
    """
    The GatewayCache allows to cache gateways status.
    """
    _cache_name = "cache_gateway"

    def __init__(self, *args, **kwargs):
        super(GatewayCache, self).__init__(*args, **kwargs)
        self._gateway_connections = GatewayConnection(self._app)

    @handle_exception("register_gateway")
    def register_gateway(self, device_id: AnyStr, payload: Dict):
        """
        Register a gateway into the cache.

        :param device_id: Define the device id
        :param payload: A dictionary of current device's telemetry
        """
        self._gateway_connections.update(device_id, json.dumps(payload).encode())

    @handle_exception("disconnect_gateway")
    def disconnect_gateway(self, device_id: AnyStr):
        """
        Disconnect a gateway.

        :param device_id: The gateway's ID.
        """
        previous_node = json.loads(self._gateway_connections.get(device_id))
        logger.warning(f"Type of node is {previous_node}")
        if previous_node:
            previous_node["connected"] = False
            self._gateway_connections.update(device_id, json.dumps(previous_node).encode())
        else:
            logger.warning(f"Gateway with client id {device_id} not found.")

    @handle_exception("get_gateway")
    def get_gateway(self, client_id):
        """
        Check for a gateway with given client ID.

        :param client_id: The client ID which queried gateway should be linked.
        :return: The gateway associated to client ID or None.
        """
        return self._gateway_connections.get(client_id)

    @handle_exception("get_gateways_list")
    def get_gateways_list(self):
        """
        Return all gateways telemetry available in the cache.

        :return: A list of dictionaries describing available gateways.
        """
        return [self._gateway_connections.get(gw) for gw in self._gateway_connections.list_elements()]
