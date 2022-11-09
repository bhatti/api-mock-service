import unittest
import requests
import time
import json

proxy_servers = {
    'http': 'http://localhost:8081',
    'https': 'http://localhost:8081',
}

class HTTPProxyTest(unittest.TestCase):
    def test_post_record_play_via_proxy(self):
        headers = {
            'Content-Type': 'application/json',
        }
        data = {'userId': 1, 'title': 'Buy milk', 'completed': False}
        resp = requests.post('https://jsonplaceholder.typicode.com/todos', json = data, headers = headers, proxies = proxy_servers, verify = False)
        self.assertEqual(201, resp.status_code)
        todo = json.loads(resp.text)
        self.assertEqual(1, todo['userId'])
        self.assertEqual('Buy milk', todo['title'])

    def test_get_record_play_via_proxy(self):
        headers = {
            'Content-Type': 'application/json',
        }
        resp = requests.get('https://jsonplaceholder.typicode.com/todos/1', headers = headers, proxies = proxy_servers, verify = False)
        self.assertEqual(200, resp.status_code)
        todo = json.loads(resp.text)
        self.assertEqual(1, todo['userId'])
        self.assertEqual(1, todo['id'])


if __name__ == '__main__':
    unittest.main()
