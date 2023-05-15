import logging
from typing import List, Dict

import pluggy
from sentry_sdk import capture_exception

from kronos.settings import settings
from .events.impl_kafka import EventsPluginImplKafka
from .events.impl_mqtt import EventsPluginImplMqtt
from .events.spec import EventsPluginSpec

pm: pluggy.PluginManager


def init_app():
    global pm

    pm = pluggy.PluginManager(settings.FAUST_APP_NAME)
    pm.add_hookspecs(EventsPluginSpec)

    if settings.KAFKA_EVENTS_PLUGIN_ENABLED:
        pm.register(EventsPluginImplKafka())

    if settings.MQTT_EVENTS_PLUGIN_ENABLED:
        pm.register(EventsPluginImplMqtt())


def publish_event(payload: List[Dict]):
    if settings.PLUGINS_ENABLED:
        try:
            pm.hook.publish_event(
                payload=payload
            )
        except Exception as ex:
            capture_exception(ex)
            logging.exception(f'Failed to publish event')
