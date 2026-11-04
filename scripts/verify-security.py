#!/usr/bin/env python3
"""Small authenticated production checks; never prints credentials or session tokens."""
from http.cookiejar import CookieJar
import json
from pathlib import Path
from urllib.error import HTTPError
from urllib.request import Request, build_opener, HTTPCookieProcessor

root = Path(__file__).resolve().parents[1]
access = json.loads((root / '.deployment/admin-access.json').read_text())
base = 'https://camellwash.cloud'
opener = build_opener(HTTPCookieProcessor(CookieJar()))
checks = {}
def request(label, path, method='GET', body=None, extra=None, expected=200):
    headers = {'Origin':base}
    if body is not None: headers['Content-Type'] = 'application/json'
    if extra: headers.update(extra)
    req = Request(base+path, data=None if body is None else json.dumps(body).encode(), headers=headers, method=method)
    try: response = opener.open(req, timeout=20)
    except HTTPError as error: response = error
    with response:
        data = response.read()
        try: result = json.loads(data)
        except ValueError: result = {}
        checks[label] = {'status':response.status, 'passed':response.status==expected}
        assert response.status == expected, label + ' returned ' + str(response.status)
        return result, response.headers

request('unauthenticated_admin_rejected','/api/v1/admin/users',expected=401)
request('environment_not_public','/.env',expected=404)
request('git_not_public','/.git/HEAD',expected=404)
login, headers = request('administrator_login','/api/v1/auth/login','POST',{'email':access['email'],'password':access['password']})
cookie = headers.get('Set-Cookie','')
assert all(value in cookie for value in ['Secure','HttpOnly','SameSite=Lax','Path=/api/v1'])
checks['secure_cookie'] = {'passed':True}
csrf = login['data']['csrf_token']
user = login['data']['user']
assert user['role'] == 'admin'
request('session_restore','/api/v1/auth/session')
request('administrator_authorization','/api/v1/admin/users')
request('csrf_missing_rejected','/api/v1/me','PATCH',{},expected=403)
request('csrf_invalid_rejected','/api/v1/me','PATCH',{},extra={'X-CSRF-Token':'invalid-security-test-token'},expected=403)
request('untrusted_origin_rejected','/api/v1/me','PATCH',{},extra={'X-CSRF-Token':csrf,'Origin':'https://untrusted.example.test'},expected=403)
request('runtime_database_write','/api/v1/me','PATCH',{'first_name':user['first_name'],'last_name':user['last_name']},extra={'X-CSRF-Token':csrf})
request('financial_report_route','/api/v1/reports/dashboard')
request('document_route_authorized','/api/v1/documents/999999999',expected=404)
request('receipt_csrf_rejected','/api/v1/payments/999999999/documents/receipt','POST',{},expected=403)
request('receipt_route_authorized','/api/v1/payments/999999999/documents/receipt','POST',{},extra={'X-CSRF-Token':csrf},expected=404)
request('reception_route_authorized','/api/v1/reception/orders/999999999/document','POST',{},extra={'X-CSRF-Token':csrf},expected=404)
request('logout','/api/v1/auth/logout','POST',{},extra={'X-CSRF-Token':csrf})
request('revoked_cookie_rejected','/api/v1/auth/session',expected=401)
(root / '.deployment/security/live-security.json').write_text(json.dumps(checks,indent=2))
print(json.dumps(checks))
