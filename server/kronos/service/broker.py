import time
from uuid import uuid4
from confluent_kafka.admin import AdminClient, NewTopic
from confluent_kafka import Consumer
from confluent_kafka import Producer
import json
from loguru import logger
from kronos.utils import Singleton
from aiocache import Cache
import threading
from kronos.settings import settings
import paho.mqtt.client as mqtt
import asyncio


class KafkaBroker(object, metaclass=Singleton):

    def __init__(self, url=settings.KAFKA_BROKER):

        KAFKA_PARAMS = {'bootstrap.servers': url}  # noqa

        self._url = url
        self._admin_client = AdminClient(KAFKA_PARAMS)
        self._consumers = {}
        self._producer = Producer(KAFKA_PARAMS)
        self._logger = logger
        self._logger.debug(f"Build KafkaBroker with url {self._url}")

    def create_topics(self, new_topics,
                      num_partitions=3,
                      replication_factor=1):
        """

        Create topics on a Kafka broker.

        :param new_topics: topic or list of topic names.
        :param num_partitions: number of partition
        :param replication_factor: replication factor

        """
        if isinstance(new_topics, str):
            new_topics = [new_topics]

        new_topics = [NewTopic(topic, num_partitions=num_partitions, replication_factor=replication_factor) for topic in
                      new_topics]
        fs = self._admin_client.create_topics(new_topics)  # Delayed

        created_topics = {}
        # Wait for each operation to finish.
        for topic, f in fs.items():
            try:
                t = f.result()  # The result itself is None
                self._logger.debug("Topic {} created".format(topic))
                created_topics[topic] = t
            except Exception as e:
                self._logger.debug("Failed to create topic {}: {}".format(topic, e))
        return created_topics

    def get_consumer(self, topics, auto_offset_reset="earliest", group_id="factory"):
        """
        Get a consumer instance for this topics.
        :param topics: a list of topics for the current consumer.
        :param auto_offset_reset: the current offset.
        :param group_id: the group-id
        :return:
        """
        if isinstance(topics, str):
            topics = (topics,)
        topics = tuple(topics)

        if topics not in self._consumers:
            consumer = Consumer({
                'bootstrap.servers': self._url,
                'group.id': group_id,
                'auto.offset.reset': auto_offset_reset
            })
            to_create = [x for x in [settings.SERVICE_REPLIES, settings.SERVICE_ERRORS,
                                     settings.SERVICE_REQUESTS] if x not in consumer.list_topics().topics]
            if len(to_create) > 0:
                self.create_topics(
                    to_create
                )
            consumer.subscribe(list(topics))
            self._consumers[topics] = consumer
        return self._consumers[topics]

    def poll(self, topics, timeout=1.0):
        """
        Poll from specified topics.
        :param topics: A list of topics.
        :param timeout: the preferred timeout.
        :return: a list of messages.
        """
        consumer = self.get_consumer(topics)
        return consumer.poll(timeout)

    def send(self, topic, value, key=None):
        """
        Send key-value to specified topic.
        :param topic: Kafka Topic
        :param value: Message to send
        :param key: The kafka key
        :return: None
        """
        self._logger.debug(f"Sending new message to topic {topic}")

        def delivery_report(err, msg):
            if err is not None:
                self._logger.error(f'Event delivery failed: {err}')
            else:
                self._logger.debug(f'Event delivered to Kafka '
                                   f'(Topic: {msg.topic()}, Partition: {msg.partition()})')

        # Trigger any available delivery report callbacks from previous produce() calls
        self._producer.poll(0)

        # Asynchronously produce a message, the delivery report callback
        # will be triggered from poll() above, or flush() below, when the message has
        # been successfully delivered or failed permanently.
        self._producer.produce(topic, key=key, value=value, callback=delivery_report)

        # Wait for any outstanding messages to be delivered and delivery report
        # callbacks to be triggered.
        self._producer.flush()
        self._logger.debug(f"Message sent")


class ResponseResolver(threading.Thread):

    def __init__(self, broker: KafkaBroker, cache: Cache.REDIS):
        """
        :param broker:
        :param cache:
        """
        super(ResponseResolver, self).__init__()
        self._broker = broker
        self._cache = cache
        self._logger = logger

    def run(self) -> None:
        loop: asyncio.events.AbstractEventLoop = asyncio.new_event_loop()
        while True:
            message = self._broker.poll([settings.SERVICE_ERRORS, settings.SERVICE_REPLIES], settings.POLL_TIMEOUT)
            if message:
                try:
                    self._logger.info(f"Received message : {message}")
                    response = json.loads(message.value())
                    loop.run_until_complete(
                        self.save_results(key=response.get("caller_request_id", None), value=response))
                except Exception:
                    self._logger.error(f"Failed to parse message {message.value()}")

    async def save_results(self, key, value):
        """

        :param key:
        :param value:
        :return:
        """
        await self._cache.set(key, value)
        assert await self._cache.get(key) == value


