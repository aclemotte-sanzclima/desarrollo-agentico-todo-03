# Feature: Prioridades de tareas (baja, media, alta) y ordenación por prioridad

## Encargo y criterios de aceptación

Encargo literal: «añade prioridades a las tareas de 3 tipos, baja, media y alta y permite ordenar las tareas según estas prioridades».

### Requisitos explícitos

- R1. Cada tarea tiene una prioridad, con exactamente tres valores: **Baja**, **Media** y **Alta**.
- R2. El usuario puede ordenar la lista de tareas según la prioridad.

### Requisitos inferidos (decisiones del plan, revisables)

- I1. La prioridad se elige al **crear** la tarea (selector en el formulario de alta), con **Media** como valor por defecto.
- I2. La prioridad se puede **cambiar** desde la pantalla de edición existente (`/tasks/{id}/edit`).
- I3. Las tareas ya existentes (seed y datos de desarrollo) pasan a tener prioridad **Media**.
- I4. La ordenación es una **opción** elegida por el usuario, no un cambio del orden por defecto: por defecto la lista sigue ordenada por fecha de creación; con `?sort=priority` se ordena **Alta → Media → Baja**, y a igual prioridad por `created_at asc, id asc` (orden actual).
- I5. La prioridad de cada tarea es visible en la lista (etiqueta textual, no solo color).
- I6. La ordenación por prioridad no separa completadas de pendientes (no se pidió).

### Criterios de aceptación

1. `GET /` muestra, para cada tarea, su prioridad como texto («Alta», «Media» o «Baja»).
2. El formulario de alta incluye un `<select name="priority">` etiquetado con las tres opciones y «Media» preseleccionada.
3. `POST /tasks` con `priority=3` crea una tarea con prioridad Alta; sin `priority` (o `0`) crea la tarea con prioridad Media.
4. `POST /tasks` con una prioridad fuera de rango (p. ej. `7`) o no numérica responde 422, muestra «La prioridad no es válida.» y no crea la tarea.
5. `GET /tasks/{id}/edit` muestra el selector con la prioridad actual seleccionada.
6. `PUT /tasks/{id}` con `priority` válida la actualiza junto con el título; sin `priority` la conserva; con un valor inválido responde 422 y no modifica la tarea. `completed`, `id` y `created_at` siguen sin poder cambiarse por esta vía.
7. `GET /?sort=priority` lista las tareas Alta → Media → Baja; empates por antigüedad.
8. `GET /` y `GET /?sort=<cualquier otro valor>` mantienen el orden actual (`created_at asc, id asc`).
9. La lista ofrece dos enlaces de ordenación («Creación» y «Prioridad»); el activo lleva `aria-current="true"`.
10. La base de datos rechaza prioridades fuera de 1–3 (restricción CHECK) y las filas existentes quedan con prioridad 2.

## Comportamiento actual y propuesta

### Actual (verificado en el repositorio)

- Tabla `tasks` (`migrations/20260901000000_create_tasks.up.fizz`): `id uuid`, `title string(200)`, `completed bool default false`, timestamps. No hay prioridad.
- `models/task.go`: struct `Task` con `Title` (`form:"title"`), resto `form:"-"`. `Validate` solo valida el título.
- `actions/tasks.go`:
  - `TasksIndex` y la rama 422 de `TasksCreate` duplican `tx.Order("created_at asc, id asc").All(&tasks)`.
  - `TasksCreate` hace `c.Bind(task)` sobre el modelo.
  - `TasksUpdate` hace bind sobre un struct anónimo con **solo** `Title` para que el resto de campos no sean asignables (lo protege `Test_TasksUpdate_PersistsTitleOnly`).
- Rutas en `actions/app.go:51-57`; no hay parámetros de consulta en uso.
- `templates/tasks/index.plush.html`: cada `<li class="task">` es un grid de 3 columnas (toggle, título, acciones); `.task-status` es `sr-only` (posición absoluta, no ocupa celda).
- `public/assets/app.js` usa `.task-title.textContent` para construir `aria-label`; por eso la prioridad **no** debe ir dentro de `.task-title`.
- Las migraciones se aplican al arrancar (`cmd/app/main.go` → `models.Migrate()`) y en CI con `soda migrate -e test`.

