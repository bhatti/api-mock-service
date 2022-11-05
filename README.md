# api-mock-service
## Mocking Distributed Micro services with Record/Play, Templates and OpenAPI Specifications
API mock service for REST/HTTP based services with following goals:
- Record API request/response by working as a proxy server (native http/https or via API) between client and remote service.
- Playback API response that were previously recorded based on request parameters.
- Define API behavior manually by specifying request parameters and response contents.
- Generate API behavior from open standards such as Open API or Swagger.
- Customize API behavior using a template language so that users can generate dynamic contents based on input parameters or other configuration. The template language can be used to generate response of any size from small to very large so that you can test performance of your system.
- Define multiple test scenarios for the API based on different input parameters or simulating various error cases that are difficult to reproduce with real services.
- Store API request/responses locally as files so that it’s easy to port stubbed request/responses to any machine.
- Allow users to define API request/response with various formats such as XML/JSON/YAML and upload them to the mock service.
- Support test fixtures that can be uploaded to the mock service and can be used to generate mock responses.
- Define a collection of helper methods to generate different kind of random data such as UDID, dates, URI, Regex, text and numeric data.
- Ability to playback all test scenarios or a specific scenario and change API behavior dynamically with different input parameters.
- Inject error conditions and artificial delays so that you can test how your system handles error conditions that are difficult to reproduce.

This service is based on an older mock-service https://github.com/bhatti/PlexMockServices, I wrote a while ago.
As, it's written in GO, you can either download GO runtime environment or use Docker to install it locally. 
If you haven't installed docker, you can download the community version from https://docs.docker.com/engine/installation/ 
or find installer for your OS on https://docs.docker.com/get-docker/.
```bash
docker build -t api-mock-service .
docker run -p 8000:8080 -p 8081:8081 -e HTTP_PORT=8080 PROXY_PORT=8081 -e DATA_DIR=/tmp/mocks \
	-e ASSET_DIR=/tmp/assets api-mock-service
```

