from kronos.settings import settings
from faust_utils.model import ResponseCode, ErrorResponse, ErrorCode, Response
from kronos.service.broker import LoopHandler
from faust_utils.utils import get_app
from loguru import logger
import uuid
import time
import traceback
import paho.mqtt.client as mqtt
import asyncio
import socket
import json
from kronos.plugins import publish_event

app = get_app(f"{settings.FAUST_APP_NAME}", f"kafka://{settings.KAFKA_BROKER}")  # Main Faust Application
# Communication Channels - Kafka Topics
output_topic = app.topic(settings.SERVICE_REPLIES, value_type=Response)
error_topic = app.topic(settings.SERVICE_ERRORS, value_type=ErrorResponse)

# MQTT Client and loop preamble
client_id = f"{settings.MQTT_CLIENT_ID}-{uuid.uuid4()}"
topics = [(t.split(",")[0], int(t.split(",")[1])) for t in settings.MQTT_TOPICS.split(";")]
logger.info(f"Topics are {topics}")
loop = asyncio.get_event_loop()
loop_handler = LoopHandler(loop=loop, topics=topics)

logger.info(f"Creating loop handler.")
loop_handler.disconnected = loop.create_future()
loop_handler.client = mqtt.Client(client_id=client_id, clean_session=True)
loop_handler.client.on_connect = loop_handler.on_connect
loop_handler.client.on_message = loop_handler.on_message
loop_handler.client.on_disconnect = loop_handler.on_disconnect
loop_handler.client.subscribe(topics)

loop_handler.client.connect(host=settings.MQTT_BROKER_HOST, port=settings.MQTT_BROKER_PORT)
loop_handler.client.socket().setsockopt(socket.SOL_SOCKET, socket.SO_SNDBUF, 2048)
logger.info(f"Mqtt client correctly configured.")


@app.agent(output_topic, concurrency=settings.AGENT_CONCURRENCY)
async def process_response(responses):
    """
    Handles external response received on the output topic.

    :param responses: A response to be processed.
    """
    async for r in responses:
        logger.debug(f"Processing response with value {r}")
        try:
            if r.response_type is not None:
                if "Transaction" in r.response_type:
                    logger.debug(f"Forward response {r} to sync topic.")
                    publish_event(r.response_body)
        except Exception as e:
            traceback_str = traceback.format_tb(e.__traceback__)
            logger.debug(f"Unexpected error occurred: {str(e)}\n\n{traceback_str}")
            error_message = ErrorResponse(caller_request_id=r.response_id,
                                          response_id=str(uuid.uuid4()),
                                          ts=time.time(),
                                          response_code=ResponseCode.ERROR,
                                          execution_time=0,
                                          error_code=ErrorCode.SERVER_ERROR,
                                          error_message=f"An error occurred while loading request {r.response_id}.",
                                          traceback=traceback_str)
            await error_topic.send(value=error_message)


async def start_loop_task():
    logger.info("Starting mqtt loop")
    loop_handler.client.loop_start()


asyncio.get_event_loop().create_task(start_loop_task())  # Start consumer mqtt

if settings.PLUGINS_ENABLED:
    from kronos.plugins import init_app

    init_app()

app.main()  # Launch the main application
