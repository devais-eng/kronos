from enum import Enum


class ResponseMessage(Enum):
    INTERNAL_SERVER_ERROR = "There was an internal server error while processing the request."
    UNPROCESSABLE_ENTITY = "Cannot process the request because it is semantically incorrect."
    FORBIDDEN = "Access is denied to the requested resource. The user might not have enough permission."
    BAD_REQUEST = "Cannot process the request because it is malformed or incorrect. For example, the requested entity " \
                  "not exists. "
    CONFLICT = "The current state conflicts with what the request expects. For example, the current entity already " \
               "exists. "