### Propuesta

Cambio vertical mínimo siguiendo los patrones existentes: columna entera `priority` (1 = baja, 2 = media, 3 = alta) con default 2 y CHECK; constantes y validación en el modelo; selector en los formularios de alta y edición; etiqueta en cada fila; parámetro `sort` en `GET /` resuelto mediante una lista blanca que mapea a cláusulas `ORDER BY` fijas.

Por qué entero y no texto: `ORDER BY priority DESC` es directo y no necesita `CASE`; el texto visible se deriva en el modelo.

## Alcance

### Incluido

- Migración para añadir `priority` a `tasks`.
- Modelo: constantes, opciones para el selector, etiqueta y validación.
- Acciones `TasksIndex`, `TasksCreate`, `TasksEdit`, `TasksUpdate`.
- Templates `index` y `edit`, y CSS para el selector, la etiqueta de prioridad y los enlaces de ordenación.
- Tests de modelo y de acciones.

### Excluido

- Filtrar por prioridad, ordenar por otros criterios u ordenación ascendente/descendente configurable.
- Conservar `sort` tras crear, editar, completar o eliminar (los redirects siguen yendo a `/`). Se puede hacer en otra iteración.
- Cambiar la prioridad desde la lista sin entrar a edición, o con JavaScript.
- Persistir la preferencia de orden (cookie/sesión).
- Cambios en `app.js`.
- **Observación fuera de alcance:** `public/assets/app.js` hace `fetch(form.action)` sin método, es decir, un `GET /tasks/{id}/toggle`, que no tiene ruta. Estoy haciendo una inferencia, sin comprobarlo en el navegador: el toggle hecho con JS probablemente no se guarda. Conviene tratarlo con `/bug` por separado.

## Diseño por capas

### Persistencia y migraciones

Nuevo par `migrations/20261014000000_add_priority_to_tasks.{up,down}.fizz`:

```fizz
// up
add_column("tasks", "priority", "integer", {"default": 2})
sql("ALTER TABLE tasks ADD CONSTRAINT tasks_priority_check CHECK (priority BETWEEN 1 AND 3)")
```

```fizz
// down
drop_column("tasks", "priority")
```

- En fizz las columnas son `NOT NULL` por defecto. Con `default 2`, PostgreSQL rellena las filas existentes, incluidas las del seed.
- La restricción se elimina con la columna en el `down`.
- `migrations/embed.go` ya incluye `*.fizz`; no hay que tocarlo.
- El seed `20260901000001_seed_tasks.up.sql` no cambia: se ejecuta antes de la nueva migración y sus filas reciben el default.

### Modelo (`models/task.go`)

```go
const (
	PriorityLow    = 1
	PriorityMedium = 2
	PriorityHigh   = 3
)

// PriorityOption es una opción del selector de prioridad.
type PriorityOption struct {
	Value int
	Label string
}

// PriorityOptions enumera las prioridades de mayor a menor.
var PriorityOptions = []PriorityOption{
	{PriorityHigh, "Alta"}, {PriorityMedium, "Media"}, {PriorityLow, "Baja"},
}
```

- Campo nuevo en `Task`: `Priority int \`json:"priority" db:"priority" form:"-"\``. Lleva `form:"-"` porque la acción lo lee y convierte a mano (ver «Acción y rutas»).
- Método de valor `func (t Task) PriorityLabel() string`: devuelve la etiqueta o `""` si el valor no es válido.
- Método de valor `func (t Task) PriorityClass() string`: devuelve `"priority-high"`, `"priority-medium"` o `"priority-low"`.
- `Validate`: si `t.Priority` no está en {1,2,3} → `verrs.Add("priority", "La prioridad no es válida.")`.
- El default (0 → Media) **no** se aplica en `Validate`. Se aplica en la acción, para que el modelo no muestre efectos ocultos y el test de modelo pueda comprobar que 0 es inválido.
  - Consecuencia: el test de modelo existente `Test_Task_Create_Valid` crea `&Task{Title: …}` con `Priority` 0 y fallaría. Hay que ajustarlo para que use `Priority: PriorityMedium`. Ver la alternativa en «Riesgos».

