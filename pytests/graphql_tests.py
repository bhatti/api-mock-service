#!/usr/bin/env python
import unittest
import requests
import time
import json

proxy_servers = {
    'http': 'http://localhost:9000',
    'https': 'http://localhost:9000',
}

class APIGraphTest(unittest.TestCase):
    def test_post_record_play_via_proxy(self):
        query = """
    query Query {
      allFilms {
        films {
          title
          director
          releaseDate
          speciesConnection {
            species {
              name
              classification
              homeworld {
                name
              }
            }
          }
        }
      }
    }
    """
        data = {"query":query, "variables":{},"operationName":"Query"}

        {"query":"query Query {\n  allFilms {\n    films {\n      title\n      director\n      releaseDate\n      speciesConnection {\n        species {\n          name\n          classification\n          homeworld {\n            name\n          }\n        }\n      }\n    }\n  }\n}\n","variables":{},"operationName":"Query"}
        headers = {
            'Content-Type': 'application/json',
        }
        resp = requests.post('https://swapi-graphql.netlify.app/.netlify/functions/index', json = data, headers = headers, proxies = proxy_servers, verify = False)

        j = json.loads(resp.text)
        self.assertTrue(len(j["data"]["allFilms"]["films"]) > 0)


if __name__ == '__main__':
    unittest.main()
