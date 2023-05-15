from typing import Optional, List

from kronos.storage.view import View, EntityModel, EntityCreateModel, EntityUpdateModel, EntityInfoModel, \
    EntityPatchModel
from kronos.storage.view.attribute import NestedAttributeCreateModel, AttributeOutputModel
from pydantic import Field, validator, StrictStr


class ItemBaseModel(EntityModel):
    """
    Base model containing mandatory information for items.
    """
    name: StrictStr = Field(None, description='The unique name of this item.')
    type: Optional[StrictStr] = Field(description='The literal type of this item.')
    customer_id: Optional[StrictStr] = Field(None, description='The customer owning this item.')
    sync_policy: Optional[StrictStr] = Field(None, description='Policy level for synchronization.')
    edge_mac: Optional[StrictStr] = Field(None, description='If given, it represent this device\'s edge mac.')
    # Attributes: Optional[List[AttributeModel]] = Field([], description='Optional list of item Attributes.')


class ItemModel(EntityModel):
    id: StrictStr = Field(None, description="The unique identifier of this entity.")
    name: StrictStr = Field(None, description='The unique name of this item.')
    type: Optional[StrictStr] = Field(description='The literal type of this item.')
    customer_id: Optional[StrictStr] = Field(None, description='The customer owning this item.')
    sync_policy: Optional[StrictStr] = Field(None, description='Policy level for synchronization.')
    edge_mac: Optional[StrictStr] = Field(None, description='If given, it represent this device\'s edge mac.')
    attributes: Optional[List[AttributeOutputModel]] = Field([], description='Optional list of item attributes.')


class ItemCreateModel(EntityCreateModel, ItemModel):
    """
    Model for an item creation.
    """
    id: StrictStr = Field(None, description="The unique identifier of this entity.")
    attributes: Optional[List[NestedAttributeCreateModel]] = Field([], description='Optional list of item attributes.')


class ItemDbCreateModel(EntityCreateModel, ItemModel):
    """
    Model for an item creation.
    """
    attributes: Optional[List[NestedAttributeCreateModel]] = Field([], description='Optional list of item attributes.')

    @validator("attributes", pre=False)
    def evaluate_entities(cls, v, values, **kwargs):  # noqa
        from kronos.storage.entity.attribute import AttributeEntity
        if isinstance(v, List):
            result = []
            for value in v:
                value_dict = {k: v for k, v in value.dict().items() if k != "resources" and k != "instance_id"}
                result.append(AttributeEntity(**value_dict, item_id=values["id"]))
            return result
        return values


class ItemUpdateModel(EntityUpdateModel, ItemBaseModel):
    """
    Model for an item update.
    """
    pass


class ItemPatchModel(EntityPatchModel):
    name: Optional[StrictStr] = Field(None, description='The unique name of this item.')
    type: Optional[StrictStr] = Field(description='The literal type of this item.')
    customer_id: Optional[StrictStr] = Field(None, description='The customer owning this item.')
    sync_policy: Optional[StrictStr] = Field(None, description='Policy level for synchronization.')
    edge_mac: Optional[StrictStr] = Field(None, description='If given, it represent this device\'s edge mac.')


class ItemInfoModel(EntityInfoModel, ItemModel):
    """
    Model for item information.
    """
    from kronos.storage.view.attribute import AttributeInfoModel
    attributes: Optional[List[AttributeInfoModel]] = Field([], description="The list of item Attributes.")

    class Config:
        orm_mode = True


class ItemView(View):
    """
    View for modeling Items.
    """

    @classmethod
    def base(cls):
        return ItemModel

    @classmethod
    def info(cls):
        return ItemInfoModel

    @classmethod
    def create(cls):
        return ItemCreateModel

    @classmethod
    def db_create(cls):
        return ItemDbCreateModel

    @classmethod
    def db_update(cls):
        """
        Produces the model used db for update.
        """
        return ItemUpdateModel

    @classmethod
    def update(cls):
        return ItemUpdateModel

    @classmethod
    def patch(cls):
        return ItemPatchModel
