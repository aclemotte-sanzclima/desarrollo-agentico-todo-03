# Feature: Paleta de la interfaz en tonos azules

Issue: [#3](https://github.com/aclemotte-sanzclima/desarrollo-agentico-todo-03/issues/3) «Cambiar la paleta de la interfaz a tonos azules».

## Encargo y criterios de aceptación

Encargo literal: «Cambia todo a tonos azules».

### Requisitos explícitos (issue)

- R1. El color principal y sus variantes (hover y fondos suaves) pasan a tonos azules.
- R2. Los textos y controles mantienen un contraste mínimo WCAG AA.
- R3. El color de foco se sigue distinguiendo del color principal.
- R4. No cambia el comportamiento de la aplicación.

### Inferencias (decisiones del plan, revisables)

- I1. «Todo» incluye los **neutros**: los grises cálidos (texto, texto atenuado, bordes, fondos de barra lateral, botones e iconos) pasan a grises azulados, para que la interfaz no mezcle un acento azul con neutros cálidos.
- I2. Los **colores semánticos no cambian**:
  - el rojo de peligro y error (`--color-danger`, `.field-error`, `.icon-danger:hover`, el fondo de `input[aria-invalid]`);
  - el verde de éxito (`.flash-success`).

  Pasarlos a azul eliminaría el significado de «error» o «éxito», y el criterio R4 pide no cambiar el comportamiento percibido. Es la inferencia más discutible: si el encargo quería literalmente todo, ver «Riesgos».
- I3. El foco, que hoy es azul (`#245ab4`) y se confundiría con el nuevo primario, pasa a **azul marino muy oscuro** (`#0b1f44`). Sigue siendo azul, pero se distingue del primario por luminancia, y el anillo se dibuja con `outline-offset: 3px` sobre fondo blanco.
- I4. El blanco puro del fondo y de los campos se mantiene.

### Criterios de aceptación

1. `public/assets/app.css` no contiene ninguno de los tonos rojizos o cálidos de acento sustituidos: `#b73526`, `#992b1e`, `#f7e9e4`, `#a03123` y `#f2ded6`.
2. Los neutros cálidos listados en la tabla quedan sustituidos por sus equivalentes azulados.
3. Contraste (WCAG 2.1, calculado con la fórmula de luminancia relativa):
   - texto normal ≥ 4.5:1 sobre su fondo;
   - bordes de controles e indicadores gráficos ≥ 3:1 sobre blanco.
4. El anillo de foco tiene un contraste ≥ 3:1 sobre blanco y un tono claramente distinto del primario.
5. Los colores de error y éxito no cambian.
6. Los tests y el build existentes siguen en verde, sin cambios en Go, templates ni JS.

## Comportamiento actual y propuesta

### Actual (verificado)

- Todos los colores están en `public/assets/app.css`. Ni `templates/` ni `app.js` tienen colores: el SVG usa `currentColor`.
- Tokens de `:root` (líneas 17–27): `--color-text #282826`, `--color-muted #686864`, `--color-border #e9e9e5`, `--color-sidebar #f7f7f5`, `--color-primary #b73526`, `--color-primary-hover #992b1e`, `--color-focus #245ab4`, `--color-danger #a52b25`.
- Además hay colores literales repartidos por el archivo, sin tokens: navegación, hover del toggle, iconos, inputs, botones, errores y flash.

### Propuesta

Cambiar solo valores de color en `app.css`: primero los tokens de `:root` y después los literales cálidos de acento y neutros. No se crean archivos ni se cambia la estructura. Tampoco se introducen tokens nuevos: se respeta el estilo actual, que mezcla tokens y literales, para mantener el diff mínimo.

## Alcance

### Incluido

- Valores de color de `public/assets/app.css` según la tabla de «Diseño por capas».

### Excluido

- Colores semánticos de error, peligro y éxito (I2).
- Tipografía, espaciado, layout y modo oscuro.
- Refactorizar los literales a tokens.
- Templates, Go y JavaScript.
- La etiqueta `.priority-high` del PR #1 (prioridades), que todavía no está en `main`; ver «Riesgos».

## Diseño por capas

### CSS (`public/assets/app.css`)

Sustituciones. Contrastes calculados con la fórmula WCAG 2.1 para este plan.

| Línea (aprox.) | Selector / token | Actual | Nuevo | Contraste comprobado |
|---|---|---|---|---|
| 18 | `--color-text` | `#282826` | `#1e2a3b` | 14.49 sobre `#fff`; 13.60 sobre la barra lateral |
| 19 | `--color-muted` | `#686864` | `#5b6b80` | 5.44 sobre `#fff`; 5.11 sobre la barra lateral; 4.84 sobre `#eef2f7` |
| 20 | `--color-border` | `#e9e9e5` | `#e3e8ef` | decorativo |
| 21 | `--color-sidebar` | `#f7f7f5` | `#f5f8fc` | — |
| 22 | `--color-primary` | `#b73526` | `#1d5fbf` | blanco sobre él: 6.10; él sobre blanco: 6.10 |
| 23 | `--color-primary-hover` | `#992b1e` | `#174c99` | blanco sobre él: 8.30 |
| 24 | `--color-focus` | `#245ab4` | `#0b1f44` | 16.23 sobre `#fff` |
| 25 | `--color-danger` | `#a52b25` | sin cambio (I2) | — |
| 48 | `.sidebar` border-right | `#efefeb` | `#e8edf3` | decorativo |
| 77 | `.nav-link` background | `#f7e9e4` | `#e8f0fb` | — |
| 78 | `.nav-link` color | `#a03123` | `#1a4f9c` | 6.92 sobre `#e8f0fb` |
| 85 | `.nav-link:hover` background | `#f2ded6` | `#dbe7f8` | texto 6.36 |
| 164 | `.completion-icon` border | `#858580` | `#7d8a9c` | 3.51 sobre `#fff` |
| 177 | `.task-toggle-button:hover .completion-icon` background | `#f7e9e4` | `#e8f0fb` | primario sobre él: 5.31 |
| 189 | `.task-completed .task-title` text-decoration-color | `#96968f` | `#8e99a8` | decorativo |
| 193–194 | `.task-completed .completion-icon` background y border | `#777771` | `#6b778a` | 4.53 sobre `#fff` |
| 218 | `.icon-action:hover` background | `#f2f2ee` | `#eef2f7` | muted 4.84 |
| 256 | `input[type="text"]` border | `#92928a` | `#8593a6` | 3.12 sobre `#fff` (≥ 3:1 para controles) |
| 280 | `.button` border | `#dadad4` | `#d3dbe6` | decorativo (el texto identifica el botón) |
| 282 | `.button` color | `#41413d` | `#2f3b4d` | 10.45 sobre `#f3f6fa` |
| 283 | `.button` background | `#f6f6f2` | `#f3f6fa` | — |
| 294 | `.button:hover` background | `#eaeae4` | `#e6ecf4` | texto 9.53 |

Sin cambios (I2/I4):

- `#ffffff`;
- `.icon-danger:hover` (`#fcf0ec`);
- `input[aria-invalid]` (`#fff9f7`);
- `.field-error` (`#912a20` / `#fff0eb`);
- `.flash-success` (`#245131` / `#eef5e9` / `#d5e4ce`).

Los selectores que usan `var(--color-primary)` heredan el cambio sin tocarlos: `.brand-dot`, `.add-mark`, `.button-primary` y el hover del toggle.

## Archivos relevantes

| Ruta | Responsabilidad |
|---|---|
| `public/assets/app.css` | Único archivo que cambia: tokens de `:root` y literales de color |
| `templates/application.plush.html` | Solo lectura: marca, navegación y SVG con `currentColor`; no cambia |
| `templates/tasks/*.plush.html` | Solo lectura: sin colores en línea; no cambian |
| `public/embed.go` | Embebe los assets; un `go build` valida que el CSS se sigue empaquetando |

## Implementación paso a paso

1. Sustituir en `:root` los siete tokens indicados; `--color-danger` no se toca.
2. Sustituir los literales de la tabla, uno a uno, comprobando el selector. Los números de línea son orientativos.
3. Verificar con `grep -nE '#(b73526|992b1e|f7e9e4|a03123|f2ded6|282826|686864|e9e9e5|f7f7f5|245ab4|efefeb|858580|96968f|777771|f2f2ee|92928a|dadad4|41413d|f6f6f2|eaeae4)' public/assets/app.css` que no queda ninguno. La salida debe estar vacía.
4. Verificar que los colores semánticos siguen presentes: `grep -nE '#(a52b25|912a20|fff0eb|245131|eef5e9|d5e4ce|fcf0ec|fff9f7)' public/assets/app.css`.
5. Validar (sección siguiente).

## Tests y validación

No se añaden tests de Go: el cambio es solo de valores CSS y la suite no cubre estilos. Un test que busque cadenas hexadecimales en el CSS duplicaría el `grep` sin aportar garantías.

### Comprobaciones automáticas

```sh
gofmt -l .            # sin salida (no cambia Go; confirma que no hay efectos colaterales)
go vet ./...
go test -p 1 ./...    # necesita PostgreSQL y migraciones de test (soda migrate -e test)
go build ./...        # el CSS sigue embebido
```

Más los dos `grep` de los pasos 3 y 4.

### Contraste

Recalcular con la fórmula WCAG 2.1 (luminancia relativa) cada par de la tabla. Los valores esperados son los de la columna «Contraste comprobado». Basta con un script ad hoc fuera del repo.

### Validación manual

Con la app levantada (`/env:start`):

- la barra lateral y la navegación activa en azul;
- botón «Añadir» azul y hover más oscuro;
- marca «.» y «+» en azul;
- hover del círculo de completar en azul claro;
- tarea completada en gris azulado;
- foco con teclado (Tab) en azul marino, visible en botones, enlaces e inputs;
- el error de título vacío sigue en rojo;
- el flash de éxito sigue en verde.

## Riesgos y recuperación

- **Interpretación de «todo».** I2 mantiene el rojo de error y el verde de éxito. Si el encargo exigía literalmente todo en azul, habría que cambiarlos y diferenciar los estados por iconos o texto, porque el color dejaría de transmitir significado. Conviene confirmarlo en la revisión del PR.
- **Interacción con el PR #1 (prioridades), aún sin fusionar.** Ese PR añade `.priority-high` con `#992b1e` sobre `#fcf0ec`, en rojo, y modifica `app.css` en otras zonas.
  - Al fusionar ambos puede haber conflictos de texto en `app.css`.
  - Además, «Alta» quedaría en rojo en una interfaz azul.
  - Decisión pendiente para quien fusione en segundo lugar: mantener el rojo como color semántico de urgencia (coherente con I2) o pasarlo a `--color-primary`.
- **Percepción del foco.** El foco azul marino sobre botones azules se distingue por luminancia y por el margen blanco, no por el tono. Si en la revisión manual no resulta evidente, la alternativa es un color de foco de otro tono, por ejemplo ámbar `#b45309` (5.0:1 sobre blanco), a costa de romper el «todo azul».
- **Recuperación.** El cambio está en un solo archivo y no hay migraciones ni datos: se revierte con `git revert`.
