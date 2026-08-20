# Guía única y prompts por módulo para desarrollar el Sistema de Gestión de Lavadero con Codex

Este es el único archivo de especificación funcional que Codex debe consultar para desarrollar el sistema. Contiene el alcance completo, las reglas generales, la continuidad visual y los prompts de ejecución por módulo. No depende de otro prompt maestro.

No se debe crear otro archivo de requisitos, alcance, memoria funcional o prompt maestro. Los documentos técnicos y de seguimiento que Codex genere sirven como evidencia del estado real del repositorio, pero no sustituyen ni contradicen esta guía.

El proyecto base ya implementado es la fuente técnica y visual: Codex debe conservar su stack, arquitectura válida, componentes, navegación, colores, tipografías y patrones de interfaz, salvo que encuentre un problema concreto de seguridad, accesibilidad o compatibilidad y documente la corrección.

## Cómo utilizarla

1. Coloca únicamente este archivo de especificación funcional en la raíz del proyecto base.
2. Ejecuta primero el Prompt 0.
3. Revisa su informe y resuelve las decisiones pendientes.
4. Ejecuta un solo módulo por vez y en orden.
5. No pegues el prompt siguiente hasta que el módulo anterior esté verificado y aprobado.
6. Sustituye únicamente los valores escritos entre corchetes.
7. No conviene enviar todos los prompts juntos: eso aumenta el alcance, mezcla dependencias y dificulta detectar regresiones.

## 1. Identificación del proyecto

- NOMBRE_DEL_PROYECTO: Sistema de Gestión de Lavadero
- RUTA_ABSOLUTA_PROYECTO_BASE: `D:\Users\arnol\Documents\Nueva Version del proyecto\Proyecto Base`
- RAMA_BASE: Pendiente; Git no se inicializará todavía por decisión expresa del usuario.
- ZONA_HORARIA_DEL_NEGOCIO: `America/La_Paz`

El proyecto base es el repositorio de trabajo. No se debe crear otro proyecto, migrar de stack ni reemplazar su arquitectura por preferencia personal. Si la ruta no corresponde a la raíz del repositorio o existen cambios ajenos que puedan entrar en conflicto, Codex debe detenerse antes de modificar archivos.

## 2. Objetivo y alcance de la versión 1

Construir sobre el proyecto base un sistema web para administrar un lavadero de vehículos. Debe centralizar usuarios, clientes, vehículos, catálogos, precios, órdenes de lavado, asignación del personal, estados, pagos e indicadores operativos.

La orden de lavado es la entidad central: puede incluir varios servicios, puede tener una o varias personas asignadas, conserva los precios usados y registra cada cambio de estado.

La versión 1 incluye:

- Autenticación segura y autorización por roles.
- Usuarios activos/inactivos.
- Clientes con uno o varios vehículos.
- Tipos de vehículo, categorías y servicios extensibles.
- Matriz de precios administrada sin cambiar código.
- Órdenes con varios servicios y varias personas asignadas.
- Precios congelados al crear la orden.
- Flujo de estados con historial completo.
- Un pago total por orden, sin anticipos ni parciales.
- Dashboard y reportes básicos.

Queda fuera de la versión 1:

- WhatsApp y otras notificaciones automáticas.
- Anticipos, saldos y pagos parciales.
- Descuentos, promociones y comisiones.
- Facturación electrónica.
- Exportaciones Excel o PDF.
- Integraciones bancarias o pasarelas de pago.
- Borrado físico de información histórica.
- Funcionalidades no descritas en este archivo.

## 3. Roles y permisos globales

Los nombres visibles serán Administrador, Encargado y Personal de lavado.

| Acción | Administrador | Encargado | Personal de lavado |
| --- | --- | --- | --- |
| Iniciar sesión si está activo | Sí | Sí | Sí |
| Gestionar usuarios y roles | Sí | No | No |
| Gestionar tipos, categorías y servicios | Sí | No | No |
| Gestionar precios | Sí | No | No |
| Registrar o editar clientes | Sí | Sí | No |
| Registrar o editar vehículos | Sí | Sí | No |
| Crear y gestionar órdenes | Sí | Sí | Solo órdenes asignadas y transiciones permitidas |
| Asignar Personal de lavado | Sí | Sí | No |
| Cancelar una orden | Sí | Sí | No |
| Registrar el pago | Sí | Sí | No |
| Ver reportes operativos | Sí | Sí, sujeto a confirmación | Solo información propia si se aprueba |
| Ver reportes financieros | Sí | Sujeto a confirmación | No |

La autorización debe verificarse siempre en el backend. Ocultar enlaces o botones no sustituye el control de permisos. El Personal de lavado no puede consultar órdenes ajenas manipulando identificadores, filtros, rutas o solicitudes.

## 4. Contrato obligatorio con el proyecto base

Codex debe comenzar cada módulo comprendiendo el área equivalente del proyecto base y reutilizar lo que ya funciona.

### 4.1 Reglas técnicas

- No asumir lenguajes, frameworks, ORM, base de datos, mecanismos de autenticación ni comandos.
- Identificar las versiones y convenciones reales del repositorio.
- Conservar la separación de capas, nomenclatura, manejo de errores, formato de API y organización existentes cuando sean correctos.
- Reutilizar utilidades y componentes antes de crear duplicados.
- No introducir otra librería de interfaz, gestión de estado, validación, fechas o HTTP si el proyecto ya tiene una solución suficiente.
- No migrar de sesiones a JWT, de cookies a Bearer, de REST a GraphQL, de un ORM a otro ni de un framework a otro sin autorización expresa.
- Si una pieza base es insegura o incompatible, documentar evidencia, impacto, alternativa y pruebas antes de sustituirla.
- No copiar secretos ni agregar credenciales reales al repositorio.

### 4.2 Continuidad visual obligatoria

El sistema de lavadero debe verse como una evolución natural del proyecto base, no como una aplicación diferente.

Antes de implementar la primera pantalla, Codex debe inventariar y documentar:

- Tipografías y jerarquías de texto.
- Colores primarios, secundarios, neutros, estados y fondos.
- Variables, tokens o tema centralizado.
- Layout principal, sidebar, header, menús y breadcrumbs.
- Anchos máximos, grillas, contenedores y breakpoints.
- Espaciado, bordes, radios, sombras e iconografía.
- Botones, inputs, selects, tablas, tarjetas, modales, alertas, badges, tabs, paginación y notificaciones.
- Estados hover, focus, active, disabled, loading, empty y error.
- Comportamiento responsive en escritorio, tablet y móvil.

Reglas visuales:

- Reutilizar componentes, clases, tokens y layouts existentes.
- No hardcodear repetidamente colores, fuentes o medidas si existe un sistema de tema.
- No cambiar logo, paleta, fuente, sidebar, header o navegación global salvo solicitud expresa.
- No agregar estilos genéricos que sobrescriban el proyecto completo.
- Toda pantalla nueva debe respetar densidad, alineación, jerarquía y patrones de interacción existentes.
- Mantener consistencia entre listados, formularios y detalles de todos los módulos.
- Incluir estados de carga, vacío, error, permiso denegado y confirmación.
- No usar solo el color para comunicar un estado; acompañarlo con texto o icono.
- Mantener contraste, foco visible, etiquetas y navegación por teclado según la capacidad del stack.
- La vista del Personal de lavado debe funcionar correctamente en móvil.

