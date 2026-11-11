# Camell Wash en VPS

Proyecto: `/opt/apps/camell-wash`. Dominio: `https://camellwash.cloud`. Nombre Compose: `camell-wash`.

El Nginx del VPS termina HTTPS y envia las peticiones a `127.0.0.1:8090`. Solo el servicio Nginx de Camell publica ese puerto; pertenece a la red interna y a una red bridge propia. PostgreSQL, backend y frontend conservan la red `internal: true`. La base de datos y su volumen pertenecen exclusivamente a este proyecto.

La configuracion secreta esta en `.env` (permiso 600). Los respaldos, registros de despliegue, comprobaciones y credenciales iniciales estan en `.deployment` (permiso 700), excluido de Git. Las cuentas de aplicacion no son cuentas del VPS ni de correo.

## Actualizar este proyecto

Antes de actualizar, respaldar su base de datos y configuracion. Ejecutar los comandos desde la raiz del proyecto:

```bash
cd /opt/apps/camell-wash
docker compose -f docker-compose.prod.yml config --quiet
docker compose -f docker-compose.prod.yml exec -T --interactive=false postgres pg_dump -U camell_wash -d camell_wash -Fc > .deployment/pre-update.dump
chmod 600 .deployment/pre-update.dump
docker compose -f docker-compose.prod.yml build
docker compose -f docker-compose.prod.yml up -d --wait
python3 scripts/configure-csp.py
docker compose -f docker-compose.prod.yml exec -T --interactive=false nginx nginx -t
docker compose -f docker-compose.prod.yml exec -T --interactive=false nginx nginx -s reload
docker compose -f docker-compose.prod.yml ps -a
```

El servicio `migrate` debe terminar con codigo 0 antes de iniciar el backend. Los chequeos de salud deben usar GET (`wget -O /dev/null`); `/health` no acepta HEAD.

`configure-csp.py` verifica en seis respuestas el importmap y la configuracion publica de Nuxt, y permite sus hashes exactos en la CSP. Ejecutarlo despues de cada compilacion, porque cambia el identificador de build. El script aborta ante un script inline desconocido o contenido que varie por ruta. Si cambia la arquitectura SSR, revisar la politica antes de publicar.

## Primer administrador

```bash
python3 scripts/bootstrap-admin.py --email admin@camellwash.cloud
```

Requiere Python 3 y el modulo `bcrypt` disponible en este VPS. Crea solamente el primer administrador mediante una transaccion, sin habilitar el seed de desarrollo. No restablece usuarios existentes. Guarda la contrasena aleatoria en `.deployment/admin-access.json`, permiso 600, y su hash bcrypt en PostgreSQL. La contrasena puede cambiarse desde el perfil de la aplicacion.

## Dominio y HTTPS

El registro A de `camellwash.cloud` debe apuntar a `179.198.124.54`. `www` conserva su CNAME existente y redirige por HTTPS al dominio principal; el certificado incluye ambos nombres. El virtual host propio esta en `/etc/nginx/sites-available/camellwash.cloud`; los archivos de los otros dominios no deben modificarse.

El challenge HTTP usa `/opt/apps/camell-wash/nginx/acme`, que debe ser legible por Nginx. Conservar su location HTTP al renovar:

```bash
certbot certonly --webroot -w /opt/apps/camell-wash/nginx/acme --cert-name camellwash.cloud -d camellwash.cloud -d www.camellwash.cloud --deploy-hook 'nginx -t && systemctl reload nginx'
certbot renew --cert-name camellwash.cloud --dry-run
systemctl is-active certbot.timer
```

El temporizador existente renueva los certificados. El hook de Camell valida y recarga Nginx sin reiniciar el servicio. El virtual host HTTPS usa el certificado en `/etc/letsencrypt/live/camellwash.cloud/`.

## Comprobaciones

```bash
curl -fsS https://camellwash.cloud/health
curl -fsS https://camellwash.cloud/login >/dev/null
docker compose -f docker-compose.prod.yml logs --tail 50 backend frontend nginx
```

Probar en navegador el login, una pantalla administrativa, restaurar la sesion y cerrar sesion. Comparar los otros contenedores con `baseline.json`: mismos identificadores, inicio, reinicios, puertos, montajes y salud, y mismos hashes de sus virtual hosts. No reiniciar Docker/VPS ni ejecutar limpiezas globales o `down -v`.

## Seguridad de produccion

El backend conecta con `DB_APP_USER` / `DB_APP_PASSWORD`, un rol sin superusuario, creacion de bases/roles ni DDL. PostgreSQL y las migraciones usan `DB_USER` / `DB_PASSWORD` como propietario. El servicio `db_permissions` ejecuta `scripts/grant-runtime-db.sh` despues de las migraciones y antes del backend; aplica los permisos actuales y por defecto de forma idempotente. No utilizar el propietario como credencial del backend.

Backend y frontend ejecutan con usuario sin privilegios, raiz de archivos de solo lectura, `/tmp` temporal, capacidades eliminadas y `no-new-privileges`. Los limites de memoria/procesos del Compose afectan unicamente a Camell.

El proxy del host debe SOBRESCRIBIR `X-Real-IP` y `X-Forwarded-For` con `$remote_addr`; el Nginx interno usa esa IP saneada para los limites individuales. Conservar esta frontera de confianza.

Comprobaciones locales del VPS:

```bash
python3 scripts/verify-rate-limit.py
python3 scripts/verify-security.py
```

La segunda usa las credenciales iniciales privadas de `.deployment/admin-access.json`. Si se cambia su contrasena, realizar el acceso de prueba con una cuenta autorizada; no restablecer administradores para ejecutar el chequeo.

La regresion de sesiones y bloqueo usa `TEST_DATABASE_DSN` con una base cuyo nombre debe empezar por `camell_wash_security_test_`. El test rechaza cualquier otra base antes de escribir. Nunca apuntarlo a produccion.
