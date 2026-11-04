#!/usr/bin/env python3
"""Check Camell's trusted-proxy limiter using local requests and test client addresses."""
from collections import Counter
import json
from urllib.error import HTTPError
from urllib.request import Request, urlopen

def status(client):
    request = Request('http://127.0.0.1:8090/api/v1/auth/session', headers={
        'Host':'camellwash.cloud', 'X-Real-IP':client,
        'X-Forwarded-For':client, 'X-Forwarded-Proto':'https'})
    try: response = urlopen(request, timeout=10)
    except HTTPError as error: response = error
    with response:
        return response.status

distinct = [status('198.18.10.' + str(i+1)) for i in range(20)]
shared = [status('198.18.20.1') for _ in range(8)]
result = {'distinct_clients':dict(Counter(distinct)), 'one_client_burst':dict(Counter(shared))}
print(json.dumps(result))
assert all(code == 401 for code in distinct), 'Different clients incorrectly share a rate limit'
assert 429 in shared and set(shared) <= {401,429}, 'One client must still be rate limited'
