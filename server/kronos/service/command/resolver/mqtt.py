from . import Resolver
import asyncio
from queue import Queue
from kronos.settings import settings
import paho.mqtt.client as mqtt


class MqttResolver(Resolver):
    bus = Queue()

    def __init__(self, device_id):
        super(MqttResolver, self).__init__(device_id=device_id)
        self._request_topic = f"{settings.SERVICE_MQTT_COMMAND_BASE_TOPIC.format(device_id)}"
        self._response_topic = f"{self._request_topic}/{self.request_id}/response"
        self.mqtt_handler = mqtt.Client(client_id=self._client_id, clean_session=True)

        self.mqtt_handler.on_connect = lambda client, userdata, flags, rc: client.subscribe(self._response_topic)

        def on_message(client, userdata, msg):
            MqttResolver.bus.put(msg.payload)

        self.mqtt_handler.on_message = on_message

        self.mqtt_handler.connect(host=settings.MQTT_BROKER_HOST, port=settings.MQTT_BROKER_PORT)
        asyncio.get_event_loop().create_task(self.start_loop_task())  # Start consumer mqt

    async def start_loop_task(self):
        self.mqtt_handler.loop_start()

    def request(self, payload, *args, **kwargs):
        self._logger.info(f"Forward request {payload} to command topic {self._request_topic}.")
        publishing = self.mqtt_handler.publish(topic=self._request_topic, payload=payload,
                                               qos=2)
        self._logger.info(
            f"Publishing is {publishing} - {publishing.is_published()}"
        )
        while not publishing.is_published():
            self.mqtt_handler.loop(0.1)

    def response(self, *args, **kwargs):
        return MqttResolver.bus.get(block=True, timeout=settings.POLL_TIMEOUT)
