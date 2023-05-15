from typing import Optional, AnyStr, Dict, Any, List, Union
from pydantic import BaseSettings, Field, validator, AnyHttpUrl


class Settings(BaseSettings):
    """
    Settings are configurations stored as environment variables. If the envvar is not set, a default may be used instead.
    """

    # Storage
    DB_SERVER: str = Field("localhost", description="URL for database server.")
    DB_USER: str = Field("root", description="The root user of the database.")
    DB_PASSWORD: str = Field("kronos-admin", description="The database root password.")
    DB_NAME: str = Field("kronos-db", description="Name of the database")
    DB_ENGINE: str = Field("mysql+pymysql", description="Engine to be used by the storage.")
    INIT_DB: Optional[bool] = Field(True, description="For development only, whether to initialize the database.")

    # REST
    BACKEND_CORS_ORIGINS: List[AnyStr] = Field(None, description="Backed cors origin")
    SENTRY_DSN: Optional[AnyHttpUrl] = None
    PROMETHEUS_ENABLED: Optional[bool] = Field(True, description="Enable prometheus")
    ENVDUMP_ENABLED: Optional[bool] = True

    SQLALCHEMY_DATABASE_URI: str = None

    @validator("SQLALCHEMY_DATABASE_URI", pre=True)
    def assemble_db_connection(cls, v: Optional[str], values: Dict[AnyStr, Any]) -> str:
        if isinstance(v, str):
            return v
        return f"{values.get('DB_ENGINE')}://{values.get('DB_USER')}:{values.get('DB_PASSWORD')}@{values.get('DB_SERVER')}/{values.get('DB_NAME')}"

    # SERVICE KAFKA MQTT
    KAFKA_BROKER: Optional[str] = Field("localhost:9092", description="Kafka broker address.")
    MQTT_CLIENT_ID: Optional[str] = Field("MQTTClientId", description="MQTT Client ID.")
    MQTT_BROKER_HOST: Optional[str] = Field("localhost", description="MQTT broker host address.")
    MQTT_BROKER_PORT: Optional[int] = Field(1883, description="MQTT broker port.")
    MQTT_SYNC_TOPIC: Optional[str] = Field("/sync", description="The default topic for synchronization message.")
    KAFKA_SYNC_TOPIC: Optional[str] = "sync"
    MQTT_POLLING_TIMER: Optional[int] = Field(2, description="The value used in faust agent to poll from mqtt.")
    MQTT_EVENTS_PLUGIN_USERNAME: Optional[str] = None
    MQTT_EVENTS_PLUGIN_PASSWORD: Optional[str] = None

    SERVICE_REPLIES: Optional[str] = Field("ServiceReply", description="This is the topic used for incoming request.")
    SERVICE_REQUESTS: Optional[str] = Field("ServiceRequest", description="This is the topic used for response.")
    SERVICE_ERRORS: Optional[str] = Field("ServiceError", description="This is the topic used for error response.")
    POLL_TIMEOUT: Optional[int] = Field(3, description="The timeout to poll a message from kafka.")
    FAUST_APP_NAME: Optional[str] = Field("ServiceManager", description="Name for current faust application.")
    FAUST_VERSION: Optional[int] = Field(1, description="Current application version for faust backend.")
    TOPIC_PARTITIONS: Optional[int] = Field(1, description="Preferred number of topic's partitions, default = 1.")
    PROCESSING_GUARANTEE: Optional[str] = Field("at_least_once", description="The processing guarantee strategy.")
    FAUST_STORAGE: Optional[str] = Field(None, description="The storage type for faust table (rocksdb, redis)")
    SERVICE_INTERNAL_TABLE_TYPE: Optional[str] = Field("CacheTable",
                                                       description="The type of table used by Faust to store version graph.")  # noqa

    SERVICE_TRANSACTION_TOPIC: Optional[str] = Field("/events", description="Topic name of transaction's agent.")

    SERVICE_MQTT_CONNECTED_TELEMETRY_TOPIC: Optional[str] = Field("$share/group/kronos/device/+/connected",  # noqa
                                                                  description="Topic name of telemetry's agent.")
    SERVICE_MQTT_DISCONNECTED_TELEMETRY_TOPIC: Optional[str] = Field("$share/group/kronos/device/+/disconnected",
                                                                     # noqa
                                                                     description="Topic for disconnection's telemetry message.")  # noqa
    MAX_TIME_TO_RESYNC_TELEMETRY: Optional[int] = Field(60000,
                                                        description="This is the max time to wait for a telemetry request.")  # noqa
    SERVICE_MQTT_COMMAND_BASE_TOPIC: Optional[str] = Field("/kronos/device/{}/commands",
                                                           description="This topic is used to submit command to edge service.")  # noqa
    MQTT_TOPICS: str = Field("/events,1;/connected,1", description="Topics.")
    LOG_NAME: str = Field("kronos.log", description="The filename of log.")

    @validator("MQTT_TOPICS", pre=True)
    def assemble_mqtt_topics(cls, v: Optional[str], values: Dict[AnyStr, Any]) -> str:
        return f"{values.get('SERVICE_TRANSACTION_TOPIC')},2;{values.get('SERVICE_MQTT_CONNECTED_TELEMETRY_TOPIC')},2;{values.get('SERVICE_MQTT_DISCONNECTED_TELEMETRY_TOPIC')},2"

    from multiprocessing import cpu_count
    AGENT_CONCURRENCY: Optional[int] = Field(cpu_count(), description="Concurrency level for agent.")
    ACTION_ON_UPDATE: Optional[str] = Field("UPDATE",
                                            description="Tagging creation/update action used in synchronization "
                                                        "message.")
    ACTION_ON_CREATE: Optional[str] = Field("CREATE",
                                            description="Tagging creation/update action used in synchronization "
                                                        "message.")
    ACTION_ON_DELETE: Optional[str] = Field("DELETE",
                                            description="Tagging delete action used in synchronization "
                                                        "message.")
    MQTT_OUTPUT_QOS: Optional[int] = Field(1, description="The quality of service for output topic.")
    REDIS_HOSTNAME: Optional[str] = Field("kronos-redis", description="The hostname of redis")
    REDIS_PORT: Optional[int] = Field(6379, description="The port of redis.")
    REDIS_NAMESPACE: Optional[str] = Field("kronos", description="The redis namespace.")
    INITIAL_VERSION: Optional[str] = Field("Origin", description="The initial tag for version.")

    # GLOBAL
    LOG_LEVEL: Optional[str] = Field("INFO", description="Log level (DEBUG|INFO|WARNING|ERROR|CRITICAL)")
    LOG_FORMAT: Optional[str] = Field("BASE",
                                      description="Log format.")
    ID_SEPARATOR: Optional[str] = Field("/", description="The separator used in api")
    PROJECT_NAME: Optional[str] = Field("Kronos-Server", description="FastAPI project name.")  # noqa
    API_V1_STR: Optional[str] = Field("/api/v1", description="FastAPI version.")
    DEFAULT_RES_INSTANCE_ID: Optional[str] = Field("0", description="Default value for instance id in resource.")
    FORCE_ENTITY_ID: Optional[bool] = Field(False, description="Flag to force auto creation of entity id.")
    INTERNAL_SERVICE_INTERVAL: Optional[Union[int, str]] = Field(5,
                                                                 description="Interval for internal cache callback.")
    PLUGINS_ENABLED: bool = True
    KAFKA_EVENTS_PLUGIN_ENABLED: bool = True
    MQTT_EVENTS_PLUGIN_ENABLED: bool = True
    MQTT_EVENTS_PLUGIN_KEEPALIVE: int = 60

    # Retry
    EXPO_BACKOFF_MIN: int = 1000
    EXPO_BACKOFF_MAX: int = 10000
    FIXED_MAX_TRIES: int = 10
    FIXED_WAIT_SECONDS: float = 0.3

    @validator("API_V1_STR", pre=True)
    def validate_api_version(cls, v: str) -> Optional[str]:
        if not v or len(v) == 0:
            return "/api/v1"
        if not v.startswith("/"):
            raise ValueError("Api version must start with /")
        return v


settings = Settings()
