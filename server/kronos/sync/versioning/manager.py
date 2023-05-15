from kronos.sync.versioning.memory import VersioningCache
from kronos.storage import SessionFactory
from kronos.sync.transaction.roles import Role


class VersionManager(object):
    """
    Transaction application calls role manager and applies transaction through it.
    """

    def __init__(self, app):
        self._faust_app = app
        self._db = SessionFactory()
        self._version_cache = VersioningCache(app=self._faust_app)

    def apply_transactions(self, transactions, role: Role):  # TODO: Apply transaction logic
        """
        Apply transactions and solve conflicts according role.
        """
        return self._version_cache.checkout(deltas=transactions, role=role, storage=self._db)
