# Bug: Marcar una tarea como completada no se guarda

Issue: https://github.com/aclemotte-sanzclima/desarrollo-agentico-todo-03/issues/8

## Encargo y resultado esperado

Al pulsar el círculo de completar (o «Reabrir») de una tarea, la interfaz cambia
de forma optimista, pero el cambio no se persiste: al recargar `/`, la tarea
vuelve a su estado anterior.

Resultado esperado tras el arreglo:

- Pulsar el círculo envía al servidor la misma petición que el formulario
  (`POST /tasks/{id}/toggle` con `_method=PUT` y `authenticity_token`) y el
  estado `completed` queda guardado en la base de datos.
- Si la petición falla (respuesta no `ok` o error de red), la interfaz deshace
  el cambio optimista para no mostrar un estado que no se ha guardado.
- Sin cambios en el servidor, las rutas ni la plantilla.

## Evidencia y causa raíz

Hechos observados en el repositorio:

- `actions/app.go:55` solo registra `app.PUT("/tasks/{task_id}/toggle", TasksToggle)`.
  No hay ruta `GET` para ese path.
- `templates/tasks/index.plush.html:23-25`: el formulario del toggle es
  `method="POST"` con los campos ocultos `authenticity_token` y `_method=PUT`,
  y lleva el atributo `data-task-toggle`.
- `public/assets/app.js:12` llama a `event.preventDefault()` para todos los
  formularios `[data-task-toggle]`, de modo que el envío nativo (correcto) no
  ocurre nunca.
- `public/assets/app.js:34` hace `fetch(form.action)` sin opciones: es una
  petición `GET` sin cuerpo, que no coincide con ninguna ruta (404 según la
  issue) y no lleva token CSRF.
- `fetch` no rechaza la promesa ante un 404, y el código no inspecciona la
  respuesta, así que el fallo es silencioso.
- `actions/tasks.go:152-167` (`TasksToggle`) funciona: `Test_TasksToggle_CompletesAndReopens`
  (`actions/tasks_test.go:64`) lo cubre vía `PUT`.

Causa raíz: el cliente envía la petición de toggle con el método y el cuerpo
equivocados (`GET` sin campos) en lugar de reproducir el envío del formulario.

Incertidumbres (inferencias, no verificadas en este entorno porque la caché de
módulos de Go no está descargada):

- Estoy haciendo una inferencia: el `MethodOverride` por defecto de Buffalo
  convierte un `POST` con campo de formulario `_method=PUT` en `PUT`, leyendo el
  campo con `FormValue`, que acepta tanto `application/x-www-form-urlencoded`
  como `multipart/form-data`. Es lo que ya usa el envío nativo del formulario,
  así que enviar el cuerpo como `application/x-www-form-urlencoded` (igual que
  el navegador) elimina esta duda.
- Estoy haciendo una inferencia: el middleware `csrf` de `gobuffalo/middleware`
  no se aplica con `GO_ENV=test`, por eso los tests con `as.HTML(...).Put`
  pasan sin token. El test de regresión no puede, por tanto, verificar el
  token CSRF; este punto queda para la verificación manual.
- `fetch` sigue por defecto la redirección `303 → GET /`, por lo que una
  respuesta correcta termina en `200` con `response.ok === true`; un 403
  (CSRF) o 404 da `ok === false`.

Reproducción mínima (manual, en navegador, con la app levantada):

1. Abrir `/` con una tarea pendiente.
2. Pulsar su círculo de completar → se ve tachada.
3. Recargar → vuelve a estar pendiente. En las herramientas de red se ve
   `GET /tasks/<id>/toggle` → 404.

## Alcance

Incluido:

- Cambiar la petición de `public/assets/app.js` para que sea
  `POST form.action` con el cuerpo del formulario codificado como
  `application/x-www-form-urlencoded` (incluye `_method=PUT` y
  `authenticity_token`).
- Comprobar la respuesta y, si no es `ok` o la red falla, revertir el cambio
  optimista de la interfaz.
- Añadir un test Go que fije el contrato que el cliente usa:
  `POST /tasks/{id}/toggle` con `_method=PUT` alterna `completed`.

Excluido:

- Cambios en `actions/`, rutas, plantillas, modelo o migraciones.
- Introducir un framework de tests de JavaScript o end-to-end (nueva
  dependencia y nueva configuración de CI; requiere decisión humana).
- Mensajes de error visibles o flashes para el fallo (basta con revertir).
- Cualquier refactor del manejador de `data-confirm`.

