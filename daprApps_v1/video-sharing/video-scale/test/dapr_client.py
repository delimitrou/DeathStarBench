from dapr.conf import settings
from dapr.proto import api_v1, api_service_v1, common_v1
from dapr.clients import DaprClient
from dapr.clients.grpc._helpers import DaprClientInterceptor
from dapr.clients.http.dapr_invocation_http_client import DaprInvocationHttpClient
from dapr.clients.exceptions import DaprInternalError
from typing import Optional, List, Union, Dict, Tuple, Callable

import grpc  # type: ignore
from grpc import (  # type: ignore
    UnaryUnaryClientInterceptor,
    UnaryStreamClientInterceptor,
    StreamUnaryClientInterceptor,
    StreamStreamClientInterceptor
)

# MyDaprClient is a wrapper of dapr python-sdk DaprClient
# that supports additional grpc options
class MyDaprClient(DaprClient):
    def __init__(self,
        address: Optional[str] = None,
        headers_callback: Optional[Callable[[], Dict[str, str]]] = None,
        options: Optional[List[Tuple]] = None,
        interceptors: Optional[List[Union[
            UnaryUnaryClientInterceptor,
            UnaryStreamClientInterceptor,
            StreamUnaryClientInterceptor,
            StreamStreamClientInterceptor]]] = None,
        http_timeout_seconds: Optional[int] = None):
        
        """Connects to Dapr Runtime and via gRPC and HTTP.

        Args:
            address (str, optional): Dapr Runtime gRPC endpoint address.
            headers_callback (lambda: Dict[str, str]], optional): Generates header for each request.
            options (List[Tuple], optional): grpc channel options
            headers_callback (lambda: Dict[str, str]], optional): Generates header for each request.
            interceptors (list of UnaryUnaryClientInterceptor or
                UnaryStreamClientInterceptor or
                StreamUnaryClientInterceptor or
                StreamStreamClientInterceptor, optional): gRPC interceptors.
            http_timeout_seconds (int): specify a timeout for http connections
        """
        #---------- initialize the grpc client ----------#
        """Connects to Dapr Runtime and initialize gRPC client stub."""
        if not address:
            address = f"{settings.DAPR_RUNTIME_HOST}:{settings.DAPR_GRPC_PORT}"
        self._address = address
        self._channel = grpc.insecure_channel(address, options=options)   # type: ignore

        if settings.DAPR_API_TOKEN:
            api_token_interceptor = DaprClientInterceptor([
                ('dapr-api-token', settings.DAPR_API_TOKEN), ])
            self._channel = grpc.intercept_channel(   # type: ignore
                self._channel, api_token_interceptor)
        if interceptors:
            self._channel = grpc.intercept_channel(   # type: ignore
                self._channel, *interceptors)

        self._stub = api_service_v1.DaprStub(self._channel)

        #---------- init invocation client ----------#
        self.invocation_client = None

        invocation_protocol = settings.DAPR_API_METHOD_INVOCATION_PROTOCOL.upper()

        if invocation_protocol == 'HTTP':
            if http_timeout_seconds is None:
                http_timeout_seconds = settings.DAPR_HTTP_TIMEOUT_SECONDS
            self.invocation_client = DaprInvocationHttpClient(headers_callback=headers_callback,
                                                              timeout=http_timeout_seconds)
        elif invocation_protocol == 'GRPC':
            pass
        else:
            raise DaprInternalError(
                f'Unknown value for DAPR_API_METHOD_INVOCATION_PROTOCOL: {invocation_protocol}')