### Acción y rutas (`actions/tasks.go`)

No hay rutas nuevas. `GET /` acepta el parámetro de consulta `sort`.

- Helper privado nuevo que sustituye a las dos llamadas duplicadas:

  ```go
  // taskOrders mapea el parámetro sort a cláusulas ORDER BY fijas (lista blanca).
  var taskOrders = map[string]string{
      "priority": "priority desc, created_at asc, id asc",
  }
  const defaultTaskOrder = "created_at asc, id asc"

  func listTasks(tx *pop.Connection, sort string) (models.Tasks, string, error)
  ```

  Devuelve las tareas y el valor de `sort` normalizado (`"priority"` o `""`). El valor del usuario **nunca** se concatena en SQL.
- `TasksIndex`:
  - `listTasks(tx, c.Param("sort"))`.
  - `c.Set("sort", sort)`, `c.Set("priorities", models.PriorityOptions)`.
  - `task` inicial: `&models.Task{Priority: models.PriorityMedium}`, para que el selector salga con «Media».
- `TasksCreate`:
  - Tras `c.Bind(task)`: `if task.Priority == 0 { task.Priority = models.PriorityMedium }`.
  - Riesgo de bind: un valor no numérico (`priority=abc`) puede hacer fallar `c.Bind` y devolver 500 en vez de 422 (incierto, depende del decoder de formularios de Buffalo). Para controlarlo, el campo `Priority` del modelo se declara con `form:"-"` y la acción lee `c.Param("priority")` y lo convierte con `strconv.Atoi`:
    - Vacío → Media.
    - Error de conversión → valor `-1`, que `Validate` rechaza con 422.
  - Se usa este enfoque en create y update, con un helper común `parsePriority(raw string, fallback int) int`.
  - En la rama 422: `listTasks(tx, "")`, más `c.Set("sort", "")` y `c.Set("priorities", …)`.
- `TasksEdit`: `c.Set("priorities", models.PriorityOptions)`.
- `TasksUpdate`:
  - Mantiene el bind solo de `Title`.
  - Después: `task.Priority = parsePriority(c.Param("priority"), task.Priority)`, es decir, si no viene se conserva la actual.
  - En la rama 422 también hace `c.Set("priorities", …)`.
- Actualizar el comentario de `TasksUpdate` («actualiza el título y la prioridad, conservando los demás datos…»).

Con `form:"-"`, el JSON y la base de datos no cambian; solo se excluye del bind automático, igual que ya ocurre con `completed`.

### Renderizado (Plush)

`templates/tasks/index.plush.html`:

- Bajo `.list-count`, añadir una navegación de orden:

  ```html
  <nav class="sort-options" aria-label="Ordenar tareas">
    <span>Ordenar por:</span>
    <a href="/" <%= if (sort != "priority") { %>aria-current="true"<% } %>>Creación</a>
    <a href="/?sort=priority" <%= if (sort == "priority") { %>aria-current="true"<% } %>>Prioridad</a>
  </nav>
  ```

  Mostrarla solo si `len(tasks) > 1` es opcional; el plan la muestra siempre que haya tareas.
- En cada `<li>`, entre `.task-title` y `.task-status`:

  ```html
  <span class="task-priority <%= t.PriorityClass() %>"><span class="sr-only">Prioridad: </span><%= t.PriorityLabel() %></span>
  ```
- En el formulario de alta, dentro de `.new-task-row` y antes del botón:

  ```html
  <label for="task-priority" class="sr-only">Prioridad</label>
  <select id="task-priority" name="priority">
    <%= for (p) in priorities { %><option value="<%= p.Value %>" <%= if (p.Value == task.Priority) { %>selected<% } %>><%= p.Label %></option><% } %>
  </select>
  ```