Verificación visual por módulo:

- Comparar cada pantalla nueva con al menos una pantalla equivalente del proyecto base.
- Probar los breakpoints que el proyecto ya utiliza.
- Revisar desbordamientos, tablas, formularios, modales y textos largos.
- Si existe herramienta de preview o pruebas de interfaz, usarla y documentar evidencia.
- No declarar continuidad visual sin inspección real del resultado.

## 5. Principios de datos y negocio

- Los importes se almacenan con un tipo decimal apto para dinero, nunca con punto flotante binario.
- Precio, total, estado, rol, encargado y cobrador son campos protegidos y validados por el servidor.
- Las operaciones sobre varias tablas deben ser transaccionales.
- Los registros usados históricamente se desactivan; no se borran.
- Las claves foráneas y reglas `ON DELETE` deben ser explícitas y evitar cascadas destructivas.
- Los timestamps se almacenan de manera consistente y se presentan en la zona horaria configurada.
- Login, CI, teléfono y placa deben normalizarse con reglas documentadas antes de verificar duplicados.
- Los listados no triviales deben tener filtros, límites o paginación.
- Evitar consultas N+1 en historiales, órdenes y reportes.
- No inventar reglas. Toda ambigüedad que afecte datos, seguridad, permisos o dinero se consulta antes de implementar.

## 6. Modelo lógico completo

Codex debe adaptar nombres y tipos a las convenciones reales del proyecto base; este modelo expresa relaciones y reglas, no impone nombres SQL concretos.

| Entidad | Datos y relaciones esenciales | Integridad mínima |
| --- | --- | --- |
| Usuario | nombre, login, credencial, rol, activo, timestamps | login único; baja lógica |
| Cliente | nombre, CI, teléfono, timestamps | política de CI por confirmar |
| TipoVehiculo | nombre, activo, timestamps | nombre único; baja lógica |
| Vehiculo | cliente, tipo, placa, marca, modelo, color, timestamps | placa única según política aprobada |
| CategoriaServicio | nombre, activo, timestamps | nombre único; baja lógica |
| Servicio | categoría, nombre, activo, timestamps | único dentro de categoría; baja lógica |
| PrecioServicio | servicio, tipo de vehículo, precio, timestamps | combinación única; decimal no negativo |
| Orden | vehículo, encargado, ingreso, entrega estimada, estado, total, timestamps | total calculado; estado controlado |
| OrdenServicio | orden, servicio, precio congelado y snapshot aprobado | sin duplicar servicio |
| OrdenPersonal | orden, usuario | sin duplicar asignación; rol válido |
| HistorialEstado | orden, estado anterior/nuevo, instante, usuario, motivo | solo anexar; transaccional |
| Pago | orden, monto, método, instante, cobrador y auditoría aprobada | una orden/un pago; monto exacto |

## 7. Flujo global de órdenes

`PENDIENTE → EN_PROCESO → CONTROL_CALIDAD → TERMINADO → ENTREGADO`

`CANCELADO` es un estado terminal alternativo. No se permiten saltos arbitrarios, retrocesos ni reapertura en la versión 1.

Propuesta de permisos que debe confirmarse antes del Módulo 5:

| Transición | Administrador | Encargado | Personal asignado |
| --- | --- | --- | --- |
| Pendiente → En proceso | Sí | Sí | Sí |
| En proceso → Control de calidad | Sí | Sí | Sí |
| Control de calidad → Terminado | Sí | Sí | No |
| Terminado → Entregado | Sí | Sí | No |
| Cancelación desde estados aprobados | Sí | Sí | No |

## 8. Seguridad mínima transversal

Codex debe verificar y probar, de acuerdo con el stack existente:

- Hash seguro de contraseñas mediante biblioteca mantenida.
- Expiración, revocación y almacenamiento seguro de sesiones o tokens.
- Cookies HttpOnly, Secure en producción y SameSite adecuado si se usan cookies.
- Protección CSRF cuando aplique.
- Rate limiting o protección razonable del inicio de sesión.
- Autorización por rol y por asignación en todos los endpoints.
- Prevención de IDOR y asignación masiva.
- Consultas parametrizadas u ORM bien utilizado.
- Validación y normalización en servidor.
- CORS limitado a orígenes configurados.
- Secretos solo mediante variables de entorno o mecanismo seguro equivalente.
- `.env.example` sin valores reales.
- Logs sin contraseñas, tokens, cookies ni CI completo.
- Errores de producción sin stack traces ni SQL.
- Seeds de desarrollo deshabilitados por defecto en producción.
- Transacciones, idempotencia y concurrencia en órdenes y pagos.

Compilar no demuestra seguridad. Cada control debe tener evidencia en código, configuración o prueba.

## 9. Reglas de ejecución, documentación y Git

- Trabajar un módulo por vez.
- Leer las instrucciones aplicables y el estado actual antes de editar.
- Preservar cambios preexistentes del usuario.
- No usar comandos destructivos para limpiar el repositorio.
- No implementar el módulo siguiente por anticipado.
- Añadir pruebas al mismo tiempo que la funcionalidad.
- Ejecutar build, tests, formato, lint, análisis estático y chequeo de tipos disponibles.
- Revisar el diff completo y `git diff --check` o equivalente.
- Actualizar `tasks.md`, `PROGRESS.md`, `lessons.md` y la documentación existente sin crear fuentes duplicadas.
- No hacer commit, merge, rebase, push, despliegue ni eliminar ramas salvo autorización expresa.
- Si una verificación falla, corregirla antes de declarar el módulo concluido.
- Si el entorno impide una prueba, declarar NO EJECUTADA con causa; nunca presentarla como aprobada.

Un módulo está terminado únicamente si su backend, interfaz, permisos, migraciones, pruebas, continuidad visual y documentación coinciden con los requisitos reales.

## 10. Orden obligatorio de desarrollo

| Orden | Módulo | Dependencias |
| --- | --- | --- |
| 0 | Auditoría del proyecto base y contrato visual | Ninguna |
| 1 | Usuarios, roles y autenticación | Fase 0 aprobada |
| 2 | Tipos de vehículo, categorías y servicios | Módulo 1 |
| 3 | Matriz de precios | Módulo 2 |
| 4 | Clientes y vehículos | Módulos 1 y 2 |
| 5 | Órdenes, asignaciones e historial | Módulos 1 a 4 |
| 6 | Pagos | Módulo 5 |
| 7 | Dashboard y reportes | Módulos 5 y 6 |
| 8 | Revisión integral y cierre | Módulos 1 a 7 |

Las notificaciones quedan como módulo futuro fuera de esta guía.

## Prompt 0. Auditoría, preparación y plan del proyecto

Actúa como arquitecto y desarrollador principal. Inicia únicamente la Fase 0 del Sistema de Gestión de Lavadero.

Datos del trabajo:

- Nombre del proyecto: Sistema de Gestión de Lavadero
- Raíz del proyecto base existente: `D:\Users\arnol\Documents\Nueva Version del proyecto\Proyecto Base`
- Rama base declarada: pendiente; Git no se inicializará todavía por decisión expresa del usuario.
- Zona horaria del negocio: `America/La_Paz`

Antes de actuar:

