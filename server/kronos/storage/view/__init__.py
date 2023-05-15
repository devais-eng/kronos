from typing import AnyStr, Optional, Type
from pydantic import BaseModel, Field, StrictStr
import time


class EntityModel(BaseModel):
    """
     This is the very base model, carrying all the information each entity should be always provided.
    """
    pass
    # version: Optional[StrictStr] = Field(None, description="The last version of this entity.")
    # id: str = Field(None, description="The unique identifier of this entity.")
    # entity_type: AnyStr = Field(None, description="The literal type of this entity.")


class EntityPatchModel(BaseModel):
    modified_by: Optional[StrictStr] = Field("SYNC", description="The last editor of this entity.")
    modified_at: Optional[int] = Field(int(time.time()), description="The timestamp when this entity was last updated.")


class EntityCreateModel(EntityModel):
    """
     The base model for any entity creation.
    """
    source_timestamp: int = Field(None, description="The timestamp when this item was generated at source.")
    created_by: StrictStr = Field("SYNC", description="The author of this entity.")
    # version_graph_id: StrictStr = Field(None, description="The version_graph_id")


class EntityUpdateModel(EntityModel):
    """
     The base model for any entity updates.
    """
    modified_by: StrictStr = Field("SYNC", description="The last editor of this entity.")
    modified_at: int = Field(int(time.time()), description="The timestamp when this entity was last updated.")


class EntityDeleteModel(BaseModel):
    """
     The base model for any entity delete.
    """
    id: str = Field(None, description="The unique identifier of this entity.")
    hard: bool = Field(False, description="Hard or soft delete.")


class EntityInfoModel(EntityModel):
    """
     The base model for the displayable information of an entity.
    """
    id: str = Field(None, description="The unique identifier of this entity.")
    created_by: str = Field(None, description="The author of this entity.")
    created_at: int = Field(None, description="The timestamp when this entity was actually crated into the storage.")
    modified_by: Optional[str] = Field(None, description="The last editor of this entity.")
    modified_at: Optional[int] = Field(None, description="The timestamp when this entity was last updated.")
    version: AnyStr = Field(None, description="The last synchronized version of this entity.")
    # Uncomment the field below to enable soft delete
    # active: Optional[bool] = Field(None, description="Whether this entity is enabled.")


class EntityQueryModel(EntityModel):
    """
     The base model for queries performable over entity.
    """
    pass


class EntityCount(BaseModel):
    count: int


class View(object):
    """
    Views are used to represent entities according to different needs.
    """

    @classmethod
    def base(cls) -> Type[EntityModel]:
        return EntityModel

    @classmethod
    def info(cls) -> Type[EntityInfoModel]:
        """
        Produces the model used to show basic information.
        """
        return EntityInfoModel

    @classmethod
    def create(cls) -> Type[EntityCreateModel]:
        """
        Produces the model used for creation.
        """
        return EntityCreateModel

    @classmethod
    def db_create(cls) -> Type[EntityCreateModel]:
        """
        Produces the model used db for creation.
        """
        return EntityCreateModel

    @classmethod
    def db_update(cls) -> Type[EntityCreateModel]:
        """
        Produces the model used db for update.
        """
        return EntityCreateModel

    @classmethod
    def update(cls) -> Type[EntityUpdateModel]:
        """
        Produces the model used for updates.
        """
        return EntityUpdateModel

    @classmethod
    def patch(cls) -> Type[EntityPatchModel]:
        return EntityPatchModel

    @classmethod
    def delete(cls) -> Type[EntityDeleteModel]:
        """
        Produces the model used for deletion.
        """
        return EntityDeleteModel

    @classmethod
    def query(cls) -> Type[EntityQueryModel]:
        """
        Produces the model (or set of possible models) used for querying the entity.
        """
        return EntityQueryModel

    @classmethod
    def count(cls) -> Type[EntityCount]:
        """
        Produces the model (or set of possible models) used for counts entities.
        """
        return EntityCount
