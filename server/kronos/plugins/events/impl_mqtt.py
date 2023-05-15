import json
from typing import Dict, List
from uuid import uuid4
from loguru import logger
import paho.mqtt.client as mqtt
import pluggy

from kronos.settings import settings

hookimpl = pluggy.HookimplMarker(settings.FAUST_APP_NAME)  # noqa


class EventsPluginImplMqtt(object):
    def __init__(self):

        self.topic = settings.MQTT_SYNC_TOPIC
        self.client_id = f'{settings.FAUST_APP_NAME}-{uuid4()}'
        self.client = mqtt.Client(self.client_id, clean_session=True)
        self.client.user_data_set(self)
        self.client.enable_logger(logger)
        self.client.on_connect = lambda c, u, flags, rc: self.on_connect(flags, rc)
        self.client.on_disconnect = lambda c, u, rc: self.on_disconnect(rc)
        self.client.on_publish = lambda client, userdata, mid: self.on_publish(client, userdata, mid)

        self.client.username_pw_set(
            settings.MQTT_EVENTS_PLUGIN_USERNAME,
            settings.MQTT_EVENTS_PLUGIN_PASSWORD,
        )

        self.client.connect_async(settings.MQTT_BROKER_HOST,
                                  settings.MQTT_BROKER_PORT,
                                  settings.MQTT_EVENTS_PLUGIN_KEEPALIVE)

        self.client.loop_start()

    @hookimpl
    def publish_event(self,
                      payload: List[Dict]):
        try:
            data_enc = 'utf-8'
            logger.info(f"Producing payload {payload} using MQTT plugin")
            data = json.dumps(payload).encode(data_enc)
            self.client.publish(topic=self.topic, payload=data, qos=settings.MQTT_OUTPUT_QOS)
            self.client.loop_stop()
        except Exception as ex:
            logger.error(f'Failed to publish event ({payload}) to MQTT: {ex}')

    def on_connect(self, flags: Dict, rc: int):
        if rc == 0:
            logger.info(f'MQTT events client {self.client_id} connected')
        else:
            logger.error(f'MQTT events client {self.client_id} failed to connect with code: {rc}')

    def on_disconnect(self, rc: int):
        logger.info(f'MQTT events client {self.client_id} disconnected with code {rc}')

    def on_publish(self, client, userdata, mid): # noqa
        logger.debug(f"MQTT events client {self.client_id} is publishing.")