1. Lee por completo este archivo. Es la única especificación funcional del sistema.
2. Lee todos los `AGENTS.md` aplicables y la documentación existente del repositorio.
3. Comprueba las rutas y el estado de Git.
4. Confirma que estás situado en la raíz del proyecto base existente y que es el repositorio que se desarrollará.
5. Si faltan archivos, la ruta es ambigua o hay cambios preexistentes que puedan entrar en conflicto, detente y explica el bloqueo.

Objetivo de esta fase:

- Entender el proyecto base antes de reutilizarlo.
- Determinar qué se reutiliza, adapta, sustituye o descarta.
- Definir la arquitectura objetivo y el plan modular.
- Resolver o enumerar las decisiones funcionales pendientes.

Inspecciona con evidencia real:

- Estructura del repositorio e instrucciones activas.
- Lenguajes, frameworks, versiones y gestores de dependencias.
- Arquitectura de backend y frontend.
- Autenticación, sesiones o tokens, autorización y usuarios existentes.
- Base de datos, migraciones, seeds, transacciones y restricciones.
- Validación, errores, logs, configuración y variables de entorno.
- Contratos de API, componentes visuales, tema y navegación.
- Tests, build, lint, análisis estático, Docker y CI/CD si existen.
- Dependencias inseguras, código acoplado al negocio anterior y deuda técnica relevante.
- Archivos sensibles o generados que deben permanecer fuera de Git: secretos, `.env`, dependencias instaladas, binarios, cachés, compilaciones, logs, uploads y bases de datos.
- Sistema visual real: tokens, fuentes, paleta, layouts, componentes, navegación, responsive y accesibilidad.

Entregables permitidos en esta fase:

- `docs/00-auditoria-proyecto-base.md`
- `docs/01-arquitectura-objetivo.md`
- `docs/02-modelo-datos-inicial.md`
- `docs/03-plan-modular.md`
- `tasks.md`
- `PROGRESS.md`
- `lessons.md`

Si ya existen documentos equivalentes, actualízalos sin duplicar fuentes de verdad. No copies el contenido funcional completo de esta guía dentro de esos documentos: referencia esta guía y registra únicamente decisiones técnicas, evidencia, avance, riesgos y verificaciones.

La auditoría debe incluir una tabla por componente con: evidencia, decisión `REUTILIZAR`/`ADAPTAR`/`SUSTITUIR`/`NO USAR`, justificación, riesgo y acción propuesta.

Incluye además un contrato visual documentado con:

- Componentes y pantallas del proyecto base que servirán de referencia.
- Tipografías, colores, espaciado, radios, sombras y breakpoints reales.
- Reglas de reutilización para formularios, tablas, tarjetas, modales, navegación y feedback.
- Evidencia visual disponible y método de verificación que se repetirá en cada módulo.
- Prohibición de introducir un segundo lenguaje visual o una nueva librería UI sin autorización.

Presenta explícitamente estas decisiones pendientes con recomendación razonada:

- CI obligatorio y único.
- Tratamiento de vehículos sin placa.
- Escritura comercial de “Raibanizada”.
- Precio cero permitido o bloqueado.
- Permisos exactos de transición del Personal de lavado.
- Quién aprueba Control de calidad.
- Momento y métodos del pago.
- Corrección o anulación de pagos erróneos.
- Acceso del Encargado a reportes financieros.
- Zona horaria operativa.

Verificación:

- Ejecuta solo comprobaciones no destructivas.
- No implementes módulos ni cambies lógica funcional.
- No hagas commit, merge, rebase, push ni elimines ramas.
- Revisa el diff final y confirma que solo contiene documentación autorizada.

Informe final obligatorio:

1. Resultado de la auditoría.
2. Inventario técnico.
3. Tabla de reutilización.
4. Riesgos de seguridad y compatibilidad.
5. Decisiones resueltas y pendientes.
6. Archivos creados o modificados.
7. Comandos ejecutados y resultados.
8. Estado de Git.
9. Plan exacto propuesto para el Módulo 1.

Al terminar, detente. No inicies el Módulo 1 hasta recibir autorización expresa.

Condición para continuar: la Fase 0 está aprobada, las rutas están definidas y no quedan decisiones bloqueantes para usuarios/autenticación.

## Prompt 1. Usuarios, roles y autenticación

Rama propuesta: `feat/modulo-01-usuarios-roles`

Implementa únicamente el Módulo 1: Usuarios, roles y autenticación del Sistema de Gestión de Lavadero.

Rama esperada: `feat/modulo-01-usuarios-roles`. Si no existe y el worktree está limpio, créala desde la rama base aprobada en la Fase 0. Si la rama base es ambigua o hay cambios que impiden cambiar de rama con seguridad, detente y reporta el estado sin limpiar trabajo ajeno.

Fuente de verdad y contexto obligatorio:

- Lee por completo este archivo único y aplica sus secciones generales, alcance, permisos, contrato visual, seguridad y reglas de ejecución.
- Lee `docs/00-auditoria-proyecto-base.md`, `docs/01-arquitectura-objetivo.md`, `docs/02-modelo-datos-inicial.md` y `docs/03-plan-modular.md`.
- Lee `AGENTS.md`, `tasks.md`, `PROGRESS.md` y `lessons.md` aplicables.
- Verifica que la Fase 0 esté concluida y aprobada.
- Inspecciona el estado de Git antes de modificar archivos. Preserva cambios del usuario.
- Antes de crear interfaz, identifica las pantallas y componentes equivalentes del proyecto base. Reutiliza su layout, tema, tipografía, paleta, controles y patrones responsive; no introduzcas un segundo estilo visual.

Alcance exclusivo:

- Autenticación y cierre de sesión.
- Consulta del usuario autenticado.
- Usuarios con nombre completo, login único, contraseña segura, rol y estado activo/inactivo.
- Roles: Administrador, Encargado y Personal de lavado.
- Gestión de usuarios reservada al Administrador.
- Autorización backend por rol.
- Creación segura y controlada del primer Administrador.

Reglas obligatorias:

- Solo el Administrador crea, edita, activa/desactiva usuarios y asigna roles.
- Encargado y Personal de lavado no acceden a administración de usuarios, aunque manipulen rutas o solicitudes.
- Un usuario inactivo no puede iniciar sesión ni renovar o conservar acceso válido según el mecanismo existente.
- Desactivar un usuario no elimina su identidad ni rompe referencias históricas futuras.
- El login se normaliza y valida según una regla documentada y tiene restricción única en la base de datos.
- Las contraseñas se procesan con una biblioteca segura y mantenida. Nunca se almacenan o registran en texto plano.
- No exponer hashes, tokens, secretos, cookies ni detalles internos en API o logs.
- Reutiliza la autenticación del proyecto base solo si la auditoría la declaró segura y compatible.
- No cambies sesiones por JWT, cookies por Bearer ni la arquitectura de autenticación sin una necesidad demostrada.
- Si usa cookies, aplica HttpOnly, Secure en producción, SameSite adecuado y protección CSRF cuando corresponda.
- Si usa refresh tokens, garantiza expiración, rotación o revocación de acuerdo con la arquitectura aprobada.
- Añade protección razonable contra fuerza bruta en el inicio de sesión.
- Los seeds de desarrollo deben exigir una bandera explícita y quedar bloqueados en producción.

Implementación esperada:

