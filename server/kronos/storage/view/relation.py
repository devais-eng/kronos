from pydantic import Field, validator, StrictStr
from typing import Optional, Type
from kronos.storage.view import View, EntityModel, EntityCreateModel, EntityInfoModel, EntityPatchModel


class RelationModel(EntityModel):
    """
     Base model for relations.
    """
    parent_id: StrictStr = Field(description='The ID of the parent item')
    child_id: StrictStr = Field(description='The ID of the child item')


class RelationCreateModel(EntityCreateModel, RelationModel):
    """
     Base model for creating relations.
    """

    id: Optional[StrictStr]


class RelationPatchModel(EntityPatchModel):
    """
     Base model for creating relations.
    """
    modified_by: Optional[StrictStr] = Field(None, description='Who modify this relations.')
    parent_id: Optional[StrictStr] = Field(description='The ID of the parent item')
    child_id: Optional[StrictStr] = Field(description='The ID of the child item')


class RelationInfoModel(EntityInfoModel, RelationModel):
    """
    Model fore relations information.
    """
    from kronos.storage.view.item import ItemInfoModel

    parent_id: StrictStr = Field(None, description="The parent item.")
    child_id: StrictStr = Field(None, description="The child item.")


class RelationView(View):
    """
    View for modeling relations.
    """

    @classmethod
    def base(cls):
        return RelationModel

    @classmethod
    def info(cls):
        return RelationInfoModel

    @classmethod
    def create(cls):
        return RelationCreateModel

    @classmethod
    def db_create(cls):
        return RelationCreateModel

    @classmethod
    def patch(cls):
        return RelationPatchModel
