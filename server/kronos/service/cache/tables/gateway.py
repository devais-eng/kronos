import json
from datetime import datetime, timezone
from faust_utils.cache.tables import CacheTable
from loguru import logger
from kronos.settings import settings
from kronos.service.command import CommandBuilder, CommandType


class GatewayConnection(CacheTable):
    """
    Monitors the connection state of gateways. The table is structured like a dictionary
    Dict[Gw-mac, last connection]
    """

    _table_name = "gateway_connection"

    def __init__(self, *args, **kwargs):
        super(GatewayConnection, self).__init__(*args, **kwargs)
        self.add_event(event_type="Refresh", callback=self.refresh_callback,
                       interval=settings.INTERNAL_SERVICE_INTERVAL)

    def refresh_callback(self):
        """
        Recurring callback
        """
        logger.debug(f"Call API refresh_callback")
        for device_id, telemetry in self.items():
            now = int(datetime.now().replace(tzinfo=timezone.utc).timestamp()) * 1000
            device_local_time = telemetry.get('timestamp_local', 0)
            force_telemetry = now - device_local_time >= settings.MAX_TIME_TO_RESYNC_TELEMETRY

            logger.debug(
                f"Processing device with id {device_id} -> last connection {telemetry.get('timestamp_local')}")

            if force_telemetry or telemetry.get("connected", True) is False:
                logger.debug(f"Device with id {device_id} is disconnected.")
                command = CommandBuilder.build( # 1629902968
                    CommandType.GET_TELEMETRY.value
                )(device_id=device_id)
                if command is not None:
                    new_telemetry = json.loads(command.execute())
                    if new_telemetry.get("success", False) is True:
                        logger.info(f"New telemetry was {new_telemetry}")
                        self.update(device_id, json.dumps(new_telemetry.get("body", {})))

    def _on_startup(self):
        """
        Function to be applied at startup
        """
        super(GatewayConnection, self)._on_startup()
        logger.debug(f"Execute on_startup of {GatewayConnection.__name__}")
        for gateway_mac, last_connection in self.items():
            logger.info(f"Gateway with mac {gateway_mac} -> last connection {last_connection}")
