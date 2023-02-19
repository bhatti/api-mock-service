import unittest
import requests
import time
import json

proxy_servers = {
    'http': 'http://localhost:9000',
    'https': 'http://localhost:9000',
}

class APIProxyTest(unittest.TestCase):
    def test_post_record_play_via_proxy(self):
        headers = {
            'Content-Type': 'application/json',
            'X-Mock-Url': 'https://jsonplaceholder.typicode.com/todos',
        }
        data = {'userId': 1, 'title': 'Buy milk', 'completed': False}
        resp = requests.post('http://localhost:9000/_proxy', json = data, headers = headers, proxies = proxy_servers, verify = False)
        original = resp.text
        self.assertEqual(201, resp.status_code)
        todo = json.loads(resp.text)
        self.assertEqual(1, todo['userId'])
        self.assertEqual('Buy milk', todo['title'])
        headers = {
            'Content-Type': 'application/json',
        }
        resp = requests.post('http://localhost:8000/todos', json = data, headers = headers)
        self.assertEqual(original, resp.text)

    def test_get_record_play_via_proxy(self):
        headers = {
            'Content-Type': 'application/json',
            'X-Mock-Url': 'https://jsonplaceholder.typicode.com/todos/1',
        }
        resp = requests.get('http://localhost:9000/_proxy', headers = headers, proxies = proxy_servers, verify = False)
        self.assertEqual(200, resp.status_code)
        headers = {
            'Content-Type': 'application/json',
        }
        resp = requests.get('http://localhost:8000/todos/1', headers = headers)
        todo = json.loads(resp.text)
        self.assertEqual(1, todo['userId'])
        self.assertEqual(1, todo['id'])

    def test_post_record_play(self):
        headers = {
            'Content-Type': 'application/json',
            'X-Mock-Url': 'https://jsonplaceholder.typicode.com/todos',
        }
        data = {'userId': 1, 'title': 'Buy milk', 'completed': False}
        resp = requests.post('http://localhost:8000/_proxy', json = data, headers = headers)
        original = resp.text
        self.assertEqual(201, resp.status_code)
        todo = json.loads(resp.text)
        self.assertEqual(1, todo['userId'])
        self.assertEqual('Buy milk', todo['title'])
        headers = {
            'Content-Type': 'application/json',
        }
        resp = requests.post('http://localhost:8000/todos', json = data, headers = headers)
        self.assertEqual(original, resp.text)

    def test_get_record_play(self):
        headers = {
            'Content-Type': 'application/json',
            'X-Mock-Url': 'https://jsonplaceholder.typicode.com/todos/1',
        }
        resp = requests.get('http://localhost:8000/_proxy', headers = headers)
        self.assertEqual(200, resp.status_code)
        headers = {
            'Content-Type': 'application/json',
        }
        resp = requests.get('http://localhost:8000/todos/1', headers = headers)
        todo = json.loads(resp.text)
        self.assertEqual(1, todo['userId'])
        self.assertEqual(1, todo['id'])


if __name__ == '__main__':
    unittest.main()