- Mensajes de error: hoy solo se pinta `errors.Get("title")` dentro de `#title-error`, y `aria-invalid` se aplica al título siempre que `errors.HasAny()`. Hay que:
  - Pintar también `errors.Get("priority")` en un `<p id="priority-error" class="field-error" role="alert">`.
  - Condicionar `aria-invalid`/`aria-describedby` de cada campo a sus propios errores: `len(errors.Get("title")) > 0` y `len(errors.Get("priority")) > 0`.
  - Mantener el texto y los atributos actuales cuando el error es de título, porque los tests existentes buscan `aria-describedby="title-error"`.

`templates/tasks/edit.plush.html`:

- Tras el bloque del título, añadir `<label for="task-priority">Prioridad</label>` y el mismo `<select>`, con la prioridad actual seleccionada y su error propio.
- Cambiar el hint: «Hasta 200 caracteres. El estado de la tarea no cambia.», porque ahora la prioridad sí cambia.
- Igual que en `index`, separar la condición de error del título de la de prioridad.

### CSS (`public/assets/app.css`)

- `.task`: `grid-template-columns: 44px minmax(0, 1fr) auto auto;` y, en el media query de 760px, `36px minmax(0, 1fr) auto auto`.
- `.task-priority`:
  - `align-self: center`, `font-size: 0.75rem`, `font-weight: 600`, padding pequeño, `border-radius: 999px`.
  - Variantes `.priority-high`, `.priority-medium` y `.priority-low` con colores de contraste AA. La alta usa tonos de `--color-primary` y fondo `#fcf0ec`; la media y la baja, tonos neutros.
  - El texto siempre está presente, así que el color no es el único indicador.
- `.task-completed .task-priority`: atenuado (`color: var(--color-muted)`).
- `select`: mismo estilo base que `input[type="text"]` (`min-height: 44px`, borde, radio, `font: inherit`), con `width: auto` en `.new-task-row`. Añadir `select:focus-visible` a la regla de foco existente y `select[aria-invalid="true"]` a la de error.
- `.sort-options`: fila flex, `font-size: 0.8rem`, color `--color-muted`; enlace con `aria-current="true"` en `--color-text` y `font-weight: 600`. Área táctil mínima de 44px de alto en los enlaces.

### JavaScript

Sin cambios. La etiqueta de prioridad queda fuera de `.task-title`, así que el `aria-label` que reconstruye `app.js` sigue siendo correcto.

## Archivos relevantes

| Ruta | Responsabilidad |
|---|---|
| `migrations/20261014000000_add_priority_to_tasks.up.fizz` (nuevo) | Añade la columna `priority` con default 2 y la restricción CHECK 1–3 |
| `migrations/20261014000000_add_priority_to_tasks.down.fizz` (nuevo) | Elimina la columna |
| `models/task.go` | Campo `Priority`, constantes, `PriorityOptions`, `PriorityLabel`, `PriorityClass`, validación |
| `models/task_test.go` | Tests de validación y persistencia de la prioridad; ajuste de `Test_Task_Create_Valid` |
| `actions/tasks.go` | `listTasks` con lista blanca de orden, `parsePriority`, uso en Index/Create/Edit/Update |
| `actions/tasks_test.go` | Tests de alta, edición, validación, render y ordenación |
| `templates/tasks/index.plush.html` | Navegación de orden, etiqueta por tarea, selector en el alta, errores por campo |
| `templates/tasks/edit.plush.html` | Selector de prioridad, hint actualizado, errores por campo |
| `public/assets/app.css` | Estilos de `select`, `.task-priority`, `.sort-options` y grid de `.task` |
| `fixtures/tasks.toml` | Sin cambios obligatorios: las filas sin `priority` reciben el default de la BD |
| `actions/app.go` | Sin cambios (no hay rutas nuevas) |

## Implementación paso a paso

