# Kronos

## Overview

This repository contains the code for the Kronos application. Kronos is an application that consists of two main components:

1. Edge side application
2. Server application

These components work together to provide functionality for creating, updating, and deleting entities through HTTP APIs.

## Usage with Docker

To build the application, simply run `docker-compose build`. To launch it, use the command `docker-compose up`.

For testing purposes, it is recommended to follow these steps between subsequent sessions:

1. Run `docker-compose down -v` to stop and remove the containers, as well as any associated volumes.
2. Run `docker-compose build kronos-backend` to rebuild the backend container.
3. Finally, run `docker-compose up` to start the application again.

By following these steps, you can ensure a clean environment for testing while utilizing the latest changes in the code.

Feel free to explore the code and make any necessary adjustments to suit your specific requirements.