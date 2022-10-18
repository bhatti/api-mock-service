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
```bash
docker build -t api-mock-service .
docker run -p 8000:8080 -e HTTP_PORT=8080 -e DATA_DIR=/tmp/mocks \
	-e ASSET_DIR=/tmp/assets api-mock-service
```

or if you have GO environment then

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
      --port int          HTTP port to listen

Use "api-mock-service [command] --help" for more information about a command
```

## Recording a Mock API Scenario
Once you have the API mock service running, you can use as a proxy service to invoke a remote API so that you can automatically record API behavior and play it back later, e.g.
```bash
curl -H "Mock-Url: https://api.stripe.com/v1/customers/cus_**/cash_balance" \
	-H "Authorization: Bearer sk_test_***" http://localhost:8080/_proxy
```

In above example, the curl command is passing the URL of real service as an HTTP header Mock-Url. In addition, you can pass other authorization headers as needed. The API mock-service will store the request/response in a YAML file under a data directory that you can specify. For example, you may see a file under:

```bash
default_mocks_data/v1/customers/cus_***/cash_balance/GET/recorded-scenario-***.scr
```

Note: the sensitive authentication or customer keys are masked in above example but you will see following contents in the captured data file:

```yaml
method: GET
name: recorded-scenario-***
path: /v1/customers/cus_***/cash_balance
description: recorded at 2022-10-18 02:56:24.417175 +0000 UTC
request:
    query_params: ""
    headers:
        Accept:
            - '*/*'
        Authorization:
            - Bearer sk_test_***
        Mock-Url:
            - https://api.stripe.com/v1/customers/cus_***/cash_balance
        User-Agent:
            - curl/7.65.2
    content_type: ""
    contents: ""
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
            - Tue, 18 Oct 2022 02:56:24 GMT
        Request-Id:
            - req_***
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
          "customer": "cus_***",
          "livemode": false,
          "settings": {
            "reconciliation_mode": "automatic"
          }
        }
    status_code: 200
wait_before_reply: 0s
```
You can optionally copy it to another folder so that you can customize it and then upload it for playback.

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
}%

```

## Upload Mock API Scenario
You can customize the recorded scenario, e.g. you can add path variables to above API as follows:
```yaml
method: GET
name: stripe-cash-balance
path: /v1/customers/:customer/cash_balance
request:
    headers:
        Accept:
            - '*/*'
        Authorization:
            - Bearer sk_test
        User-Agent:
            - curl/7.65.2
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

In above example, I assigned a name stripe-cash-balance to the mock scenario and changed API path to /v1/customers/:customer/cash_balance so that it can capture customer-id as a path variable. I also added dynamic properties such as {{.customer}}, {{.page}} and {{.pageSize}}. The mock scenario uses builtin template syntax of GO. You can then upload it as follows:
```bash
curl -H "Content-Type: application/yaml" --data-binary @fixtures/stripe-customer.yaml \
	http://localhost:8080/_scenarios