1. Migraciones y restricciones de integridad.
2. Modelos/repositorios o capa de persistencia.
3. Casos de uso/servicios con reglas de autorización.
4. Endpoints y validación.
5. Interfaz de inicio de sesión y administración de usuarios.
6. Guards/middleware de navegación sin tratarlos como sustituto de la autorización backend.
7. Estados de carga, error, vacío y permiso denegado.
8. Tests y documentación.

Pruebas mínimas:

- Login correcto e incorrecto.
- Login inexistente sin filtrar si el usuario existe.
- Usuario inactivo rechazado.
- Login duplicado bloqueado por aplicación y base de datos.
- Contraseña nunca devuelta ni registrada.
- Administrador puede gestionar usuarios.
- Encargado y Personal reciben permiso denegado en endpoints administrativos.
- Sesión o token expirado/revocado se rechaza.
- Seed de desarrollo se niega sin confirmación y en producción.
- Migraciones funcionan desde una base vacía.
- Build, lint, formato, análisis estático y tests del repositorio pasan.

Límites:

- No implementes clientes, vehículos, catálogos, precios, órdenes, pagos ni reportes.
- No agregues permisos genéricos que anticipen módulos sin necesidad técnica.
- No hagas refactorizaciones amplias fuera del módulo.
- No hagas commit, merge, rebase, push ni elimines ramas.

Documentación y seguimiento:

- Actualiza `tasks.md`, `PROGRESS.md`, `lessons.md` y la especificación del Módulo 1.
- Registra decisiones técnicas reales y comandos de verificación.
- Si una decisión bloqueante aparece, detente antes de crear una solución inventada.

Informe final:

- Resultado realmente terminado.
- Archivos cambiados y propósito.
- Migraciones creadas.
- Endpoints/pantallas implementados.
- Matriz de permisos comprobada.
- Tabla de pruebas y comandos con resultado.
- Riesgos o pendientes.
- Estado de Git.
- Propuesta concreta para el Módulo 2.

Al terminar, detente. No implementes el Módulo 2.

Criterios de aceptación del Módulo 1:

- Los tres roles existen y se aplican en backend.
- Solo el Administrador gestiona usuarios.
- Un usuario inactivo no obtiene acceso.
- Las credenciales y sesiones están protegidas.
- Migraciones, tests y build pasan.
- El módulo está documentado y no contiene trabajo adelantado.

## Prompt 2. Tipos de vehículo, categorías y servicios

Rama propuesta: `feat/modulo-02-catalogos`

Implementa únicamente el Módulo 2: catálogos de tipos de vehículo, categorías de servicio y servicios.

Rama esperada: `feat/modulo-02-catalogos`. Si no existe y el worktree está limpio, créala desde el punto de integración aprobado del Módulo 1. Si no puedes determinarlo con evidencia, detente; no adivines la rama base.

Contexto obligatorio:

- Lee por completo este archivo único y todos los `AGENTS.md` aplicables.
- Lee la auditoría, arquitectura, modelo de datos, plan modular, `tasks.md`, `PROGRESS.md` y `lessons.md`.
- Verifica con evidencia que el Módulo 1 esté concluido y que autenticación/autorización funcionen.
- Revisa el estado de Git y preserva cambios preexistentes.
- Implementa las pantallas con los componentes administrativos existentes del proyecto base. Conserva exactamente su jerarquía visual, navegación, formularios, tablas, botones, feedback y responsive.

Resuelve antes de implementar seeds:

- Confirma la escritura comercial exacta de “Raibanizada”.
- Confirma los valores iniciales. Propuesta actual:
  - Tipos: Moto, Torito/Mototaxi, Automóvil, SUV/Camioneta, Furgoneta y Camión.
  - Categorías: Lavado y Detailing.
  - Servicios de Lavado: Chasis, Básico y Premium con aspirado.
  - Servicios de Detailing: Raibanizada Americana, Raibanizada China y PPF.

Si esta decisión sigue pendiente, implementa el catálogo sin inventar seeds comerciales y reporta el bloqueo parcial.

Alcance exclusivo:

- CRUD administrativo con baja lógica de tipos de vehículo.
- CRUD administrativo con baja lógica de categorías.
- CRUD administrativo con baja lógica de servicios vinculados a una categoría.
- Listados activos para futuros formularios operativos.
- Búsqueda, filtros y paginación según las convenciones del proyecto.

Reglas obligatorias:

- Solo el Administrador crea, edita, activa o desactiva catálogos.
- Encargado y Personal de lavado pueden consultar únicamente los catálogos activos que necesiten, pero no administrarlos.
- Nombre de tipo de vehículo único después de normalización acordada.
- Nombre de categoría único después de normalización acordada.
- Nombre de servicio único dentro de su categoría, con restricción en base de datos.
- No realizar borrado físico de registros con valor histórico.
- Un elemento inactivo no aparece en selectores operativos nuevos.
- La vista administrativa sí permite filtrar activos e inactivos.
- No codificar catálogos en enums cerrados si deben ser extensibles.
- Los seeds deben ser idempotentes, seguros y compatibles con el ambiente.

Implementación esperada:

1. Migraciones, claves foráneas, índices y restricciones.
2. Persistencia y reglas de baja lógica.
3. Endpoints administrativos y endpoints de consulta activa.
4. Validación y normalización consistente.
5. Pantallas administrativas reutilizando el diseño del proyecto base.
6. Confirmaciones antes de desactivar y mensajes claros cuando existan dependencias.
7. Tests y documentación.

Pruebas mínimas:

- Administrador gestiona cada catálogo.
- Encargado y Personal no modifican catálogos manipulando la API.
- Duplicados normalizados son rechazados.
- Dos servicios con el mismo nombre en la misma categoría son rechazados.
- La regla para nombres iguales en categorías distintas queda documentada y probada.
- Elementos inactivos no aparecen en listados operativos.
- Desactivar no elimina registros.
- Seeds pueden ejecutarse más de una vez sin duplicar datos.
- Migraciones desde base vacía, tests, build, lint y análisis estático pasan.

Límites:

- No implementes todavía precios, clientes, vehículos, órdenes, pagos o reportes.
- No agregues una matriz de precios incompleta dentro de este módulo.
- No hagas cambios de autenticación salvo corrección imprescindible y documentada.
- No hagas commit, merge, rebase, push ni elimines ramas.

Actualiza `tasks.md`, `PROGRESS.md`, `lessons.md` y la documentación del módulo con evidencia real.

Informe final:

- Resultado.
- Catálogos y reglas implementadas.
- Seeds definitivos o decisión pendiente.
- Archivos y migraciones.
- Endpoints y pantallas.
- Pruebas ejecutadas y resultados.
- Riesgos y pendientes.
- Estado de Git.
- Plan exacto del Módulo 3.

Al terminar, detente. No implementes el Módulo 3.

Criterios de aceptación del Módulo 2:

- Los tres catálogos son extensibles y administrables.
- La baja es lógica y conserva referencias.
- Las restricciones de unicidad funcionan en aplicación y base de datos.
- Los permisos se validan en backend.
- Los valores iniciales no fueron inventados.

## Prompt 3. Matriz de precios

Rama propuesta: `feat/modulo-03-precios`

Implementa únicamente el Módulo 3: matriz de precios por servicio y tipo de vehículo.

Rama esperada: `feat/modulo-03-precios`. Si no existe y el worktree está limpio, créala desde el punto de integración aprobado del Módulo 2. Si no puedes determinarlo con evidencia, detente.

