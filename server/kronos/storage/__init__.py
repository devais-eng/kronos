from typing import Generator
from sqlalchemy.orm import Session
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker, scoped_session
from sqlalchemy_utils import database_exists, create_database
from tenacity import after_log, before_log, retry, stop_after_attempt, wait_fixed
import logging
from kronos.settings import settings
# Abstract entities
from kronos.storage.entity import BaseEntity
from kronos.storage.entity import Entity  # noqa
# Kronos Entities
from kronos.graph.storage import GraphEntity  # noqa
from kronos.graph.storage.edge import EdgeEntity  # noqa
from kronos.graph.storage.versioning import VersioningGraphEntity  # noqa
from kronos.graph.storage.versioning.version import VersionNodeEntity  # noqa
from kronos.graph.storage.versioning.changelog import ChangelogEntity  # noqa
# Versioning Entities
from kronos.storage.entity.item import ItemEntity  # noqa
from kronos.storage.entity.relation import RelationEntity  # noqa
from kronos.storage.entity.attribute import AttributeEntity  # noqa
# Conflict
from kronos.sync.transaction.conflicts.storage import ConflictEntity  # noqa
from loguru import logger


def init_db(db: Session) -> None:
    """
    Initialize the database for a given session.
    """
    engine = db.get_bind()  # noqa
    if not database_exists(engine.url):
        create_database(engine.url)
    BaseEntity.metadata.create_all(bind=engine)


engine = create_engine(settings.SQLALCHEMY_DATABASE_URI, pool_pre_ping=True)  # Instantiate an engine based on settings
SessionFactory = scoped_session(
    sessionmaker(autocommit=False, autoflush=False, bind=engine, expire_on_commit=True))  # Used to create new sessions


def get_db() -> Generator:
    """
    Get an instance of sqlalchemy's session.
    :return:
    """
    try:
        db = SessionFactory()
        yield db
    finally:
        db.close()


# FOLLOWS PRE-START CODE TO INITIALIZE ENVIRONMENT AUTOMATICALLY ### -->


max_tries = 60 * 5  # 5 minutes
wait_seconds = 1


@retry(
    stop=stop_after_attempt(max_tries),
    wait=wait_fixed(wait_seconds),
    before=before_log(logger, logging.INFO),
    after=after_log(logger, logging.WARN),
)
def init() -> None:
    try:
        # Try to create session to check if DB is awake
        db = SessionFactory()
        engine = db.get_bind()
        if not database_exists(engine.url):
            create_database(engine.url)
        init_db(db)
    except Exception as e:
        logger.error(e)
        raise e


def main():
    logger.info("Initializing service")
    init()
    logger.info("Service finished initializing")


if __name__ == "__main__":
    main()

# END OF PRE-START ### <--