```
and then play it back as follows:
```bash
curl -v "http://localhost:8080/v1/customers/123/cash_balance?page=2&pageSize=55"
```

and it will generate:
```json
{
  "object": "cash_balance",
  "available": null,
  "customer": 123
  "livemode": false,
  "page": 2
  "pageSize": 55
  "settings": {
    "reconciliation_mode": "automatic"
}

```
As you can see, the values of customer, page and pageSize are dynamically updated. You can upload multiple mock scenarios for the same API and the mock API service will play it back sequentially. For example, you can upload another scenario with above API as follows:
```yaml
method: GET
name: stripe-customer-failure
path: /v1/customers/:customer/cash_balance
request:
    headers:
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
with following error response
```json
* Mark bundle as not supporting multiuse
< HTTP/1.1 500 Internal Server Error
< Content-Type: application/json
< Stripe-Version: 2018-09-06
< Vary: Origin
< Date: Tue, 18 Oct 2022 03:38:35 GMT
< Content-Length: 15
```

## Dynamic Mock API Scenario
You can use loops and conditional primitives of template language to generate dynamic responses as follows:
```yaml
method: GET
name: get_devices
path: /devices
description: ""
request:
    query_params:
    content_type: "application/json; charset=utf-8"
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
        "EnrollmentStatus": "Enrolled",
        "ComplianceStatus": "NonCompliant",
        "BatteryLevel": "{{RandNumMax 100}}%",
        "ProcessorArchitecture": {{RandNumMax 1000}},
        "TotalPhysicalMemory": {{RandNumMax 1000000}},
        "VirtualMemory": {{RandNumMax 1000000}},
        "AvailablePhysicalMemory": {{RandNumMax 1000000}},
        "CompromisedStatus": {{RandBool}}
      }{{if LastIter $val $.PageSize}}{{else}},  {{end}}
{{ end }}
     ],
     "Page": {{.page}},
     "PageSize": {{.pageSize}},
     "Total": {{.pageSize}}
     }
    {{if LT .page 10 }}
    status_code: 200
    {{else}}
    status_code: 400
    {{end}}
wait_before_reply: {{.page}}s
```
Above example includes a number of template primitives and custom functions to generate dynamic contents such as:
### Loops
GO template support loops that can be used to generate multiple data entries in the respons, e.g.
```yaml
{{- range $val := Iterate .pageSize }}
```
### Custom functions
```yaml
"SerialNumber": "{{Udid}}",
"Name": "{{SeededCity $val}}",
"TotalPhysicalMemory": {{RandNumMax 1000000}},
```

### Conditional Logic
The template syntax allows you to define a conditional logic such as:
```yaml
{{if LT .page 10 }}
    status_code: 200
{{else}}
    status_code: 400
{{end}}
```
### Test fixtures
The mock service allows you to upload a test fixture that you can refer in your template, e.g. 
```bash
"Line": { {{SeededFileLine "lines.txt" $val}}, "Type": "Public", "IsManaged": false },
```
Above example loads a random line from a lines.txt fixture. As you may need to generate a deterministic random data in some cases, you can use Seeded functions to generate predictable data so that the service returns same 
data. (More examples of test fixtures are described below.) This template file will generate content as follows:
```json
{ "Devices": [
  {
    "Udid": "8d94d137-fb53-47b4-b71e-bb8cf5000000",
    "Line": { "ApplicationName": "Settings", "Version": "8.1.0", "ApplicationIdentifier": "com.android.settings", "Type": "Public", "IsManaged": false },
    "SerialNumber": "c2d9f005-b79f-4222-802a-ba2d38a35995",
    "MacAddress": "cf621298-4ab0-4c03-b536-f4f5ae0423df",
    "Imei": "07ee02fb-f58d-40b8-a9ff-111aa0929dcb",
    "AssetNumber": "2evjQvCjO3fAGk6IwYDk",
    "LocationGroupId": {
     "Id": {
       "Value": 51,
     },
     "Name": "Singapore",
     "Udid": "0b7d4f61-17ce-4973-8cb9-52277e97eb66"
    },
    "DeviceFriendlyName": "Device for Katherine",
    "LastSeen": "2022-10-18T09:12:53-07:00",
    "EnrollmentStatus": "Enrolled",
    "ComplianceStatus": "NonCompliant",
    "BatteryLevel": "84%",
    "ProcessorArchitecture": 8,
    "TotalPhysicalMemory": 160369,
    "VirtualMemory": 800601,
    "AvailablePhysicalMemory": 784811,
    "CompromisedStatus": true
  },
...
 ], "Page": 2, "PageSize": 55, "Total": 55 }   
```

## Playing back a specific mock scenario
You can pass a header for Mock-Scenario to specify the name of scenario if you have multiple scenarios for the same API, e.g. 
```bash
curl -v -H "Mock-Scenario: stripe-cash-balance" \
	"http://localhost:8080/v1/customers/123/cash_balance?page=2&pageSize=55"
```

## Using Test Fixtures
You can define a test data in your test fixtures and then upload as follows:
```bash
curl -H "Content-Type: application/yaml" --data-binary @fixtures/lines.txt \
	http://localhost:8080/_fixtures/GET/lines.txt/devices
```

In above example, a test fixture for lines.txt will be uploaded and will be available for all GET requests under /devices URL path. You can then refer to above fixture in your templates. You can also use this to serve any binary files, e.g. you can define an image template file as follows:

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

## Static Assets
The mock service can serve any static assets from a user-defined folder and then serve it as follows:
```bash
cp static-file default_assets
curl http://localhost:8080/_assets/default_assets
```

## Summary
Building and testing distributed systems often requires deploying a deep stack of dependent services, which makes development hard on a local environment with limited resources. Ideally, you should be able to deploy and test entire stack without using network or requiring remote access so that you can spend more time on building features instead of configuring your local environment. Above examples show how you use a https://github.com/bhatti/api-mock-service to mock APIs for testing purpose and define test scenarios for simulating both happy and error handling as well as inject fault tolerance or network delays in your testing processes. I have found a great use of tools like this when developing micro services and hopefully you find it useful. Feel free to connect with your feedback or suggestions.