1. **Migración.** Crear los dos `.fizz` descritos. Aplicar en local: `soda migrate -e development` (o arrancar la app) y `soda migrate -e test`. Verificar `\d tasks` en psql: `priority integer NOT NULL DEFAULT 2` y `tasks_priority_check`.
2. **Modelo.**
   - Añadir constantes, `PriorityOption`, `PriorityOptions`, el campo `Priority` (con `form:"-"`), `PriorityLabel`, `PriorityClass` y la validación.
   - Ajustar `Test_Task_Create_Valid` para que use `Priority: PriorityMedium`.
   - Añadir los tests de modelo (ver más abajo).
3. **Acciones.**
   - Introducir `taskOrders`, `defaultTaskOrder`, `listTasks` y `parsePriority`, y reemplazar las dos consultas duplicadas.
   - Aplicar la prioridad en `TasksCreate` y `TasksUpdate`.
   - Hacer `c.Set("priorities", …)` y `c.Set("sort", …)` en todas las ramas que renderizan `index` o `edit`.
   - Actualizar el comentario de `TasksUpdate`.
4. **Templates.** `index.plush.html` (navegación de orden, etiqueta y selector, errores por campo) y `edit.plush.html` (selector, hint, errores por campo).
5. **CSS.** Reglas descritas. Revisar a mano en 380px, 760px y escritorio que la fila no desborda con títulos largos y la etiqueta.
6. **Tests de acción** (ver más abajo).
7. **Validación completa** (sección siguiente).

## Tests y validación

### Tests de modelo (`models/task_test.go`)

- `Test_Task_Priority_ValidValues`: 1, 2 y 3 crean la tarea sin errores, y `Find` devuelve la misma prioridad.
- `Test_Task_Priority_Invalid`: 0, 4 y -1 dan error en `priority` con «La prioridad no es válida.» y no se crea nada.
- `Test_Task_PriorityLabel`: 1 → «Baja», 2 → «Media», 3 → «Alta», 9 → «».
- `Test_Task_Priority_DBDefault`: insertar con `RawQuery` sin la columna `priority` y comprobar que vale 2. Así se verifica la migración.

### Tests de acción (`actions/tasks_test.go`)

- `Test_TasksCreate_WithPriority`: `Post(url.Values{"title": {"Urgente"}, "priority": {"3"}})` → 303, y la tarea queda con `Priority == 3`.
- `Test_TasksCreate_DefaultPriority`: el `Test_TasksCreate_Valid` existente (sin `priority`) debe seguir pasando; añadirle `as.Equal(models.PriorityMedium, task.Priority)`.
- `Test_TasksCreate_InvalidPriority`: `priority` = `"7"` y `"abc"` → 422, el cuerpo contiene «La prioridad no es válida.» y el conteo es 0.
- `Test_TasksIndex_ShowsPriority`: crear tareas con prioridades distintas y comprobar que aparecen «Alta», «Baja» y `class="task-priority priority-high`.
- `Test_TasksIndex_NewTaskDefaultsToMedium`: el cuerpo contiene `value="2" selected` (o el fragmento equivalente según el render final).
- `Test_TasksIndex_SortByPriority`:
  - Crear, en este orden, una Baja «B», una Alta «A» y una Media «M».
  - `GET /?sort=priority` → `strings.Index(body, "A") < Index("M") < Index("B")`. Usar títulos únicos y no ambiguos, p. ej. «Tarea alta», porque «Alta» también aparece como etiqueta.
  - El enlace «Prioridad» tiene `aria-current`.
- `Test_TasksIndex_SortByPriority_TieBreakByCreation`:
  - Dos tareas Alta. Forzar `created_at` con `as.DB.RawQuery("UPDATE tasks SET created_at = ? WHERE id = ?", …)` para no depender de la resolución del reloj.
  - Comprobar que la más antigua va primero.
