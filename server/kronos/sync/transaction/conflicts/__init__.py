from kronos.sync.transaction.roles import Role
from kronos.transaction import Transaction
from loguru import logger


class Conflict(Exception):

    def __init__(self, error_message="A conflict was raised."):
        self._logger = logger
        super(Conflict, self).__init__(error_message)

    def solve(self, role: Role) -> Transaction:
        """
        Solve conflict according given roles.
        """
        if role == Role.FORCE:
            return True
        else:
            return False


class ReadSyncEntityNotExists(Conflict):

    def __init__(self, entity_type, entity_id):
        super(ReadSyncEntityNotExists, self).__init__(
            f"An entity of type {entity_type} with ID {entity_id} was expected but not found.")

    def solve(self, role: Role) -> Transaction:
        """
        Solve conflict according given roles.
        """
        return False


class ReadSyncMismatch(Conflict):

    def __init__(self, entity_type, entity_id, expected, found):
        super(ReadSyncMismatch, self).__init__(
            f"An entity of type {entity_type} with ID {entity_id} was expected to have body {expected} but has {found} instead.")

    def solve(self, role: Role) -> Transaction:
        """
        Solve conflict according given roles.
        """
        if role == Role.FORCE:
            return True
        else:
            return False


class CreateOnExistingEntity(Conflict):

    def __init__(self, entity_type, entity_id, new, found):
        self._entity_type = entity_type
        self._entity_id = entity_id
        self._new = new
        self._found = found
        super(CreateOnExistingEntity, self).__init__(
            f"An entity of type {entity_type} with ID {entity_id} can't be created with body {new} "
            f"because it already exists with body {found}.")

    def solve(self, role: Role) -> Transaction:
        """
        Solve conflict according given roles.
        """
        if role == Role.FORCE:
            from kronos.storage.entity import get_by_type, EntityType
            entity = get_by_type(EntityType[self._entity_type])()  # noqa
            entity.update(self._new, self._found, close=True) # TODO: Fix update with attributes
            return True
        else:
            return False


class EntityAlreadyDeleted(Conflict):

    def __init__(self, entity_type, entity_id):
        super(EntityAlreadyDeleted, self).__init__(
            f"An entity of type {entity_type} with ID {entity_id} was already deleted.")

    def solve(self, role: Role) -> Transaction:
        """
        Solve conflict according given roles.
        """
        return True


class NoChangeUpdate(Conflict):

    def __init__(self, entity_type, entity_id, body):
        super(NoChangeUpdate, self).__init__(
            f"No change detected for updated of {entity_type} with ID {entity_id} having same body {body}.")

    def solve(self, role: Role) -> Transaction:
        """
        Solve conflict according given roles.
        """
        return True


class VersionNotFound(Conflict):

    def __init__(self, version):
        super(VersionNotFound, self).__init__(f"Cannot checkout version {version} because it is not found.")

    def solve(self, role: Role) -> Transaction:
        """
        Solve conflict according given roles.
        """
        # TODO: Restore a compatible master version
        return True
