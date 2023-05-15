from typing import Optional
from kronos.storage.view import View, EntityModel, \
    EntityCreateModel, EntityUpdateModel, EntityInfoModel, \
    EntityPatchModel
from pydantic import Field, StrictStr, BaseModel


class AttributeBaseModel(EntityModel):
    """
    Base model for Attributes.
    """
    id: StrictStr = Field(default="attributeId", description="The unique identifier of this entity.")
    type: StrictStr = Field(default="attributeType", description='The literal type of this attribute')
    name: StrictStr = Field(default="attributeName", description="The name of the attribute")
    value: Optional[StrictStr] = Field(default="attributeValue", description="The value of this attribute")
    value_type: Optional[StrictStr] = Field(default="attributeValueType", description="The value of this attribute")
    sync_policy: Optional[StrictStr] = Field(default="attributeSyncPolicy", description="The value of this attribute")


class AttributeModel(AttributeBaseModel):
    id: StrictStr = Field(description="The unique identifier of this entity.")


class AttributeOutputModel(AttributeModel):
    item_id: StrictStr = Field(description='ID of the item which this Attribute is attached to')


class NestedAttributeBaseModel(EntityModel):
    id: StrictStr = Field(description="The unique identifier of this entity.")
    type: StrictStr = Field(description='The literal type of this Attribute')
    name: StrictStr = Field(description="The name of the Attribute")
    value: Optional[StrictStr] = Field(default="attributeValue", description="The value of this attribute")
    value_type: Optional[StrictStr] = Field(default="attributeValueType", description="The value of this attribute")
    sync_policy: Optional[StrictStr] = Field(default="attributeSyncPolicy", description="The value of this attribute")


class NestedAttributeCreateModel(EntityCreateModel, NestedAttributeBaseModel):
    """
    Model for Attributes creation.
    """
    pass


class AttributeCreateModel(EntityCreateModel, AttributeModel):
    """
    Model for Attributes creation.
    """
    item_id: StrictStr = Field(description='ID of the item which this Attribute is attached to')


class AttributeUpdateModel(EntityUpdateModel):
    """
    Model for Attributes update.
    """
    instance_id: Optional[StrictStr] = Field(description='Attribute instance identifier.')
    type: Optional[StrictStr] = Field(description='The literal type of this Attribute')
    name: Optional[StrictStr] = Field(description="The name of the Attribute")

    # item_id: StrictStr = Field(description='ID of the item which this Attribute is attached to')
    # resources: Optional[List[NestedResourceCreateModel]] = Field([], description='New resources to include.')


class AttributeInfoModel(EntityInfoModel, AttributeModel):
    """
    Model for Attributes information.
    """
    item_id: StrictStr = Field(description='ID of the item which this Attribute is attached to')

    class Config:
        orm_mode = True


class AttributeExternalInfoModel(AttributeInfoModel):
    """
    Model for Attributes information.
    """
    instance_id: Optional[StrictStr] = Field(description='ID of the item which this Attribute is attached to')

    class Config:
        orm_mode = True


class AttributePatchModel(EntityPatchModel):
    type: Optional[StrictStr] = Field(description='The literal type of this Attribute')
    name: Optional[StrictStr] = Field(description="The name of the Attribute")
    value: Optional[StrictStr] = Field(default="attributeValue", description="The value of this attribute")
    value_type: Optional[StrictStr] = Field(default="attributeValueType", description="The value of this attribute")
    sync_policy: Optional[StrictStr] = Field(default="attributeSyncPolicy", description="The value of this attribute")


class AttributeValueModel(BaseModel):
    id: StrictStr
    value: StrictStr


class AttributeView(View):
    """
    View for modeling Attributes.
    """

    @classmethod
    def base(cls):
        return AttributeOutputModel

    @classmethod
    def info(cls):
        return AttributeInfoModel

    @classmethod
    def create(cls):
        return AttributeCreateModel

    @classmethod
    def db_create(cls):
        return AttributeCreateModel

    @classmethod
    def db_update(cls):
        return AttributeUpdateModel

    @classmethod
    def update(cls):
        return AttributeUpdateModel

    @classmethod
    def external_info(cls):
        return AttributeExternalInfoModel

    @classmethod
    def patch(cls):
        return AttributePatchModel
