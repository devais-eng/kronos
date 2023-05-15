from pydantic import BaseSettings, Field, validator
from typing import Any, AnyStr, Dict, Optional, List, Union


class Settings(BaseSettings):
    # Server API settings
    SERVER_APP_PROTOCOL: Optional[str] = Field("http", description="Protocol used for communicating with the API.")
    SERVER_APP_HOSTNAME: Optional[str] = Field("localhost", description="Hostname of the API to be tested.")
    SERVER_APP_PORT: Optional[int] = Field(8081, description="Port of the API to be tested.")
    SERVER_APP_API_PATH: Optional[str] = Field("api/v1", description="Path of the API to be tested.")
    SERVER_APP_BASE_URL: Optional[str] = Field(None, description="Base URL of the API to be tested.")

    @validator("SERVER_APP_BASE_URL", pre=True)
    def assemble_server_base_url(cls, v: Optional[str], values: Dict[AnyStr, Any]) -> str:
        if isinstance(v, str):
            return v

        url = "%s://%s:%s" % (
            values.get("SERVER_APP_PROTOCOL"),
            values.get("SERVER_APP_HOSTNAME"),
            values.get("SERVER_APP_PORT")
        )

        if isinstance(values.get("SERVER_APP_API_PATH"), str):
            url += "/%s" % values.get("SERVER_APP_API_PATH")

        return url

    # Gateway API settings
    GATEWAY_APP_PROTOCOL: Optional[str] = Field("http", description="Protocol used for communicating with the API.")
    GATEWAY_APP_HOSTNAME: Optional[str] = Field("localhost", description="Hostname of the API to be tested.")
    GATEWAY_APP_PORT: Optional[int] = Field(5000, description="Port of the API to be tested.")
    GATEWAY_APP_API_PATH: Optional[str] = Field(None, description="Path of the API to be tested.")
    GATEWAY_APP_BASE_URL: Optional[str] = Field(None, description="Base URL of the API to be tested.")

    @validator("GATEWAY_APP_BASE_URL", pre=True)
    def assemble_gateway_base_url(cls, v: Optional[str], values: Dict[AnyStr, Any]) -> str:
        if isinstance(v, str):
            return v

        url = "%s://%s:%s" % (
            values.get("GATEWAY_APP_PROTOCOL"),
            values.get("GATEWAY_APP_HOSTNAME"),
            values.get("GATEWAY_APP_PORT")
        )

        if isinstance(values.get("GATEWAY_APP_API_PATH"), str):
            url += "/%s" % values.get("GATEWAY_APP_API_PATH")

        return url

    # Docker configurations
    DOCKER_COMPOSE_FILE_PATH: Optional[str] = None
    DOCKER_COMPOSE_PROJECT_NAME: Optional[str] = "kronos-test-stack" # noqa
    DOCKER_COMPOSE_CLEAN_UP_POLICY: Optional[str] = "down -v --remove-orphans"

    class Config:
        env_prefix = 'TEST_'  # Prefix of environment variables
        env_file = '.env'
        env_file_encoding = 'utf-8'


settings = Settings()
