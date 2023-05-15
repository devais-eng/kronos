from kronos.sync.telemetry.model import TelemetryRequest, TelemetryCommand
from kronos.service.cache.gateway import GatewayCache
from faust_utils.apps import ServiceApp
from faust_utils.model import Response, ResponseCode
from loguru import logger
import uuid
import time


class TelemetryApp(ServiceApp):

    def __init__(self, faust_app):
        super(TelemetryApp, self).__init__(faust_app=faust_app)
        self._gw_cache = GatewayCache(faust_app=faust_app)
        self._add_command(TelemetryCommand.CONNECTION, self.connection_req)
        self._add_command(TelemetryCommand.DISCONNECTION, self.disconnection_req)

    async def disconnection_req(self, telemetry_request: TelemetryRequest):
        device_id = telemetry_request.payload.device_id
        logger.info(
            f"Processing disconnection for device: {device_id}")  # noqa
        self._gw_cache.disconnect_gateway(device_id=device_id)
        return Response(caller_request_id=telemetry_request.request_id,
                        response_id=str(uuid.uuid1()),
                        response_code=ResponseCode.OK,
                        ts=time.time(),
                        execution_time=time.time() - telemetry_request.ts)

    async def connection_req(self, telemetry_request: TelemetryRequest):
        """
        pass
        """
        start_ts = time.time()
        device_id = telemetry_request.payload.device_id
        payload = telemetry_request.payload.telemetry
        logger.info(
            f"Processing telemetry for device: {device_id} and payload : {payload}")  # noqa
        self._gw_cache.register_gateway(device_id, payload)
        logger.debug(
            f"Cache contents is {self._gw_cache.get_gateways_list()}")
        return Response(caller_request_id=telemetry_request.request_id,
                        response_id=str(uuid.uuid1()),
                        response_code=ResponseCode.OK,
                        ts=time.time(),
                        execution_time=time.time() - start_ts)
