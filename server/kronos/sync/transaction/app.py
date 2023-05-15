from faust_utils.apps import ServiceApp
from faust_utils.model import Response, ResponseCode
from loguru import logger
import uuid
import time
from kronos.sync.transaction.model import TransactionCommand, TransactionRequest, TransactionErrorCode, \
    TransactionErrorResponse
from kronos.exception import get_error_code
from kronos.transaction import TransactionType
from kronos.sync.versioning.manager import VersionManager
from kronos.transaction import get_by_type
from kronos.transaction.delta import DeltaTransaction
from kronos.sync.transaction.roles import Role
import traceback


class TransactionApp(ServiceApp):

    def __init__(self, faust_app):
        super(TransactionApp, self).__init__(faust_app=faust_app)
        self._add_command(TransactionCommand.CRUD, self.crud)
        self._version_manager = VersionManager(app=self._faust_app)

    async def crud(self, crud_request: TransactionRequest):
        """
        Crud operation
        :param crud_request:
        :return:
        """
        start_ts = time.time()
        logger.debug(f"Processing request of type {type(crud_request)} with payload {crud_request}")
        payload = sorted(crud_request.payload, key=lambda x: x.tx_index)
        transactions = []

        for t in payload:
            t_dict = t.asdict()
            if t_dict.get("entity_id", None) is None:
                t_dict["entity_id"] = str(uuid.uuid4())
            transactions.append(get_by_type(TransactionType[t.tx_type])(**t_dict))
        try:
            delta = DeltaTransaction(body=transactions)
            responses = self._version_manager.apply_transactions(delta, role=Role.FORCE)
            logger.debug(f"response : {responses}")

            return Response(caller_request_id=crud_request.request_id,
                            response_id=str(uuid.uuid1()),
                            response_code=ResponseCode.OK,
                            response_type=crud_request.request_type,
                            response_body=responses,
                            ts=time.time(),
                            execution_time=time.time() - start_ts)
        except Exception as e:
            error_code = get_error_code(e)
            traceback_str = traceback.format_tb(e.__traceback__)
            logger.error(f"Unexpected error occurred: {str(e)}\n\n{traceback_str}")
            return TransactionErrorResponse(caller_request_id=crud_request.request_id,
                                            response_id=str(uuid.uuid1()),
                                            ts=time.time(),
                                            response_code=ResponseCode.ERROR,
                                            execution_time=0,
                                            error_code=error_code,
                                            error_message=str(e),
                                            traceback=traceback_str)
