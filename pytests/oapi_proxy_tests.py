import unittest
import requests
import time
import json

proxy_servers = {
    'http': 'http://localhost:8081',
    'https': 'http://localhost:8081',
}

class OAPIProxyTest(unittest.TestCase):
    def test_generate_oapi_twilio_scenarios_via_proxy(self):
        data = open('../fixtures/oapi/twilio_accounts_v1.yaml', 'r').read()
        headers = {
            'Content-Type': 'application/yaml',
        }
        resp = requests.post('http://localhost:8081/_oapi', data = data, headers = headers, proxies = proxy_servers, verify = False)
        self.assertEqual(200, resp.status_code)

        headers = {
            'Content-Type': 'application/json',
        }
        resp = requests.post('http://localhost:8081/v1/AuthTokens/Promote', headers = headers, proxies = proxy_servers, verify = False)
        self.assertEqual(200, resp.status_code)

        resp = requests.post('http://localhost:8081/v1/AuthTokens/Secondary', headers = headers, proxies = proxy_servers, verify = False)
        self.assertEqual(201, resp.status_code)

        resp = requests.get('http://localhost:8081/_scenarios', headers = headers, proxies = proxy_servers, verify = False)
        self.assertEqual(200, resp.status_code)

    def test_generate_oapi_twilio_scenarios(self):
        data = open('../fixtures/oapi/twilio_accounts_v1.yaml', 'r').read()
        headers = {
            'Content-Type': 'application/yaml',
        }
        resp = requests.post('http://localhost:8080/_oapi', data = data, headers = headers)
        self.assertEqual(200, resp.status_code)

        headers = {
            'Content-Type': 'application/json',
        }
        resp = requests.post('http://localhost:8080/v1/AuthTokens/Promote', headers = headers)
        self.assertEqual(200, resp.status_code)

        resp = requests.post('http://localhost:8080/v1/AuthTokens/Secondary', headers = headers)
        self.assertEqual(201, resp.status_code)

        resp = requests.get('http://localhost:8080/_scenarios', headers = headers)
        self.assertEqual(200, resp.status_code)

    def test_generate_oapi_jobs_scenarios(self):
        data = open('../fixtures/oapi/jobs-openapi.json', 'r').read()
        resp = requests.post('http://localhost:8080/_oapi', data = data)
        self.assertEqual(200, resp.status_code)

        headers = {
            'Content-Type': 'application/json',
        }
        resp = requests.post('http://localhost:8080/v1/AuthTokens/Promote', headers = headers)
        self.assertEqual(200, resp.status_code)

        resp = requests.post('http://localhost:8080/v1/AuthTokens/Secondary', headers = headers)
        self.assertEqual(201, resp.status_code)

        resp = requests.get('http://localhost:8080/_scenarios', headers = headers)
        self.assertEqual(200, resp.status_code)

if __name__ == '__main__':
    unittest.main()
