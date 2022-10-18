# api-mock-service
## Mocking Distributed Micro services
API mock service for REST/HTTP based services with following goals:
- Record API request/response by working as a proxy server between client and remote service.
- Playback API response that were previously recorded based on request parameters.
- Store API request/responses locally as files so that it's easy to port stubbed request/responses to any machine.
- Allow users to define API request/response with various formats such as XML/JSON/YAML and upload them to the mock service.
- Support template language for generating responses so that users can generate dynamic contents based on input parameters or other configuration. The template language can be used to generate response of any size from small to very large so that you can test performance of your system.
- Support test fixtures that can be uploaded to the mock service and can be used to generate mock responses.
- Generate mock responses using random data or test fixtures.
- Capture mock responses for various input parameters or failure cases so that you can test error handling and fault tolerance of your system.
- Inject error conditions and artificial delays so that you can test how your system handles those error cases.
 
This service is based on an older mock-service https://github.com/bhatti/PlexMockServices, I wrote a while ago.
As, it's written in GO, you can either download GO runtime environment or use Docker to install it locally. 
If you haven't installed docker, you can download the community version from https://docs.docker.com/engine/installation/ 
or find installer for your OS on https://docs.docker.com/get-docker/.

