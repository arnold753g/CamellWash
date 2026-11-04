#!/usr/bin/env python3
"""Create the first production administrator without enabling development seeding."""
import argparse
import bcrypt
import json
import os
from pathlib import Path
import re
import secrets
import subprocess

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--email', required=True)
args = parser.parse_args()
email = args.email.strip().lower()
if not re.fullmatch(r'[^\s@]+@[^\s@]+\.[^\s@]+', email) or len(email) > 320:
    parser.error('Correo no valido')
root = Path(__file__).resolve().parents[1]
config = dict(line.split('=', 1) for line in (root / '.env').read_text().splitlines() if '=' in line and not line.startswith('#'))
if config.get('APP_ENV') != 'production':
    raise SystemExit('Este comando requiere APP_ENV=production')
private = root / '.deployment'
private.mkdir(mode=0o700, exist_ok=True)
os.chmod(private, 0o700)
access = private / 'admin-access.json'
if access.exists():
    raise SystemExit('Ya existe el archivo de acceso; no se modifico ninguna cuenta')
password = secrets.token_urlsafe(24)
password_hash = bcrypt.hashpw(password.encode(), bcrypt.gensalt(rounds=12)).decode()
pending = private / 'admin-access.pending.json'
fd = os.open(pending, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
with os.fdopen(fd, 'w') as f:
    json.dump({'url': 'https://camellwash.cloud/login', 'email': email, 'password': password}, f, indent=2)
command = ['docker', 'compose', '--project-directory', str(root), '-f', str(root / 'docker-compose.prod.yml'), 'exec', '-T', 'postgres', 'psql', '-U', config['DB_USER'], '-d', config['DB_NAME'], '-v', 'ON_ERROR_STOP=1', '-v', 'admin_email=' + email, '-v', 'admin_hash=' + password_hash, '-qAt']
sql = """BEGIN;
LOCK TABLE users IN SHARE ROW EXCLUSIVE MODE;
INSERT INTO users (email,password_hash,first_name,last_name,role,status)
SELECT :'admin_email', :'admin_hash', 'Administrador', 'Camell Wash', 'admin', 'active'
WHERE NOT EXISTS (SELECT 1 FROM users WHERE role='admin')
RETURNING email;
COMMIT;
"""
result = subprocess.run(command, input=sql, text=True, capture_output=True)
if result.returncode or result.stdout.strip() != email:
    (private / 'bootstrap-error.log').write_text(result.stderr)
    pending.unlink()
    raise SystemExit('No se creo una cuenta nueva. Puede existir ya un administrador; revisa .deployment/bootstrap-error.log')
pending.replace(access)
print('Administrador creado. Credenciales privadas: ' + str(access))
