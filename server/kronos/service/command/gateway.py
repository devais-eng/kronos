import json

from . import Command


class GetTelemetry(Command):
    """
    This command processes requests of telemetry.
    """

    def __init__(self, device_id):
        super(GetTelemetry, self).__init__(device_id=device_id)

    def _get_payload(self):
        return json.dumps(
            {
                "command_type": "GET_TELEMETRY",
                "uuid": self._resolver.request_id
            }
        )