Contexto y precondiciones:

- Lee por completo este archivo único, `AGENTS.md`, documentación técnica, `tasks.md`, `PROGRESS.md` y `lessons.md`.
- Verifica que los Módulos 1 y 2 estén concluidos y probados.
- Revisa el estado de Git y no dañes cambios del usuario.
- Confirma antes de migrar si se permite precio cero. Recomendación: precio estrictamente mayor que cero salvo servicios gratuitos reales.
- Usa la tabla, inputs monetarios, filtros, botones y patrones de feedback existentes. No agregues una nueva librería UI ni estilos paralelos para la matriz.

Alcance exclusivo:

- Matriz servicio × tipo de vehículo.
- Creación y actualización administrativa de un precio por combinación.
- Consulta filtrada de precios.
- Detección de combinaciones activas sin precio.
- Interfaz administrativa de matriz con faltantes visibles.

Reglas obligatorias:

- La combinación servicio/tipo de vehículo es única en aplicación y base de datos.
- Usa un tipo decimal apto para dinero; nunca `float`/`double` binario.
- Solo el Administrador modifica precios.
- Encargado puede consultar precios activos necesarios para la operación futura, sin modificarlos.
- Personal de lavado no accede a administración ni a información financiera innecesaria.
- No se crean precios para servicios o tipos inexistentes.
- Define y prueba qué ocurre con combinaciones inactivas sin borrar información histórica.
- Crear o reactivar un servicio/tipo debe reflejar inmediatamente combinaciones faltantes.
- La API no confía en identificadores o importes sin validación.
- Implementa control de concurrencia o upsert seguro según la arquitectura aprobada para evitar duplicados.

Interfaz esperada:

- Tabla/matriz comprensible por tipo y servicio.
- Filtros por categoría, servicio, tipo y estado si son necesarios.
- Celdas faltantes claramente identificadas, no solo mediante color.
- Edición validada con confirmación o feedback inequívoco.
- Estados de carga, error y vacío.

Preparación para órdenes:

- Implementa una función/caso de uso backend confiable que obtenga el precio vigente de una combinación.
- No implementes órdenes ni copies todavía el precio a una orden.
- Deja documentado el contrato que usará el Módulo 5 para congelar el precio.

Pruebas mínimas:

- Administrador crea y actualiza un precio.
- Encargado y Personal no modifican precios.
- Combinación duplicada es rechazada incluso con solicitudes simultáneas.
- Precio negativo, formato inválido y precio cero según decisión son rechazados.
- Faltantes se calculan correctamente al agregar/reactivar catálogo.
- Elementos inactivos no se ofrecen como combinaciones nuevas.
- La consulta devuelve el precio decimal sin pérdida de precisión.
- Migraciones, tests, build, lint, formato y análisis estático pasan.

Límites:

- No implementes órdenes, clientes, vehículos, pagos ni reportes.
- No hardcodees precios comerciales de ejemplo como datos reales.
- No hagas cambios extensos a catálogos o autenticación.
- No hagas commit, merge, rebase, push ni elimines ramas.

Actualiza `tasks.md`, `PROGRESS.md`, `lessons.md` y la documentación del módulo.

Informe final:

- Resultado funcional.
- Modelo y restricciones.
- Endpoints/pantallas.
- Manejo de precisión y concurrencia.
- Tabla de pruebas y comandos.
- Faltantes o riesgos.
- Estado de Git.
- Plan del Módulo 4.

Al terminar, detente. No implementes el Módulo 4.

Criterios de aceptación del Módulo 3:

- Existe una sola tarifa vigente por combinación.
- Los importes no pierden precisión.
- Los huecos de la matriz son visibles.
- Solo el Administrador modifica precios.
- El contrato para calcular precios en órdenes queda probado y documentado.

## Prompt 4. Clientes y vehículos

Rama propuesta: `feat/modulo-04-clientes-vehiculos`

Implementa únicamente el Módulo 4: clientes y vehículos.

Rama esperada: `feat/modulo-04-clientes-vehiculos`. Si no existe y el worktree está limpio, créala desde el punto de integración aprobado del Módulo 3. Si no puedes determinarlo con evidencia, detente.

Contexto y precondiciones:

- Lee por completo este archivo único, `AGENTS.md`, auditoría, arquitectura, modelo de datos, plan modular, `tasks.md`, `PROGRESS.md` y `lessons.md`.
- Verifica que los Módulos 1, 2 y 3 estén concluidos.
- Inspecciona Git y preserva cambios preexistentes.
- Reutiliza los patrones del proyecto base para listados, búsqueda, formularios, detalle y navegación relacionada. Verifica escritorio y móvil con el mismo lenguaje visual.

Decisiones obligatorias antes de crear restricciones:

- ¿El CI es siempre obligatorio?
- ¿El CI debe ser único después de normalización?
- ¿Cómo se registrará un vehículo que todavía no tiene placa?

Recomendación: CI obligatorio y único; para vehículos sin placa usar un identificador temporal controlado, nunca una placa falsa. Si el usuario no confirma, detente antes de fijar una migración irreversible.

Alcance exclusivo:

- Clientes: nombre completo, CI y teléfono.
- Vehículos: propietario, tipo de vehículo, placa o mecanismo temporal aprobado, marca, modelo y color.
- Relación un cliente a muchos vehículos sin límite fijo.
- Búsqueda, filtros, paginación y detalle.
- Edición autorizada por Administrador y Encargado.

Reglas obligatorias:

- Administrador y Encargado crean y editan clientes/vehículos.
- Personal de lavado no administra clientes ni vehículos.
- Normaliza CI, teléfono y placa con reglas explícitas y probadas.
- No uses valores ficticios para satisfacer campos obligatorios.
- La placa es única cuando existe una placa real, según la decisión aprobada.
- Solo se seleccionan tipos de vehículo activos para registros nuevos.
- Un tipo desactivado sigue visible en registros históricos existentes.
- Cambiar propietario debe estar prohibido o auditado según la decisión documentada; no lo permitas silenciosamente.
- Evita exponer CI y teléfono en listados o respuestas donde no son necesarios.
- La autorización se valida en backend.

Historial:

- El historial completo depende de las órdenes del Módulo 5.
- En este módulo implementa la estructura de navegación y contratos mínimos solo si no generan datos ficticios ni endpoints falsos.
- Deja el historial de servicios marcado como pendiente del Módulo 5 y complétalo allí.

Implementación esperada:

1. Migraciones, claves foráneas, índices y restricciones.
2. Persistencia y reglas de normalización.
3. API de clientes y vehículos.
4. Pantallas de listado, búsqueda, alta, edición y detalle.
5. Vista clara de varios vehículos por cliente.
6. Validaciones consistentes en frontend y backend.
7. Tests y documentación.

Pruebas mínimas:

- Un cliente admite varios vehículos.
- Reglas de CI obligatorio/único según decisión.
- Placa duplicada normalizada es rechazada.
- Vehículo sin placa se comporta exactamente según la política aprobada.
- Tipo inactivo no puede usarse en un vehículo nuevo.
- Nombre y teléfono obligatorios se validan conforme a las reglas aprobadas.
- Administrador y Encargado operan; Personal recibe permiso denegado.
- Búsqueda no filtra datos indebidos y respeta paginación.
- Datos inválidos o normalizados de forma equivalente no generan duplicados.
- Migraciones, tests, build, lint y análisis estático pasan.

