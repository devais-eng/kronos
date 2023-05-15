from kronos.settings import settings
import requests


class Singleton(type):
    """
    It is used as a metaclass. Allows to build singleton instances of the decorated class.

    Example:

        class A(..., metaclass=Singleton):
            # Only a single instance of class A will be generated and shared.
            ...

    """

    _instances = {}

    def __call__(cls, *args, **kwargs):
        if cls not in cls._instances:
            cls._instances[cls] = super(Singleton, cls).__call__(*args, **kwargs)
        return cls._instances[cls]


def get_resource_id(entity):  # noqa
    split_id = entity.get("id", None).split(f"{settings.ID_SEPARATOR}")
    item_id = split_id[0].strip()
    object_id = split_id[1].strip()
    object_instance_id = split_id[2].strip()
    resource_id = split_id[3].strip()
    resource_instance_id = split_id[4].strip()
    return item_id, object_id, object_instance_id, resource_id, resource_instance_id


def get_object_id(entity):  # noqa
    split_id = entity.get("id", None).split(f"{settings.ID_SEPARATOR}")
    item_id = split_id[0].strip() if len(split_id) > 0 else None
    object_id = split_id[1].strip() if len(split_id) > 0 else None
    object_instance_id = split_id[2].strip() if len(split_id) > 1 else None
    return item_id, object_id, object_instance_id