- `Test_TasksIndex_DefaultOrderUnchanged`: `GET /` y `GET /?sort=foo'; drop table tasks;--` mantienen el orden por creación y responden 200. Cubre la lista blanca.
- `Test_TasksEdit_ShowsCurrentPriority`: una tarea Alta → la opción `value="3"` tiene `selected`.
- `Test_TasksUpdate_ChangesPriority`: `PUT` con título y `priority=1` → 303 y prioridad 1.
- `Test_TasksUpdate_KeepsPriorityWhenAbsent`: ampliar `Test_TasksUpdate_PersistsTitleOnly`, que no envía `priority`, con `as.Equal(priorOriginal, task.Priority)`.
- `Test_TasksUpdate_InvalidPriority`: `priority=9` → 422 con el mensaje, y la tarea no cambia (ni título ni prioridad).
- Ajustar `taskForEditing` si hace falta, para que cree con `Priority: models.PriorityMedium`, porque la validación ahora exige prioridad.
  - Revisar también los demás tests que llaman a `ValidateAndCreate(&models.Task{Title: …})` directamente (`Test_TasksToggle_CompletesAndReopens`, `Test_TasksDestroy`, `Test_TasksEdit_EscapesTitle`): con prioridad 0 ahora darían `verrs`. Hay que añadirles `Priority: models.PriorityMedium` o crear un helper `newTask(title)`.

### Condiciones de ejecución

- PostgreSQL accesible (`DB_HOST`, por defecto `127.0.0.1`) con la base `todo_test`.
- Migraciones de test aplicadas **incluida la nueva**: `soda migrate -e test`. Sin ella, los tests fallan con «column priority does not exist».

### Comandos

```sh
gofmt -l .            # sin salida
go vet ./...
soda migrate -e test
go test -p 1 ./...
go build ./...
```

### Validación manual

Arrancar la app (`/env:start`) y comprobar:

- Que las tareas del seed muestran «Media».
- Crear una tarea Alta.
- Pulsar «Prioridad» y ver el orden.
- Editar la prioridad.
- Probar con teclado: el foco es visible en el selector y en los enlaces.
- Revisar en un ancho de 380px.

## Riesgos y recuperación

- **Fallos en cadena en tests existentes por la nueva validación.** Todo `ValidateAndCreate` sin prioridad falla. Mitigación: ajustar esos tests (paso 6).
  - Alternativa considerada: tratar 0 como Media dentro del modelo con un hook `BeforeValidate`/`BeforeCreate`. Es menos explícito, pero reduce los cambios en los tests. Si quien revisa prefiere minimizar el diff de tests, se puede cambiar antes de implementar.
- **Bind de un valor no numérico.** Se mitiga con `form:"-"` y una conversión manual. Es incierto si `c.Bind` devolvería un error con `priority=abc`; el diseño lo evita en cualquier caso.
- **Orden de migraciones.** El timestamp `20261014000000` es posterior a los existentes. Si otra rama añade migraciones, hay que revisar que no colisionen.
- **Rollback.** `soda migrate down -e <env>` elimina la columna y la restricción. Se pierden las prioridades asignadas; no se pierden tareas. Para revertir el código basta con `git revert`. Si se revierte el código sin bajar la migración, la app sigue funcionando: la columna tiene default y pop solo lee y escribe las columnas del struct.
- **Producción.** En PostgreSQL 11 o superior, `ADD COLUMN … DEFAULT` constante no reescribe la tabla (fuente: notas de la versión 11 de PostgreSQL). La CHECK sí recorre la tabla, pero es trivial con este volumen. No he verificado la versión de PostgreSQL que se usa en producción; la de `compose.yaml` no la he comprobado.
- **El parámetro `sort` se pierde tras las acciones.** Es una limitación conocida (está en «Excluido») y puede sorprender al usuario: después de crear una tarea con el orden por prioridad activo, vuelve el orden por creación.
- **Accesibilidad.** La prioridad no depende solo del color (siempre hay texto). Hay que verificar el contraste de las variantes de color a mano.
- **Ambigüedad del encargo.** I1–I6 son inferencias. La más discutible es I4: el usuario podría querer que la lista se ordene **siempre** por prioridad. Si es así, basta con cambiar `defaultTaskOrder` y eliminar la navegación.