or pull an image from docker hub (https://hub.docker.com/r/plexobject/api-mock-service), e.g.
```bash
docker pull plexobject/api-mock-service:latest
docker run -p 8000:8080 -p 8081:8081 -e HTTP_PORT=8080 PROXY_PORT=8081 -e DATA_DIR=/tmp/mocks \
	-e ASSET_DIR=/tmp/assets plexobject/api-mock-service:latest
```

Alternatively, you can run it locally with GO environment, e.g.,

```bash
make && ./out/bin/api-mock-service
```

For full command line options, execute api-mock-service -h that will show you command line options such as:
```bash
./out/bin/api-mock-service -h
Starts mock service

Usage:
  api-mock-service [flags]
  api-mock-service [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  version     Version will output the current build information

Flags:
      --assetDir string   asset dir to store static assets/fixtures
      --config string     config file
      --dataDir string    data dir to store mock scenarios
  -h, --help              help for api-mock-service
      --httpPort int      HTTP API port to listen
      --proxyPort int     proxy port to listen

Use "api-mock-service [command] --help" for more information about a command
```

## API Docs
See Swagger API docs at https://petstore.swagger.io?url=https://raw.githubusercontent.com/bhatti/api-mock-service/main/docs/swagger.yaml

## Recording a Mock Scenario via HTTP/HTTPS Proxy
Once you have the API mock service running, the mock service will start two ports on startup, first port
(default 8080) will be used to record/play mock scenarios, updating templates or uploading OpenAPIs. The second
port (default 8081) will setup an HTTP/HTTPS proxy server that you can point to record your scenarios, e.g.
```shell
export http_proxy="http://localhost:8081"
export https_proxy="http://localhost:8081"

curl -k -v -H "Authorization: Bearer sk_test_xxxx" https://api.stripe.com/v1/customers/cus_xxx/cash_balance
```

Above curl command will automatically record all requests and responses and create mock scenario to play it back. For example, if you call the same 
API again, it will return a local response instead of contacting the server. You can customize the proxy behavior for record by adding `X-Mock-Record: true`  header to your request.

## Recording a Mock Scenario via API 
Alternatively, you can use invoke an internal API as a pass through to invoke a remote API so that you can 
automatically record API behavior and play it back later, e.g.

```bash
curl -H "X-Mock-Url: https://api.stripe.com/v1/customers/cus_**/cash_balance" \
	-H "Authorization: Bearer sk_test_***" http://localhost:8080/_proxy
```

In above example, the curl command is passing the URL of real service as an HTTP header ``X-Mock-Url``. In addition, you can pass 
other authorization headers as needed. 

## Viewing the Recorded Mock Scenario
The API mock-service will store the request/response in a YAML file under a data directory that you can specify. For example, you may see a file under:

```bash
default_mocks_data/v1/customers/cus_***/cash_balance/GET/recorded-scenario-***.scr
```

Note: the sensitive authentication or customer keys are masked in above example but you will see following contents in the captured data file:

```yaml
method: GET
name: recorded-v1-customers-cus
path: /v1/customers/cus_**/cash_balance
description: recorded at 2022-10-29 04:26:17.24776 +0000 UTC
request:
    match_query_params: {}
    match_headers: {}
    match_content_type: ""
    match_contents: ""
    example_path_params: {}
    example_query_params: {}
    example_headers:
        Accept: '*/*'
        Authorization: Bearer sk_test_xxx
        User-Agent: curl/7.65.2
        X-Mock-Url: https://api.stripe.com/v1/customers/cus_/cash_balance
    example_contents: ""
response:
    headers:
        Access-Control-Allow-Credentials:
            - "true"
        Access-Control-Allow-Methods:
            - GET, POST, HEAD, OPTIONS, DELETE
        Access-Control-Allow-Origin:
            - '*'
        Access-Control-Expose-Headers:
            - Request-Id, Stripe-Manage-Version, X-Stripe-External-Auth-Required, X-Stripe-Privileged-Session-Required
        Access-Control-Max-Age:
            - "300"
        Cache-Control:
            - no-cache, no-store
        Content-Length:
            - "168"
        Content-Type:
            - application/json
        Date:
            - Sat, 29 Oct 2022 04:26:17 GMT
        Request-Id:
            - req_xxx
        Server:
            - nginx
        Strict-Transport-Security:
            - max-age=63072000; includeSubDomains; preload
        Stripe-Version:
            - "2018-09-06"
    content_type: application/json
    contents: |-
        {
          "object": "cash_balance",
          "available": null,
          "customer": "cus_",
          "livemode": false,
          "settings": {
            "reconciliation_mode": "automatic"
          }
        }
    contents_file: ""
    status_code: 200
wait_before_reply: 0s
```

Above example defines a mock scenario for testing /v1/customers/cus_**/cash_balance path. A test scenario includes:
### Request Matching Parameters:

The matching request parameters will be used to select the mock scenario to execute and you can use regular expressions to validate:

    - URL Query Parameters
    - URL Request Headers
    - Request Body

### Example Request Parameters:

The example request parameters show the contents captured from the record/play so that you can use and customize to define matching parameters.

    - URL Query Parameters
    - URL Request Headers
    - Request Body

### Response Properties

The response properties will include:

    - Response Headers
    - Response Body statically defined or loaded from a test fixture

You can copy recorded scenario to another folder and use templates to customize it and then upload it for playback.

## Playback the Mock API Scenario

You can playback the recorded response from above example as follows:
```bash
% curl http://localhost:8080/v1/customers/cus_***/cash_balance
```

Which will return captured response such as:

```json

{
  "object": "cash_balance",
  "available": null,
  "customer": "cus_***",
  "livemode": false,
  "settings": {
    "reconciliation_mode": "automatic"
  }
}

```

Though, you can customize your template with dynamic properties or conditional logic but you can also send HTTP headers 
for `X-Mock-Response-Status` to override HTTP status to return or `X-Mock-Wait-Before-Reply` to add artificial latency using duration syntax. 

## Upload Mock API Scenario
You can customize the recorded scenario, e.g. you can add path variables to above API as follows:
```yaml
method: GET
name: stripe-cash-balance
path: /v1/customers/:customer/cash_balance
request:
  match_headers:
    Authorization: Bearer sk_test_[0-9a-fA-F]{10}$
response:
  headers:
    Access-Control-Allow-Credentials:
      - "true"
    Access-Control-Allow-Methods:
      - GET, POST, HEAD, OPTIONS, DELETE
    Access-Control-Allow-Origin:
      - '*'
    Access-Control-Expose-Headers:
      - Request-Id, Stripe-Manage-Version, X-Stripe-External-Auth-Required, X-Stripe-Privileged-Session-Required
    Access-Control-Max-Age:
      - "300"
    Cache-Control:
      - no-cache, no-store
    Content-Type:
      - application/json
    Request-Id:
      - req_2
    Server:
      - nginx
    Strict-Transport-Security:
      - max-age=63072000; includeSubDomains; preload
    Stripe-Version:
      - "2018-09-06"
  content_type: application/json
  contents: |-
    {
      "object": "cash_balance",
      "available": null,
      "customer": {{.customer}}
      "livemode": false,
      "page": {{.page}}
      "pageSize": {{.pageSize}}
      "settings": {
        "reconciliation_mode": "automatic"
      }
    }
  status_code: 200
wait_before_reply: 1s
```

In above example, I assigned a name stripe-cash-balance to the mock scenario and changed API path to 
`/v1/customers/:customer/cash_balance` so that it can capture customer-id as a path variable. I added a regular expression to 
ensure that the HTTP request includes an Authorization header matching `Bearer sk_test_[0-9a-fA-F]{10}$` and defined dynamic properties 
such as `{{.customer}}, {{.page}} and {{.pageSize}}` so that they will be replaced at runtime.

```bash
curl -H "Content-Type: application/yaml" --data-binary @fixtures/stripe-customer.yaml \
	http://localhost:8080/_scenarios
```
and then play it back as follows:
```bash
curl -v -H "Authorization: Bearer sk_test_0123456789" \
	"http://localhost:8080/v1/customers/123/cash_balance?page=2&pageSize=55"
```

and it will generate:
```json
{
  "object": "cash_balance",
  "available": null,
  "customer": 123,
  "livemode": false,
  "page": 2,
  "pageSize": 55,
  "settings": {
    "reconciliation_mode": "automatic"
  }
}

```
As you can see, the values of customer, page and pageSize are dynamically updated. You can upload multiple mock scenarios for the same API and the mock API service will play it back sequentially. For example, you can upload another scenario for above API as follows:
```yaml
method: GET
name: stripe-customer-failure
path: /v1/customers/:customer/cash_balance
request:
    match_headers:
        Authorization:
            - Bearer sk_test
response:
    headers:
        Stripe-Version:
            - "2018-09-06"
    content_type: application/json
    contents: My custom error
    status_code: 500
wait_before_reply: 1s
```
And then play it back:
```bash
curl -v "http://localhost:8080/v1/customers/123/cash_balance?page=2&pageSize=55"
```
which will return response with following error response
```
> GET /v1/customers/123/cash_balance?page=2&pageSize=55 HTTP/1.1
> Host: localhost:8080
> User-Agent: curl/7.65.2
> Accept: */*
> Authorization: Bearer sk_test_0123456789
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 500 Internal Server Error
< Content-Type: application/json
< X-Mock-Request-Count: 1
< X-Mock-Scenario: stripe-customer-failure
< Stripe-Version: 2018-09-06
< Vary: Origin
< Date: Sat, 29 Oct 2022 17:29:15 GMT
< Content-Length: 15
```

## Dynamic Templates with Mock API Scenarios
You can use loops and conditional primitives of template language (https://golangdocs.com/templates-in-golang) to generate dynamic responses as follows:
```yaml
method: GET
name: get_devices
path: /devices
description: ""
request:
    match_content_type: "application/json; charset=utf-8"
response:
    headers:
        "Server":
            - "SampleAPI"
        "Connection":
            - "keep-alive"
    content_type: application/json
    contents: >
     {
     "Devices": [
{{- range $val := Iterate .pageSize }}
      {
        "Udid": "{{SeededUdid $val}}",
        "Line": { {{SeededFileLine "lines.txt" $val}}, "Type": "Public", "IsManaged": false },
        "Amount": {{JSONFileProperty "props.yaml" "amount"}},
        "SerialNumber": "{{Udid}}",
        "MacAddress": "{{Udid}}",
        "Imei": "{{Udid}}",
        "AssetNumber": "{{RandString 20}}",
        "LocationGroupId": {
         "Id": {
           "Value": {{RandNumMax 1000}},
         },
         "Name": "{{SeededCity $val}}",
         "Udid": "{{Udid}}"
        },
        "DeviceFriendlyName": "Device for {{SeededName $val}}",
        "LastSeen": "{{Time}}",
        "Email": "{{RandEmail}}",
        "Phone": "{{RandPhone}}",
        "EnrollmentStatus": {{SeededBool $val}}
        "ComplianceStatus": {{RandRegex "^AC[0-9a-fA-F]{32}$"}}
        "Group": {{RandCity}},
        "Date": {{TimeFormat "3:04PM"}},
        "BatteryLevel": "{{RandNumMax 100}}%",
        "StrEnum": {{EnumString "ONE TWO THREE"}},
        "IntEnum": {{EnumInt 10 20 30}},
        "ProcessorArchitecture": {{RandNumMax 1000}},
        "TotalPhysicalMemory": {{RandNumMax 1000000}},
        "VirtualMemory": {{RandNumMax 1000000}},
        "AvailablePhysicalMemory": {{RandNumMax 1000000}},
        "CompromisedStatus": {{RandBool}},
        "Add": {{Add 2 1}},
        "Dict": {{Dict "one" 1 "two" 2 "three" 3}}
      }{{if LastIter $val $.PageSize}}{{else}},  {{end}}
{{ end }}
     ],
     "Page": {{.page}},
     "PageSize": {{.pageSize}},
     "Total": {{.pageSize}}
     }
    {{if NthRequest 10 }}
    status_code: {{EnumInt 500 501}}
    {{else}}
    status_code: {{EnumInt 200 400}}
    {{end}}
wait_before_reply: {{.page}}s
```
Above example includes a number of template primitives and custom functions to generate dynamic contents such as:
### Loops
GO template support loops that can be used to generate multiple data entries in the response, e.g.
```yaml
{{- range $val := Iterate .pageSize }}
```

### Artificial Delays
You can specify artificial delay for the API request as follows:
```yaml
wait_before_reply: {{.page}}s
```
Above example shows delay based on page number but you can use any parameter to customize this behavior.

### Builtin functions
Go template allows custom functions that can provide customized behavior for generating test data, e.g.:
```yaml
        "SerialNumber": "{{Udid}}",
        "AssetNumber": "{{RandString 20}}",
        "LastSeen": "{{Time}}",
        "Email": "{{RandEmail}}",
        "Phone": "{{RandPhone}}",
        "EnrollmentStatus": {{SeededBool $val}}
        "ComplianceStatus": {{RandRegex "^AC[0-9a-fA-F]{32}$"}}
        "Group": {{RandCity}},
        "Date": {{TimeFormat "3:04PM"}},
        "BatteryLevel": "{{RandNumMax 100}}%",
        "StrEnum": {{EnumString "ONE TWO THREE"}},
        "IntEnum": {{EnumInt 10 20 30}},
```

### Conditional Logic
The template syntax allows you to define a conditional logic such as:
```yaml
    {{if NthRequest 10 }}
    status_code: {{AnyInt 500 501}}
    {{else}}
    status_code: {{AnyInt 200 400}}
    {{end}}
```
In above example, the mock API will return HTTP status 500 or 501 for every 10th request and 200 or 400 for other requests. You can use conditional syntax to simulate different error status or customize response.

### Test fixtures
The mock service allows you to upload a test fixture that you can refer in your template, e.g. 
```bash
  "Line": { {{SeededFileLine "lines.txt" $val}}, "Type": "Public", "IsManaged": false },
```
Above example loads a random line from a lines.txt fixture. As you may need to generate a deterministic random data in some cases, 
you can use Seeded functions to generate predictable data so that the service returns same data. Following example will read a 
text fixture to load a property from a file:
```bash
  "Amount": {{JSONFileProperty "props.yaml" "amount"}},
```

This template file will generate content as follows:
```json
{ "Devices": [
 {
   "Udid": "fe49b338-4593-43c9-b1e9-67581d000000",
   "Line": { "ApplicationName": "Chase", "Version": "3.80", "ApplicationIdentifier": "com.chase.sig.android", "Type": "Public", "IsManaged": false },
   "Amount": {"currency":"$","value":100},
   "SerialNumber": "47c2d7c3-c930-4194-b560-f7b89b33bc2a",
   "MacAddress": "1e015eac-68d2-42ee-9e8f-73fb80958019",
   "Imei": "5f8cae1b-c5e3-4234-a238-1c38d296f73a",
   "AssetNumber": "9z0CZSA03ZbUNiQw2aiF",
   "LocationGroupId": {
    "Id": {
      "Value": 980
    },
    "Name": "Houston",
    "Udid": "3bde6570-c0d4-488f-8407-10f35902cd99"
   },
   "DeviceFriendlyName": "Device for Alexander",
   "LastSeen": "2022-10-29T11:25:25-07:00",
   "Email": "john.smith@abc.com",
   "Phone": "1-408-454-1507",
   "EnrollmentStatus": true,
   "ComplianceStatus": "ACa3E07B0F2cA00d0fbFe88f5c6DbC6a9e",
   "Group": "Chicago",
   "Date": "11:25AM",
   "BatteryLevel": "43%",
   "StrEnum": "ONE",
   "IntEnum": 20,
   "ProcessorArchitecture": 243,
   "TotalPhysicalMemory": 320177,
   "VirtualMemory": 768345,
   "AvailablePhysicalMemory": 596326,
   "CompromisedStatus": false,
   "Add": 3
 },
...
 ], "Page": 2, "PageSize": 55, "Total": 55 }  
```

## Playing back a specific mock scenario
You can pass a header for ``X-Mock-Scenario`` to specify the name of scenario if you have multiple scenarios for the same API, e.g. 
```bash
curl -v -H "X-Mock-Scenario: stripe-cash-balance" -H "Authorization: Bearer sk_test_0123456789" \
	"http://localhost:8080/v1/customers/123/cash_balance?page=2&pageSize=55"
```

You can also customize response status by overriding the request header with `X-Mock-Response-Status` and delay before return by
overriding `X-Mock-Wait-Before-Reply` header.

## Using Test Fixtures
You can define a test data in your test fixtures and then upload as follows:
```bash
curl -H "Content-Type: application/yaml" --data-binary @fixtures/lines.txt \
	http://localhost:8080/_fixtures/GET/lines.txt/devices
	
curl -v -H "Content-Type: application/yaml" --data-binary @fixtures/props.yaml \
    http://localhost:8080/_fixtures/GET/props.yaml/devices	
```

In above example, test fixtures for `lines.txt` and `props.yaml` will be uploaded and will be available for all `GET` requests 
under `/devices` URL path. You can then refer to above fixture in your templates. You can also use this to serve any binary files, 
e.g. you can define an image template file as follows:

```yaml
method: GET
name: test-image
path: /images/mock_image
description: ""
request:
response:
    headers:
      "Last-Modified":
        - {{Time}}
      "ETag":
        - {{RandString 10}}
      "Cache-Control":
        - max-age={{RandNumMinMax 1000 5000}}
    content_type: image/png
    contents_file: mockup.png
    status_code: 200
```

Then upload a binary image using:
```bash
curl -H "Content-Type: application/yaml" --data-binary @fixtures/mockup.png \
	http://localhost:8080/_fixtures/GET/mockup.png/images/mock_image
```

And then serve the image using:
```bash
curl -v "http://localhost:8080/images/mock_image"
```

## Custom Functions 
### Numeric Random Data

Following functions can be used to generate numeric data within a range or with a seed to always generate deterministic test data:

- Random
- SeededRandom
- RandNumMinMax
- RandIntArrayMinMax

### Text Random Data

Following functions can be used to generate numeric data within a range or with a seed to always generate deterministic test data:

- RandStringMinMax
- RandStringArrayMinMax
- RandRegex
- RandEmail
- RandPhone

### Boolean

Following functions can be used to generate boolean data:

- RandBool
- SeededBool

### UDID

Following functions can be used to generate UDIDs:

- Udid
- SeededUdid

### String Enums

Following functions can be used to generate a string from a set of Enum values:

- EnumString

### Integer Enums

Following functions can be used to generate an integer from a set of Enum values:

-  EnumInt

### Random Names

Following functions can be used to generate random names:

- RandName
- SeededName

### City Names

Following functions can be used to generate random city names:

- RandCity
- SeededCity

### Country Names or Codes

Following functions can be used to generate random country names or codes:

- RandCountry
- SeededCountry
- RandCountryCode
- SeededCountryCode

### File Fixture

Following functions can be used to generate random data from a fixture file:

- RandFileLine
- SeededFileLine
- FileProperty
- JSONFileProperty
- YAMLFileProperty

## Generate Mock API Behavior from OpenAPI or Swagger Specifications
If you are using Open API (https://www.openapis.org/) or Swagger for API specifications, you can simply upload a YAML based API specification. For example, here is a sample Open API specification from Twilio:
```yaml
openapi: 3.0.1
paths:
  /v1/AuthTokens/Promote:
    servers:
    - url: https://accounts.twilio.com
    description: Auth Token promotion
    x-twilio:
      defaultOutputProperties:
      - account_sid
      - auth_token
      - date_created
      pathType: instance
      mountName: auth_token_promotion
    post:
      description: Promote the secondary Auth Token to primary. After promoting the
        new token, all requests to Twilio using your old primary Auth Token will result
        in an error.
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/accounts.v1.auth_token_promotion'
          description: OK
      security:

...


   schemas:
     accounts.v1.auth_token_promotion:
       type: object
       properties:
         account_sid:
           type: string
           minLength: 34
           maxLength: 34
           pattern: ^AC[0-9a-fA-F]{32}$
           nullable: true
           description: The SID of the Account that the secondary Auth Token was created
             for
         auth_token:
           type: string
           nullable: true
           description: The promoted Auth Token
         date_created:
           type: string
           format: date-time
           nullable: true
           description: The ISO 8601 formatted date and time in UTC when the resource
             was created
         date_updated:
           type: string
           format: date-time
           nullable: true
           description: The ISO 8601 formatted date and time in UTC when the resource
             was last updated
         url:
           type: string
           format: uri
           nullable: true
           description: The URI for this resource, relative to `https://accounts.twilio.com`
...         
```
You can then upload the API specification as:
```bash
curl -H "Content-Type: application/yaml" --data-binary @fixtures/oapi/twilio_accounts_v1.yaml \
		http://localhost:8080/_oapi
```a

It will generate a mock scenarios for each API based on mime-type, status-code, parameter formats, regex, data ranges, e.g.,

```yaml
name: UpdateAuthTokenPromotion-xx
path: /v1/AuthTokens/Promote
description: Promote the secondary Auth Token to primary. After promoting the new token, all requests to Twilio using your old primary Auth Token will result in an error.
request:
    match_query_params: {}
    match_headers: {}
    match_content_type: ""
    match_contents: ""
response:
    headers: {}
    content_type: application/json
    contents: '{"account_sid":"{{RandRegex `^AC[0-9a-fA-F]{32}$`}}",\
    "auth_token":"{{RandStringMinMax 0 0}}","date_created":"{{Time}}",\
    "date_updated":"{{Time}}","url":"https://{{RandName}}.com"}'
    contents_file: ""
    status_code: 200
wait_before_reply: 0s
```

In above example, the account_sid uses regex to generate data and URI format to generate URL. Then invoke the mock API as:

```bash
curl -v -X POST http://localhost:8080/v1/AuthTokens/Promote
```

Which will generate dynamic response as follows:

```json
{
  "account_sid": "ACF3A7ea7f5c90f6482CEcA77BED07Fb91",
  "auth_token": "PaC7rKdGER73rXNi6rVKZMN1Jw0QYxPFeEkqyvnM7Ojw2nziOER7SMWkIV6N2hXYTKxAfDMfS9t0",
  "date_created": "2022-10-29T11:54:46-07:00",
  "date_updated": "2022-10-29T11:54:46-07:00",
  "url": "https://Billy.com"
}
```
## Listing all Mock Scenarios
You can list all available mock APIs using:
```shell
curl -v http://localhost:8080/_scenarios
```
Which will return summary of APIs such as:

```json
{
  "/_scenarios/GET/FetchCredentialAws-8b2fcf02dfb7dc190fb735a469e1bbaa3ccb5fd1a24726976d110374b13403c6/v1/Credentials/AWS/{Sid}": {
    "method": "GET",
    "name": "FetchCredentialAws-8b2fcf02dfb7dc190fb735a469e1bbaa3ccb5fd1a24726976d110374b13403c6",
    "path": "/v1/Credentials/AWS/{Sid}",
    "match_query_params": {},
    "match_headers": {},
    "match_content_type": "",
    "match_contents": "",
    "LastUsageTime": 0,
    "RequestCount": 0
  },
  "/_scenarios/GET/FetchCredentialPublicKey-60a01dcea5290e6d429ce604c7acf5bd59606045fc32c0bc835e57ac2b1b8eb6/v1/Credentials/PublicKeys/{Sid}": {
    "method": "GET",
    "name": "FetchCredentialPublicKey-60a01dcea5290e6d429ce604c7acf5bd59606045fc32c0bc835e57ac2b1b8eb6",
    "path": "/v1/Credentials/PublicKeys/{Sid}",
    "match_query_params": {},
    "match_headers": {},
    "match_content_type": "",
    "match_contents": "",
    "LastUsageTime": 0,
    "RequestCount": 0
  },
  "/_scenarios/GET/ListCredentialAws-28717701f05de4374a09ec002066d308043e73e30f25fec2dcd4c3d3c001d300/v1/Credentials/AWS": {
    "method": "GET",
    "name": "ListCredentialAws-28717701f05de4374a09ec002066d308043e73e30f25fec2dcd4c3d3c001d300",
    "path": "/v1/Credentials/AWS",
    "match_query_params": {
      "PageSize": "\\d+"
    },
    "match_headers": {},
    "match_content_type": "",
    "match_contents": "",
    "LastUsageTime": 0,
    "RequestCount": 0
  },
...  
```
## Static Assets
The mock service can serve any static assets from a user-defined folder and then serve it as follows:
```bash
cp static-file default_assets

# execute the API mock server
make && ./out/bin/api-mock-service

# access assets
curl http://localhost:8080/_assets/default_assets
```
