FROM tiangolo/uvicorn-gunicorn-fastapi:python3.7

WORKDIR /kronos/

RUN apt-get update && apt-get install -y \
    libxmlsec1-dev \
    pkg-config \
    wget \
    && rm -rf /var/lib/apt/lists/*

# Install Poetry
ENV PYTHONBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    POETRY_VERSION=1.3.0 \
    POETRY_VIRTUALENVS_IN_PROJECT=false \
    POETRY_NO_INTERACTION=1 \
    PIP_DEFAULT_TIMEOUT=500

RUN curl -sSL https://install.python-poetry.org | POETRY_HOME=/opt/poetry python && \
    ln -s /opt/poetry/bin/poetry /usr/local/bin/poetry && \
    poetry config virtualenvs.create false

# Copy only pyproject.toml and poetry.lock to leverage Docker layer caching
COPY ./pyproject.toml ./poetry.lock /kronos/

# TODO: change with ssh
ARG GITLAB_USERNAME
ARG GITLAB_PASSWORD
ENV POETRY_HTTP_BASIC_GITLAB_USERNAME=$GITLAB_USERNAME \
    POETRY_HTTP_BASIC_GITLAB_PASSWORD=$GITLAB_PASSWORD

# Allow installing dev dependencies to run tests
ARG INSTALL_DEV=false

RUN poetry install --no-root $(if [ $INSTALL_DEV = "true" ]; then echo "--no-dev"; fi)

ENV DOCKERIZE_VERSION v0.6.1
RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && tar -C /usr/local/bin -xzvf dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && rm dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz

RUN apt-get purge -y wget && apt-get autoremove -y && rm -rf /var/lib/apt/lists/*

COPY . /kronos/
COPY prestart.sh /app/
COPY kronos/main.py /app/
ENV PYTHONPATH=/kronos:/app
