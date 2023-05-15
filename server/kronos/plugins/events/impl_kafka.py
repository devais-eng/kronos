import json
from loguru import logger
from typing import List, Dict

import pluggy

from kronos.settings import settings
from kronos.service.broker import KafkaBroker

hookimpl = pluggy.HookimplMarker(settings.FAUST_APP_NAME)  # noqa


class EventsPluginImplKafka(object):
    def __init__(self):

        servers = settings.KAFKA_BROKER

        self.broker = KafkaBroker(url=servers)
        self.topic = settings.KAFKA_SYNC_TOPIC

        logger.info(f'Kafka events plugin connected to {servers}')

    @hookimpl
    def publish_event(self,
                      payload: List[Dict]):
        try:
            data_enc = 'utf-8'
            for value in payload:
                logger.info(f"Processing value {value}")
                data = json.dumps(value).encode(data_enc)
                self.broker.send(
                    topic=self.topic,
                    value=data
                )
        except Exception as ex:
            logger.error(f'Failed to publish event ({payload}) to Kafka: {ex}')
