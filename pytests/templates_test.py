#!/usr/bin/env python
import unittest
import requests
import time
import json

proxy_servers = {
    'http': 'http://localhost:9000',
    'https': 'http://localhost:9000',
}

class TemplatesTest(unittest.TestCase):

    def test_dynamic_template_with_proxy(self):
        headers = {
            'Content-Type': 'application/yaml',
        }
        data = open('../fixtures/stripe-customer.yaml', 'r').read()
        resp = requests.post('http://localhost/_scenarios', data = data, headers = headers, proxies = proxy_servers, verify = False)
        self.assertEqual(200, resp.status_code)

        headers = {
            'Content-Type': 'application/yaml',
        }
        data = open('../fixtures/stripe-customer-failure.yaml', 'r').read()
        resp = requests.post('http://localhost/_scenarios', data = data, headers = headers, proxies = proxy_servers, verify = False)
        self.assertEqual(200, resp.status_code)

        headers = {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer sk_test_0123456789',
        }
        resp = requests.get('http://localhost/v1/customers/123/cash_balance?page=2&pageSize=55', headers = headers, proxies = proxy_servers, verify = False)
        cash = json.loads(resp.text)
        self.assertEqual(55, cash['pageSize'])
        self.assertEqual(2, cash['page'])


    def test_dynamic_template(self):
        headers = {
            'Content-Type': 'application/yaml',
        }
        data = open('../fixtures/stripe-customer.yaml', 'r').read()
        resp = requests.post('http://localhost:8000/_scenarios', data = data, headers = headers)
        self.assertEqual(200, resp.status_code)

        headers = {
            'Content-Type': 'application/yaml',
        }
        data = open('../fixtures/stripe-customer-failure.yaml', 'r').read()
        resp = requests.post('http://localhost:8000/_scenarios', data = data, headers = headers)
        self.assertEqual(200, resp.status_code)

        headers = {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer sk_test_0123456789',
        }
        resp = requests.get('http://localhost:8000/v1/customers/123/cash_balance?page=2&pageSize=55', headers = headers)
        cash = json.loads(resp.text)
        self.assertEqual(55, cash['pageSize'])
        self.assertEqual(2, cash['page'])

if __name__ == '__main__':
    unittest.main()