Límites:

- No implementes órdenes, pagos o reportes.
- No inventes historial de servicios antes del Módulo 5.
- No amplíes el perfil del cliente con campos no solicitados.
- No hagas commit, merge, rebase, push ni elimines ramas.

Actualiza `tasks.md`, `PROGRESS.md`, `lessons.md` y la documentación del módulo.

Informe final:

- Resultado y decisiones adoptadas.
- Reglas de normalización y privacidad.
- Migraciones, endpoints y pantallas.
- Pruebas y comandos con resultado.
- Funciones diferidas al Módulo 5.
- Riesgos y estado de Git.
- Plan del Módulo 5.

Al terminar, detente. No implementes el Módulo 5.

Criterios de aceptación del Módulo 4:

- Un cliente puede tener múltiples vehículos.
- CI y placa se gestionan con reglas reales, no datos ficticios.
- Los permisos y la privacidad están protegidos en backend.
- El historial dependiente de órdenes está explícitamente diferido.

## Prompt 5. Órdenes de lavado, asignaciones e historial

Rama propuesta: `feat/modulo-05-ordenes`

Implementa únicamente el Módulo 5: órdenes de lavado, servicios solicitados, personal asignado e historial de estados. Este es el núcleo del sistema.

Rama esperada: `feat/modulo-05-ordenes`. Si no existe y el worktree está limpio, créala desde el punto de integración aprobado del Módulo 4. Si no puedes determinarlo con evidencia, detente.

Contexto y precondiciones:

- Lee por completo este archivo único y todos los `AGENTS.md` aplicables.
- Lee auditoría, arquitectura, modelo de datos, plan modular, `tasks.md`, `PROGRESS.md` y `lessons.md`.
- Verifica que los Módulos 1 a 4 estén concluidos y que sus pruebas pasen.
- Revisa el estado de Git y preserva cambios ajenos.
- La vista operativa debe extender el diseño existente del proyecto base. Reutiliza tarjetas, tablas, badges, modales, formularios y navegación; no rediseñes la aplicación. Prioriza la experiencia móvil del Personal de lavado.

Decisiones que deben estar resueltas:

- Personal asignado: Pendiente → En proceso y En proceso → Control de calidad.
- Administrador/Encargado: Control de calidad → Terminado y Terminado → Entregado.
- Cancelación reservada a Administrador/Encargado con motivo.
- Confirmar si el nombre del servicio también se congela como snapshot además del precio.
- Confirmar hasta qué estado pueden editarse vehículo, servicios, fecha estimada y personal asignado.
- Confirmar la zona horaria operativa.

Si falta una decisión que cambia datos o permisos, pregunta antes de implementarla.

Modelo funcional:

- Una orden pertenece a un vehículo y conserva el encargado que la registró.
- Una orden tiene uno o varios servicios.
- Cada línea conserva el precio congelado obtenido del backend.
- Una orden tiene una o varias personas de lavado activas asignadas.
- Monto total calculado exclusivamente en el servidor.
- Fechas de ingreso y entrega estimada.
- La entrega estimada no puede ser anterior al ingreso y debe respetar la zona horaria aprobada.
- Estado actual e historial inmutable de transiciones.

Flujo permitido:

- `PENDIENTE → EN_PROCESO → CONTROL_CALIDAD → TERMINADO → ENTREGADO`.
- `CANCELADO` como salida autorizada desde los estados aprobados.
- No se permiten saltos, retrocesos ni reapertura en la versión 1.

Reglas críticas:

- Administrador y Encargado crean órdenes y asignan personal.
- Personal de lavado solo ve órdenes donde está asignado.
- La protección contra acceso a órdenes ajenas debe existir en listados, detalle y acciones; prueba manipulación de ID.
- Solo se usan vehículo, servicios, precios y personal válidos/activos según su contexto.
- No aceptar precio, total, encargado, cobrador o estado arbitrario enviado por el frontend.
- El backend obtiene los precios vigentes por tipo de vehículo y los congela en `OrdenServicio`.
- Cambiar la matriz después no modifica órdenes existentes.
- No repetir el mismo servicio ni la misma persona dentro de una orden.
- Crear orden, líneas, asignaciones, total y primer historial debe ser una transacción atómica.
- Cambiar estado e insertar historial debe ser una transacción atómica.
- Cada historial registra origen, destino, instante, actor y motivo/observación cuando corresponda.
- Cancelar exige motivo y no borra la orden.
- Aplica control de concurrencia para impedir transiciones incompatibles o actualizaciones perdidas.
- Una orden entregada o cancelada es terminal.
- No implementes todavía cobro; si la política final exige pago antes de Entregado, puedes bloquear temporalmente Entregado y documentar que se habilitará en el Módulo 6. No simules un pago.

Completa también:

- Historial del cliente con órdenes de todos sus vehículos.
- Historial del vehículo con órdenes de ese vehículo.
- Vista operativa de órdenes del día.
- Vista del Personal de lavado limitada a asignaciones propias.
- Filtros por estado, fecha, vehículo, cliente y personal según permisos.

Interfaz mínima:

- Listado operativo.
- Creación de orden con cliente/vehículo, servicios y precios visibles.
- Asignación múltiple de Personal de lavado.
- Detalle con total e historial cronológico.
- Acciones de estado disponibles según rol y estado actual.
- Cancelación con motivo y confirmación.
- Estados de carga, error, vacío, concurrencia y permiso denegado.

Pruebas mínimas obligatorias:

- Orden con varios servicios suma correctamente en servidor.
- Orden con varias personas no duplica asignaciones.
- Fecha estimada anterior al ingreso es rechazada.
- Combinación sin precio bloquea creación.
- Precio o total manipulado desde el cliente se ignora/rechaza.
- Cambio posterior de precio no altera orden previa.
- Personal solo accede a órdenes propias.
- Matriz completa de transiciones válidas e inválidas por rol.
- Cada transición válida crea exactamente un historial.
- Una transición fallida no cambia estado ni crea historial parcial.
- Cancelación exige motivo y solo roles permitidos.
- Entregada/cancelada no se reabre.
- Dos solicitudes simultáneas no corrompen el estado.
- Historial de cliente y vehículo devuelve las órdenes correctas sin N+1 evidente.
- Migraciones desde base vacía, tests, build, lint y análisis estático pasan.

Límites:

- No implementes pagos reales ni reportes financieros.
- No implementes WhatsApp ni notificaciones.
- No introduzcas estados o descuentos no definidos.
- No hagas commit, merge, rebase, push ni elimines ramas.

Actualiza `tasks.md`, `PROGRESS.md`, `lessons.md` y la documentación del Módulo 5 con la máquina de estados y permisos finales.

Informe final:

- Flujo implementado y decisiones finales.
- Modelo, transacciones y concurrencia.
- Endpoints y pantallas.
- Matriz de permisos y transiciones probada.
- Evidencia del congelamiento de precios.
- Pruebas/comandos y resultados.
- Función temporalmente bloqueada por pagos, si aplica.
- Riesgos y estado de Git.
- Plan del Módulo 6.

Al terminar, detente. No implementes el Módulo 6.

Criterios de aceptación del Módulo 5:

