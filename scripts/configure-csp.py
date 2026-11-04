#!/usr/bin/env python3
"""Authorize the exact Nuxt runtime/importmap scripts for the current production build."""
import base64
import hashlib
from html.parser import HTMLParser
import json
from pathlib import Path
import re
from urllib.request import Request, urlopen

class Scripts(HTMLParser):
    def __init__(self):
        super().__init__(convert_charrefs=False)
        self.current = None
        self.items = []
    def handle_starttag(self, tag, attrs):
        if tag == 'script': self.current = [dict(attrs), '']
    def handle_data(self, data):
        if self.current is not None: self.current[1] += data
    def handle_endtag(self, tag):
        if tag == 'script' and self.current is not None:
            self.items.append(self.current)
            self.current = None

expected = None
paths = ['/login', '/', '/admin', '/app', '/accountant', '/admin/users', '/print/orders/1', '/login']
for path in paths:
    request = Request('http://127.0.0.1:8090' + path, headers={'Host':'camellwash.cloud', 'X-Forwarded-Proto':'https'})
    with urlopen(request, timeout=15) as response:
        parser = Scripts()
        parser.feed(response.read().decode())
    hashes = set()
    for attrs, content in parser.items:
        if attrs.get('src') or attrs.get('type') in ['application/json', 'application/ld+json']: continue
        if attrs.get('type') == 'importmap':
            imports = json.loads(content)
            assert set(imports) == {'imports'} and set(imports['imports']) == {'#entry'}, 'Unknown importmap'
            assert re.fullmatch(r'/_nuxt/[A-Za-z0-9_.-]+\.js', imports['imports']['#entry']), 'Unknown entrypoint'
        else:
            assert content.startswith('window.__NUXT__={};window.__NUXT__.config='), 'Unknown inline executable script'
        hashes.add(base64.b64encode(hashlib.sha256(content.encode()).digest()).decode())
    assert len(hashes) == 2, 'Expected runtime configuration and importmap'
    if expected is None: expected = hashes
    assert hashes == expected, 'Inline scripts change between routes; static hashes are not suitable'
root = Path(__file__).resolve().parents[1]
path = root / 'nginx/default.conf'
source = path.read_text()
policy = "script-src 'self' " + ' '.join("'sha256-" + value + "'" for value in sorted(expected)) + ';'
updated, count = re.subn(r'script-src[^;]*;', lambda match: policy, source)
assert count == 1, 'Expected exactly one CSP script policy'
path.write_text(updated)
print(f'CSP updated with 2 exact hashes, verified across {len(paths)} responses. Validate and reload only the Camell nginx container.')