## Archivos relevantes

- `public/assets/app.js` — manejador cliente del toggle; aquí está el fallo.
- `templates/tasks/index.plush.html` — formulario del toggle (`POST`,
  `_method=PUT`, `authenticity_token`, `data-task-toggle`). Solo lectura.
- `actions/app.go` — rutas; `PUT /tasks/{task_id}/toggle`. Solo lectura.
- `actions/tasks.go` — `TasksToggle`. Solo lectura.
- `actions/tasks_test.go` — tests de acciones; aquí va el nuevo test.

## Implementación paso a paso

1. En `public/assets/app.js`, extraer la actualización visual que hoy está en
   las líneas 15-32 a una función local (por ejemplo `applyToggleState(task, form, completed)`)
   que, dado el estado `completed`, ajuste la clase `task-completed`, el texto
   de `.toggle-label`, `aria-pressed`, `aria-label`, `title` y el
   `span.task-status`. Debe usar `classList.toggle("task-completed", completed)`
   con el segundo argumento para poder fijar el estado en ambos sentidos.
   Mantener los textos actuales («Reabrir», «Completar», «Completada»).
2. En el manejador, calcular `var completed = !task.classList.contains("task-completed");`,
   llamar a `applyToggleState(task, form, completed)` (actualización optimista).
3. Sustituir `fetch(form.action);` por:

   ```js
   fetch(form.action, {
     method: "POST",
     body: new URLSearchParams(new FormData(form)),
     credentials: "same-origin",
     headers: { Accept: "text/html" }
   })
     .then(function (response) {
       if (!response.ok) {
         throw new Error("HTTP " + response.status);
       }
     })
     .catch(function () {
       applyToggleState(task, form, !completed);
     });
   ```

   `new URLSearchParams(new FormData(form))` envía los campos ocultos
   `authenticity_token` y `_method=PUT` como `application/x-www-form-urlencoded`,
   igual que el envío nativo.
4. No cambiar nada más del archivo (el bloque de `data-confirm` se queda igual).
5. En `actions/tasks_test.go`, añadir `Test_TasksToggle_PostWithMethodOverride`:
   crear una tarea pendiente, hacer
   `as.HTML("/tasks/%s/toggle", task.ID).Post(url.Values{"_method": {"PUT"}})`,
   comprobar `303`, recargar y comprobar `task.Completed == true`; repetir y
   comprobar `false`. Si el helper `Post` de `httptest` de Buffalo no aplica el
   method override, documentarlo en el resultado y no forzar el test.

## Tests y validación

- `tr -d '\r' < actions/tasks_test.go | gofmt -l` (y `gofmt -l .` sobre LF):
  sin salida.
- `go vet ./...`: sin errores.
- `soda migrate -e test` y `go test -p 1 ./...`: todos pasan, incluido el nuevo
  `Test_TasksToggle_PostWithMethodOverride`.
- `go build ./...`: compila (el JS va embebido en `public`).
- `node --check public/assets/app.js`: sintaxis válida (Node está disponible en
  el host; no es una dependencia del repo).

Limitación conocida: el repositorio no tiene tests de JavaScript, así que
**ningún test automático falla antes del arreglo**. El nuevo test Go protege el
contrato del servidor del que depende el cliente, pero pasa también sobre
`main`. La regresión real del cliente se valida manualmente:

1. Levantar la app (`docker compose up`), abrir `/` con una tarea pendiente.
2. Pulsar el círculo; en la pestaña de red debe verse
   `POST /tasks/<id>/toggle` → `303` → `GET /` → `200`.
3. Recargar: la tarea sigue completada. Pulsar «Reabrir» y recargar: sigue
   pendiente.
4. Con la app parada, pulsar el círculo: la interfaz vuelve al estado previo.

Si la verificación manual no puede hacerse en la ejecución, debe indicarse en
el PR como pendiente.

## Riesgos y recuperación

- CSRF en desarrollo/producción: si el token no viaja, el servidor responde
  403 y la interfaz revierte (ya no hay fallo silencioso de persistencia). La
  verificación manual del paso 2 lo confirma.
- Doble clic rápido: dos peticiones en vuelo alternan dos veces; el estado
  final del servidor coincide con el visual salvo que una falle. Riesgo menor,
  no cubierto.
- El `fetch` sigue la redirección y descarga el HTML de `/`; coste pequeño y
  aceptable.
- Recuperación: revertir el commit restaura el comportamiento anterior; no hay
  migraciones ni cambios de datos.