- La orden se crea atómicamente y el total proviene del servidor.
- Precios y, si se aprobó, nombres quedan congelados.
- Los estados y permisos no se pueden evadir.
- El Personal de lavado no ve órdenes ajenas.
- Toda transición tiene historial consistente.
- Los historiales de cliente y vehículo funcionan.

## Prompt 6. Pagos

Rama propuesta: `feat/modulo-06-pagos`

Implementa únicamente el Módulo 6: pago total asociado a una orden.

Rama esperada: `feat/modulo-06-pagos`. Si no existe y el worktree está limpio, créala desde el punto de integración aprobado del Módulo 5. Si no puedes determinarlo con evidencia, detente.

Contexto y precondiciones:

- Lee por completo este archivo único, `AGENTS.md`, documentación del proyecto, `tasks.md`, `PROGRESS.md` y `lessons.md`.
- Verifica que los Módulos 1 a 5 estén concluidos y que el flujo de órdenes sea estable.
- Revisa Git y preserva cambios preexistentes.
- El flujo de cobro debe usar los formularios, modales, alertas, formato monetario y estados de envío del proyecto base, manteniendo continuidad visual y accesibilidad.

Decisiones obligatorias antes de implementar:

- Momento del cobro: recomendación, registrar el pago antes o dentro de Terminado → Entregado.
- Métodos de pago aceptados y si serán catálogo o conjunto validado.
- Procedimiento de anulación/corrección de un cobro erróneo.
- Si se exige comprobante o referencia externa.

No inventes métodos de pago ni una política contable. Si falta una decisión, detente y pregunta.

Alcance exclusivo:

- Un máximo de un pago válido por orden.
- Pago por el monto total exacto de la orden.
- Registro de método, fecha/hora y usuario que cobró.
- Consulta del pago desde el detalle de la orden.
- Integración segura con la transición a `ENTREGADO` según la política aprobada.
- Corrección/anulación solo si fue definida y con trazabilidad; nunca borrado silencioso.

Reglas críticas:

- No existen anticipos, saldos ni pagos parciales en la versión 1.
- Restricción única de base de datos para `orden_id`.
- Administrador y Encargado registran pagos; Personal de lavado no accede a acciones de cobro.
- El backend toma el monto desde la orden. No acepta un importe distinto enviado por el cliente.
- No se paga una orden cancelada.
- No se duplica un pago por reintento, doble clic o solicitud concurrente.
- Pago y transición relacionada deben ejecutarse atómicamente o mediante una estrategia idempotente y consistente aprobada.
- Si el pago falla, la orden no debe quedar entregada.
- Si la transición falla, no debe quedar un pago huérfano o incoherente.
- No exponer datos financieros innecesarios al Personal de lavado.

Interfaz mínima:

- Acción de cobro solo cuando corresponda.
- Monto total de solo lectura obtenido del servidor.
- Selector de método validado.
- Confirmación clara antes de registrar.
- Prevención de doble envío.
- Estado de pago visible en detalle para roles autorizados.
- Errores de concurrencia e idempotencia comprensibles.

Pruebas mínimas obligatorias:

- Pago exacto exitoso por Administrador y Encargado.
- Personal de lavado recibe permiso denegado.
- Intentos de pago parcial, superior o manipulado son rechazados/ignorados de forma segura.
- Segunda operación sobre la misma orden no crea otro pago.
- Solicitudes simultáneas producen como máximo un pago.
- Orden cancelada no se paga.
- Orden no se entrega si el pago falla.
- Transacción o idempotencia mantiene consistencia ante fallos intermedios.
- Corrección/anulación cumple la política aprobada y deja auditoría.
- Migraciones, tests, build, lint y análisis estático pasan.

Límites:

- No implementes pagos parciales, anticipos, descuentos, devoluciones o facturación no definidos.
- No implementes dashboard o reportes salvo el ajuste mínimo para mostrar el pago en la orden.
- No implementes integraciones bancarias ni externas sin solicitud expresa.
- No hagas commit, merge, rebase, push ni elimines ramas.

Actualiza `tasks.md`, `PROGRESS.md`, `lessons.md` y la documentación del módulo.

Informe final:

- Política de cobro aplicada.
- Modelo, restricción única e idempotencia/transacción.
- Endpoints y pantallas.
- Integración con `ENTREGADO`.
- Pruebas y evidencia de concurrencia.
- Riesgos, pendientes y estado de Git.
- Plan del Módulo 7.

Al terminar, detente. No implementes el Módulo 7.

Criterios de aceptación del Módulo 6:

- Cada orden admite como máximo un pago por el total exacto.
- No existe pago parcial ni importe manipulable desde el frontend.
- La entrega y el cobro mantienen consistencia.
- Los reintentos no duplican pagos.
- Los permisos y la trazabilidad están probados.

## Prompt 7. Dashboard y reportes

Rama propuesta: `feat/modulo-07-reportes`

Implementa únicamente el Módulo 7: dashboard y reportes operativos/financieros básicos.

Rama esperada: `feat/modulo-07-reportes`. Si no existe y el worktree está limpio, créala desde el punto de integración aprobado del Módulo 6. Si no puedes determinarlo con evidencia, detente.

Contexto y precondiciones:

- Lee por completo este archivo único, `AGENTS.md`, documentación, `tasks.md`, `PROGRESS.md` y `lessons.md`.
- Verifica que los Módulos 1 a 6 estén concluidos y sus modelos sean la fuente real de datos.
- Revisa Git y preserva cambios del usuario.
- El dashboard debe usar las tarjetas, tablas, filtros, tipografías, colores y componentes de visualización existentes. No incorpores otra librería de gráficos salvo que el proyecto no tenga una y exista autorización.

Decisiones obligatorias:

- Confirma si el Encargado ve reportes financieros o solo operativos.
- Confirma zona horaria y límites de día/rango.
- Define cómo atribuir producción cuando una orden tiene varias personas. Recomendación inicial: contar participación por persona, pero contar orden e ingreso una sola vez en totales generales.
- Define qué pagos anulados o estados terminales participan en ingresos según la política aprobada.

No inventes métricas ni fórmulas si afectan resultados del negocio.

Indicadores mínimos:

- Vehículos actualmente en el lavadero agrupados por estado.
- Órdenes del día.
- Producción por integrante del Personal de lavado.
- Ingresos por rango de fechas.
- Ingresos por servicio.
- Ingresos por tipo de vehículo.
- Tiempo promedio por etapa usando `HistorialEstado`.
- Servicios más solicitados.

Reglas críticas:

- Los ingresos provienen de pagos válidos, no de órdenes creadas o terminadas sin cobrar.
- Una orden con varios servicios o personas no puede duplicar el ingreso total.
- Diferencia claramente conteo de órdenes, vehículos, servicios y participaciones del personal.
- Los filtros de fecha usan la zona horaria aprobada y límites inclusivos/exclusivos documentados.
- Administrador ve los reportes autorizados.
- Encargado ve solo el alcance aprobado.
- Personal de lavado no ve reportes financieros generales; solo su información operativa si fue aprobada.
- No exponer CI, teléfono u otros datos personales innecesarios.
- Evita consultas N+1 y agregaciones sin índices.
- Aplica paginación o límites a detalles extensos.
- No uses datos simulados en producción ni mezcles mocks con consultas reales.

Implementación esperada:

