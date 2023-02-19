import unittest
import requests
import time
import json

proxy_servers = {
    'http': 'http://localhost:9000',
    '': 'http://localhost:9000',
}

class FixturesTest(unittest.TestCase):
    def test_record_play_fixtures_via_proxy(self):
        headers = {
            'Content-Type': 'application/yaml',
        }
        data = open('../fixtures/lines.txt', 'r').read()
        resp = requests.post('http://localhost/_fixtures/GET/lines.txt/devices', data = data, headers = headers, proxies = proxy_servers, verify = False)
        self.assertEqual(200, resp.status_code)

        headers = {
            'Content-Type': 'application/yaml',
        }
        data = open('../fixtures/props.yaml', 'r').read()
        resp = requests.post('http://localhost/_fixtures/GET/props.yaml/devices', data = data, headers = headers, proxies = proxy_servers, verify = False)
        self.assertEqual(200, resp.status_code)

        headers = {
            'Content-Type': 'application/yaml',
        }
        data = open('../fixtures/devices.yaml', 'r').read()
        resp = requests.post('http://localhost/_scenarios', data = data, headers = headers, proxies = proxy_servers, verify = False)
        self.assertEqual(200, resp.status_code)

        headers = {
            'Content-Type': 'application/json',
        }
        resp = requests.get('http://localhost/devices?page=1&pageSize=55', headers = headers, proxies = proxy_servers, verify = False)
        self.assertTrue(resp.status_code == 200 or resp.status_code == 400)

    def test_record_play_image_fixture_via_proxy(self):
        headers = {
            'Content-Type': 'application/yaml',
        }
        data = open('../fixtures/mockup.png', 'rb').read()
        resp = requests.post('http://localhost/_fixtures/GET/mockup.png/images/mock_image', data = data, headers = headers, proxies = proxy_servers, verify = False)
        self.assertEqual(200, resp.status_code)

        headers = {
            'Content-Type': 'application/yaml',
        }
        data = open('../fixtures/image.yaml', 'r').read()
        resp = requests.post('http://localhost/_scenarios', data = data, headers = headers, proxies = proxy_servers, verify = False)
        self.assertEqual(200, resp.status_code)

        headers = {
            'Content-Type': 'application/json',
        }
        resp = requests.get('http://localhost/images/mock_image', headers = headers, proxies = proxy_servers, verify = False)
        self.assertEqual(200, resp.status_code)

    def test_record_play_fixtures(self):
        headers = {
            'Content-Type': 'application/yaml',
        }
        data = open('../fixtures/lines.txt', 'r').read()
        resp = requests.post('http://localhost:8000/_fixtures/GET/lines.txt/devices', data = data, headers = headers)
        self.assertEqual(200, resp.status_code)

        headers = {
            'Content-Type': 'application/yaml',
        }
        data = open('../fixtures/props.yaml', 'r').read()
        resp = requests.post('http://localhost:8000/_fixtures/GET/props.yaml/devices', data = data, headers = headers)
        self.assertEqual(200, resp.status_code)

        headers = {
            'Content-Type': 'application/yaml',
        }
        data = open('../fixtures/devices.yaml', 'r').read()
        resp = requests.post('http://localhost:8000/_scenarios', data = data, headers = headers)
        self.assertEqual(200, resp.status_code)

        headers = {
            'Content-Type': 'application/json',
        }
        resp = requests.get('http://localhost:8000/devices?page=1&pageSize=55', headers = headers)
        self.assertTrue(resp.status_code == 200 or resp.status_code == 400)

    def test_record_play_image_fixture(self):
        headers = {
            'Content-Type': 'application/yaml',
        }
        data = open('../fixtures/mockup.png', 'rb').read()
        resp = requests.post('http://localhost:8000/_fixtures/GET/mockup.png/images/mock_image', data = data, headers = headers)
        self.assertEqual(200, resp.status_code)

        headers = {
            'Content-Type': 'application/yaml',
        }
        data = open('../fixtures/image.yaml', 'r').read()
        resp = requests.post('http://localhost:8000/_scenarios', data = data, headers = headers)
        self.assertEqual(200, resp.status_code)

        headers = {
            'Content-Type': 'application/json',
        }
        resp = requests.get('http://localhost:8000/images/mock_image', headers = headers)
        self.assertEqual(200, resp.status_code)

if __name__ == '__main__':
    unittest.main()
