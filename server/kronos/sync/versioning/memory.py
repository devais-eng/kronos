import time

from kronos.transaction.delta import DeltaTransaction
from kronos.sync.versioning.model import VersionGraphResponse
from kronos.storage.entity import EntityType, get_by_type
from kronos.settings import settings
from fastapi.encoders import jsonable_encoder
from kronos.sync.transaction.conflicts import Conflict
from loguru import logger


class VersionGraphTable(object):
    """
    Version Graph Table
    """

    def get_version_graph(self, entity_type, entity_id, storage):  # noqa
        """
        Get or create a Version graph for (entity_type,entity_id).
        :param entity_type:
        :param entity_id:
        :return:
        """
        entity_type = get_by_type(entity_type)()  # noqa
        try:
            found = entity_type.read(entity_id=entity_id,
                                     raise_not_found=True)
            logger.debug(
                f"Entity {found} founded with graph {found.version_graph_id}")
            version_graph = found.version_graph.to_graph(found.version_graph_id)
            logger.warning(
                f"For tuple ({entity_id, entity_type} a version graph already exists. {version_graph._entity_id}")
            return version_graph
        except Exception:  # noqa
            from kronos.graph.storage.versioning import VersioningGraphEntity
            logger.warning(
                f"Impossible to find a version graph for {(entity_id, entity_type)}. Registering new graph.")
            graph_id = VersioningGraphEntity.new_graph()
            from kronos.graph.storage.versioning import VersioningGraphEntity
            return VersioningGraphEntity.to_graph(graph_id=graph_id)

    def update_version_graph(self, graph):  # noqa
        """

        :param entity_type:
        :param entity_id:
        :param graph:
        :param storage:
        :return:
        """
        new_deltas, new_master = graph.assign_version()
        return new_master, new_deltas


class VersioningCache(object):

    def __init__(self, *args, **kwargs):
        self._graphs = VersionGraphTable()

    def checkout(self, deltas, role, storage):  # noqa
        """
        When confirmed, commit deltas and updates versions.
        :param deltas:
        :param storage:
        :return:
        """
        graphs = {}
        to_update = {}

        revert = []
        for t in deltas._body:  # noqa
            entity_type = t._entity_type  # noqa
            entity_id = t._entity_id  # noqa
            entity_tuple = (entity_type, entity_id)
            if not entity_tuple in graphs:  # noqa
                graphs[entity_tuple] = self._graphs.get_version_graph(*entity_tuple, storage)
            if graphs[entity_tuple] is not None:
                current_version = graphs[entity_tuple].master
                entity_version = t._body.get("sync_version", None) if t._body else None  # noqa
                logger.debug(f"Graph master version is {current_version}")
                logger.debug(f"Entity version is {entity_version}")
                if t._body is None or entity_version is None or entity_version is not None and not current_version.__repr__() == entity_version:  # noqa
                    op = DeltaTransaction(body=[t])
                    graphs[entity_tuple].apply(op)
                    revert.extend(op.get_inverse())
                    to_update[entity_tuple] = True
        # deltas.commit()
        responses = []
        for entity_tuple, version_graph in graphs.items():
            logger.debug(f"Processing {entity_tuple}")

            if entity_tuple in to_update:
                entity_type = get_by_type(EntityType[entity_tuple[0]])()  # noqa
                # Get entity and refresh data with current master version
                result_entity = entity_type.read(entity_id=entity_tuple[1],
                                                 raise_not_found=False)  # noqa
                entity_id = entity_tuple[1]
                try:
                    master_version, _ = self._graphs.update_version_graph(graphs[entity_tuple])
                    logger.debug(f"Graph master version after update is {master_version}")
                except Conflict as c:
                    solved = c.solve(role)
                    if not solved:
                        for r in revert[::-1]:
                            r.apply()
                        from kronos.sync.transaction.conflicts.storage import ConflictEntity
                        ConflictEntity.new_conflict(graphs[entity_tuple]._entity_id, deltas, role, None, str(c))  # noqa
                        raise c
                    else:
                        result_entity = entity_type.read(entity_id=entity_tuple[1],
                                                         raise_not_found=False)  # noqa
                        master_version = result_entity.version
                        logger.debug(f"For entity of type {entity_type} with id {entity_id} conflict {c} was solved.")
                if result_entity is not None and result_entity.active:
                    entity_type.update(previous=result_entity,
                                       new={"version": master_version,
                                            "id": result_entity.id,
                                            # noqa
                                            "version_graph_id": graphs[entity_tuple]._entity_id})  # noqa
                    # Build VersionGraph responses with type,id and version.
                    logger.debug(f"Forwarding entity with id: {result_entity.id}")

                    responses.append(VersionGraphResponse(
                        entity_type=entity_tuple[0],
                        entity_id=result_entity.id,
                        timestamp=result_entity.modified_at,
                        version=master_version,  # noqa
                        action=settings.ACTION_ON_UPDATE if result_entity is not None else settings.ACTION_ON_CREATE,
                        payload={k: v for k, v in
                                 entity_type.view.base()(**jsonable_encoder(result_entity)).dict().items() if
                                 v is not None}
                    ))
                else:
                    # Entity was deleted. Send synchronization message with no payload.
                    from kronos.storage.entity.attribute import AttributeEntity
                    delete_payload = {}
                    if isinstance(entity_type, AttributeEntity):
                        delete_payload["item_id"] = result_entity.item_id
                    responses.append(VersionGraphResponse(
                        entity_type=entity_tuple[0],
                        entity_id=entity_id,
                        version=master_version,  # noqa
                        timestamp=result_entity.modified_at if result_entity else int(time.time()),
                        action=settings.ACTION_ON_DELETE,
                        payload=delete_payload
                    ))
        return responses