1. Consultas/agregaciones backend con contratos claros.
2. Índices justificados por consultas reales.
3. Endpoints protegidos y filtros validados.
4. Dashboard con estados de carga, vacío y error.
5. Tablas o visualizaciones solo donde ayuden a entender los datos.
6. Formato monetario y fechas coherentes.
7. Tests con conjuntos de datos diseñados para detectar duplicación.
8. Documentación de cada fórmula.

Casos de prueba obligatorios:

- Orden con dos servicios no duplica ingreso total.
- Orden con dos personas no duplica ingreso total ni conteo general.
- La producción por persona refleja la fórmula aprobada.
- Orden sin pago no aparece como ingreso.
- Pago anulado/corregido se trata según política.
- Rangos de fecha respetan zona horaria y límites.
- Tiempo por etapa se calcula desde transiciones válidas.
- Roles no autorizados reciben permiso denegado.
- Dataset vacío no rompe el dashboard.
- Consultas principales tienen rendimiento razonable y sin N+1 evidente.
- Tests, build, lint, formato y análisis estático pasan.

Límites:

- No agregues exportación Excel/PDF, metas, comisiones o inteligencia de negocio no solicitada.
- No implementes WhatsApp ni notificaciones.
- No cambies reglas históricas para facilitar reportes.
- No hagas commit, merge, rebase, push ni elimines ramas.

Actualiza `tasks.md`, `PROGRESS.md`, `lessons.md` y documenta definiciones/fórmulas de cada indicador.

Informe final:

- Indicadores implementados.
- Fórmulas y permisos.
- Endpoints, consultas, índices y pantallas.
- Evidencia de no duplicación.
- Pruebas y comandos.
- Rendimiento, riesgos y estado de Git.
- Plan para la revisión integral de la versión 1.

Al terminar, detente. No implementes notificaciones ni funciones futuras.

Criterios de aceptación del Módulo 7:

- Los indicadores usan datos reales y fórmulas documentadas.
- Los ingresos no se duplican.
- Fechas y zona horaria son coherentes.
- Cada rol ve únicamente la información aprobada.
- Consultas, tests y build pasan.

## Prompt 8. Integración, revisión de seguridad y cierre de la versión 1

Rama recomendada: realizar la revisión sobre la rama de integración definida por el usuario. No crear ni fusionar ramas sin autorización expresa.

Realiza la revisión integral y el cierre técnico de la versión 1 del Sistema de Gestión de Lavadero. No agregues funcionalidades nuevas.

Contexto obligatorio:

- Lee por completo este archivo único.
- Lee todos los `AGENTS.md`, documentación por módulo, `tasks.md`, `PROGRESS.md` y `lessons.md`.
- Verifica que los Módulos 1 a 7 estén marcados como concluidos con evidencia.
- Inspecciona el estado de Git y preserva cambios preexistentes.

Objetivo:

- Comprobar integración real de extremo a extremo.
- Detectar regresiones, contradicciones entre módulos, vulnerabilidades y documentación desactualizada.
- Corregir únicamente defectos necesarios para cumplir requisitos ya aprobados.
- No ampliar alcance.

Revisión funcional completa:

1. Autenticación con Administrador, Encargado y Personal de lavado.
2. Restricciones de permisos backend y frontend.
3. Alta de cliente y varios vehículos.
4. Catálogos y matriz de precios administrables.
5. Orden con varios servicios y precios congelados.
6. Asignación de varias personas.
7. Flujo Pendiente → En proceso → Control de calidad → Terminado → Entregado.
8. Cancelación autorizada con motivo.
9. Historial completo, consistente e inmutable.
10. Pago único por total exacto e integración con entrega.
11. Historiales de cliente y vehículo.
12. Dashboard y reportes sin duplicación.

Revisión técnica y de seguridad:

- Migraciones desde una base vacía en orden correcto.
- Restricciones, claves foráneas, índices y bajas lógicas.
- Transacciones de orden, estado y pago.
- Concurrencia e idempotencia en cambios críticos.
- Autorización por rol y por asignación/propiedad.
- Pruebas contra IDOR y manipulación de precio, total, rol, estado y cobrador.
- Contraseñas, sesiones/tokens, cookies, CSRF, CORS y rate limiting según arquitectura.
- Validación, normalización, errores y logs sin datos sensibles.
- Seeds bloqueados en producción.
- Dependencias y configuración sin secretos.
- Rendimiento básico y ausencia de N+1 en historiales/reportes críticos.

Verificaciones:

- Ejecuta la suite completa de tests.
- Ejecuta build de backend y frontend.
- Ejecuta formato, lint, análisis estático y chequeo de tipos disponibles.
- Ejecuta migraciones en una base limpia de prueba.
- Ejecuta pruebas de integración o end-to-end del flujo principal.
- Ejecuta `git diff --check` o equivalente.
- Revisa manualmente el diff final para cambios fuera de alcance, secretos y código muerto.

Reglas de corrección:

- Si una prueba falla, diagnostica y corrige la causa dentro del alcance aprobado.
- Si corregir exige cambiar una regla de negocio, detente y solicita decisión.
- No ocultes pruebas fallidas, no rebajes validaciones y no marques como aprobado algo no ejecutado.
- No agregues notificaciones por WhatsApp, pagos parciales, descuentos, facturación, exportaciones u otras funciones futuras.
- No hagas commit, merge, rebase, push, despliegue ni elimines ramas sin autorización expresa.

Documentación final:

- Actualiza `tasks.md`, `PROGRESS.md` y `lessons.md`.
- Alinea README, variables de entorno de ejemplo, comandos de desarrollo, migración, seed y pruebas con el comportamiento real.
- Cierra cada módulo solo si tiene evidencia.
- Registra pendientes reales como pendientes; no los borres para aparentar cierre.

Informe final obligatorio:

- Veredicto: APTO PARA REVISIÓN / NO APTO.
- Requisitos cumplidos y no cumplidos.
- Defectos encontrados y correcciones realizadas.
- Matriz final de permisos.
- Resultado del flujo end-to-end.
- Tabla completa de comandos y pruebas con resultado.
- Riesgos residuales y deuda técnica.
- Archivos cambiados.
- Estado de Git.
- Acciones que requieren autorización: commit, merge, push o despliegue.

Al terminar, detente. No publiques ni despliegues el sistema.

Criterios de aceptación del cierre:

- El flujo completo funciona con los tres roles.
- No se evaden permisos ni se manipulan importes o estados.
- Migraciones, tests, build y análisis pasan.
- No existen ingresos duplicados ni pagos parciales.
- La documentación coincide con el código.
- Cualquier pendiente está declarado con evidencia.

## Prompt opcional para autorizar el commit de un módulo ya aprobado

Este bloque solo se usa después de revisar el informe y aceptar el módulo.

El módulo actual fue revisado y aprobado. Antes de crear el commit:

1. Verifica que la rama sea la correcta.
2. Muestra `git status` y el resumen del diff.
3. Confirma que no haya secretos, cambios ajenos ni archivos fuera del módulo.
4. Repite las verificaciones finales relevantes si el árbol cambió después del informe.
5. Crea un único commit intencional con un mensaje que describa el alcance real.
6. No hagas push, merge, rebase ni elimines ramas.

Después informa el hash, mensaje, archivos incluidos, pruebas finales y estado del worktree.