class AsyncioHelper(metaclass=Singleton):
    def __init__(self, loop, client):
        self.loop = loop
        self.client = client
        self.client.on_socket_open = self.on_socket_open
        self.client.on_socket_close = self.on_socket_close
        self.client.on_socket_register_write = self.on_socket_register_write
        self.client.on_socket_unregister_write = self.on_socket_unregister_write

    def on_socket_open(self, client, userdata, sock):

        def cb():
            client.loop_read()

        self.loop.add_reader(sock, cb)
        self.misc = self.loop.create_task(self.misc_loop())

    def on_socket_close(self, client, userdata, sock):
        self.loop.remove_reader(sock)
        self.misc.cancel()

    def on_socket_register_write(self, client, userdata, sock):

        def cb():
            client.loop_write()

        self.loop.add_writer(sock, cb)

    def on_socket_unregister_write(self, client, userdata, sock):
        self.loop.remove_writer(sock)

    async def misc_loop(self):
        while self.client.loop_misc() == mqtt.MQTT_ERR_SUCCESS:
            try:
                await asyncio.sleep(1)
            except asyncio.CancelledError:
                break


class LoopHandler(metaclass=Singleton):
    def __init__(self, loop, topics):
        self.loop = loop
        self.topics = topics
        self.logger = logger
        self.broker = KafkaBroker()

    def on_connect(self, client, userdata, flags, rc):
        self.logger.debug(f"Subscribing: {userdata}, {flags}, {rc}")
        client.subscribe(self.topics)

    def on_message(self, client, userdata, msg):
        self.logger.info(f"Processing message {msg.payload} from topic {msg.topic}")
        try:
            payload = self._build_payload(topic=msg.topic, payload=msg.payload)  # noqa
            self.logger.debug(f"Payload is {payload}")
            self.broker.send(topic=settings.SERVICE_REQUESTS, value=payload)
        except Exception as e:
            self.logger.error(f"Unable to forward {msg.payload} to kafka. {e}")

    def on_disconnect(self, client, userdata, rc):
        self.logger.debug(f"Disconnecting: {userdata}, {rc}")
        self.disconnected.set_result(rc)

    def on_publish(self, client, userdata, mid):  # noqa
        self.logger.debug(f"Publishing using {userdata}")

    def _build_payload(self, topic, payload):  # noqa
        """
        Build payload for correct service. If topic == events -> return TransactionModel
        :param topic:
        :param payload:
        :return:
        """
        self.logger.info(f"Processing request from topic: {topic}")
        topic = f"/{topic.split('/')[-1]}"
        if topic == f"/{settings.SERVICE_TRANSACTION_TOPIC.split('/')[-1]}":
            self.logger.debug(f"Creating transaction model from {payload}")
            from kronos.sync.transaction.model import TransactionRequest, TransactionCommand
            return TransactionRequest(
                request_id=str(uuid4()),
                ts=int(time.time()),
                command=TransactionCommand.CRUD.value,
                request_type="TransactionApp",
                payload=json.loads(payload)
            ).dumps(serializer="json")
        elif topic in [f"/{settings.SERVICE_MQTT_CONNECTED_TELEMETRY_TOPIC.split('/')[-1]}",
                       f"/{settings.SERVICE_MQTT_DISCONNECTED_TELEMETRY_TOPIC.split('/')[-1]}"]:
            self.logger.debug(f"Creating connected model from {payload}")
            try:
                from kronos.sync.telemetry.model import TelemetryRequest, TelemetryCommand
                if topic == f"/{settings.SERVICE_MQTT_CONNECTED_TELEMETRY_TOPIC.split('/')[-1]}":
                    command = TelemetryCommand.CONNECTION.value
                else:
                    command = TelemetryCommand.DISCONNECTION.value
                return TelemetryRequest(
                    request_id=str(uuid4()),
                    ts=int(time.time()),
                    command=command,
                    payload=json.loads(payload),
                    request_type="TelemetryApp"
                ).dumps(serializer="json")
            except Exception as e:
                self.logger.error(e)
        elif topic == "command":
            pass
        raise Exception(f"Cannot build payload for current request. {topic}-{payload}")
