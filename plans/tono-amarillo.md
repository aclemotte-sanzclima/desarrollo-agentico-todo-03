# Feature: Tono del diseño en amarillo

Issue: [#6](https://github.com/aclemotte-sanzclima/desarrollo-agentico-todo-03/issues/6) «Cambiar el tono del diseño a amarillo».

## Encargo y criterios de aceptación

Encargo literal: «Cambia el tono del diseño a amarillo». La issue no incluye criterios.

### Requisito explícito

- R1. El tono (color de acento) del diseño pasa a ser amarillo.

### Inferencias (decisiones del plan, revisables)

- I1. **«Tono» se interpreta como el color de acento**, hoy el rojo teja `#b73526` y sus variantes:
  - botón principal y su hover;
  - navegación activa;
  - marca «todo.» y «+»;
  - hover del círculo de completar.

  Los neutros ya son cálidos (`#282826`, `#f7f7f5`, etc.) y armonizan con el amarillo, así que no cambian.
- I2. **Accesibilidad (WCAG 2.1 AA) como requisito implícito.** El amarillo tiene muy poca luminancia frente al blanco: `#f5c518` sobre blanco da 1.63:1. Por eso:
  - el amarillo se usa como **relleno**, con texto oscuro encima;
  - para texto o trazos sobre fondo claro se usa un **ocre oscuro** del mismo tono (`#8a6100`);
  - el texto blanco del botón principal pasa a `--color-text`.
- I3. Los **colores semánticos no cambian**: el rojo de error y peligro y el verde de éxito.
- I4. El **foco** sigue en azul (`#245ab4`): ya se distingue del amarillo, con 6.57:1 sobre blanco y 4.03:1 sobre el relleno amarillo.

### Criterios de aceptación

1. `public/assets/app.css` ya no contiene los tonos de acento rojizos: `#b73526`, `#992b1e`, `#f7e9e4`, `#a03123` y `#f2ded6`.
2. El botón «Añadir» tiene fondo amarillo y texto oscuro (≥ 4.5:1), y en hover un amarillo más intenso (≥ 4.5:1 con el texto).
3. La marca «.» y el «+» se ven en ocre, con ≥ 3:1 sobre su fondo (lo exige un elemento gráfico).
4. La navegación activa tiene fondo amarillo pálido y texto ocre oscuro (≥ 4.5:1).
5. Error, éxito y foco no cambian.
6. Los tests y el build existentes siguen en verde, sin cambios en Go, templates ni JS.

## Comportamiento actual y propuesta

### Actual (verificado en `main`)

- Todos los colores están en `public/assets/app.css`. Los templates y `app.js` no tienen colores: el SVG usa `currentColor`.
- `--color-primary` (`#b73526`) cumple dos papeles que con el rojo eran compatibles:
  - **relleno** con texto blanco encima (`.button-primary`);
  - **tinta** sobre fondo claro (`.brand-dot`, `.add-mark` y el hover de `.completion-icon`).
- Literales de acento: `.nav-link` (`#f7e9e4` / `#a03123`), `.nav-link:hover` (`#f2ded6`) y el hover del toggle (`#f7e9e4`).

### Propuesta

Separar los dos papeles, porque con amarillo ya no son compatibles:

- `--color-primary` pasa a ser el **relleno** amarillo.
- Un token nuevo, `--color-primary-ink`, es la **tinta** ocre para texto y trazos sobre fondo claro.

Es el único token nuevo y es imprescindible para cumplir I2. El resto son cambios de valor.

## Alcance

### Incluido

- `public/assets/app.css`: tokens de acento, literales de acento y color del texto del botón principal.

### Excluido

- Neutros (I1), colores semánticos (I3) y foco (I4).
- Templates, Go y JavaScript.
- Integración con los PR abiertos #1 (prioridades, con `.priority-high` en rojo) y #4 (paleta azul, que es incompatible con este cambio); ver «Riesgos».

## Diseño por capas

### CSS (`public/assets/app.css`)

Contrastes calculados con la fórmula WCAG 2.1 para este plan.

| Línea (aprox.) | Selector / token | Actual | Nuevo | Contraste comprobado |
|---|---|---|---|---|
| 22 | `--color-primary` | `#b73526` | `#f5c518` (relleno) | texto `#282826` sobre él: 9.06 |
| 23 | `--color-primary-hover` | `#992b1e` | `#e0ad00` | texto `#282826` sobre él: 7.14 |
| 23+ | `--color-primary-ink` (nuevo, tras `--color-primary-hover`) | — | `#8a6100` | 5.54 sobre `#fff`; 5.16 sobre `#f7f7f5` |
| 67 | `.brand-dot` color | `var(--color-primary)` | `var(--color-primary-ink)` | 5.16 sobre la barra lateral |
| 77 | `.nav-link` background | `#f7e9e4` | `#fdf3c4` | — |
| 78 | `.nav-link` color | `#a03123` | `#6b4e00` | 6.92 |
| 85 | `.nav-link:hover` background | `#f2ded6` | `#fbe8a0` | texto 6.32 |
| 177 | `.task-toggle-button:hover .completion-icon` background | `#f7e9e4` | `#fdf3c4` | — |
| 178–179 | ídem border-color y color | `var(--color-primary)` | `var(--color-primary-ink)` | 4.96 sobre `#fdf3c4` |
| 242 | `.add-mark` color | `var(--color-primary)` | `var(--color-primary-ink)` | 5.54 sobre `#fff` |
| 298–299 | `.button-primary` background / border-color | `var(--color-primary)` / `var(--color-primary)` | `var(--color-primary)` / **`var(--color-primary-ink)`** | el borde ocre da 5.54:1 frente al blanco y delimita el botón |
| 300 | `.button-primary` color | `#ffffff` | `var(--color-text)` | 9.06 |
| 304–305 | `.button-primary:hover` background / border-color | `var(--color-primary-hover)` ×2 | `var(--color-primary-hover)` / `var(--color-primary-ink)` | texto 7.14 |

Sin cambios:

- neutros;
- `--color-focus` y `--color-danger`;
- `.icon-danger:hover`, `.field-error`, `input[aria-invalid]` y `.flash-success`.

## Archivos relevantes

| Ruta | Responsabilidad |
|---|---|
| `public/assets/app.css` | Único archivo que cambia |
| `templates/application.plush.html`, `templates/tasks/*.plush.html` | Solo lectura: sin colores en línea |
| `public/embed.go` | Embebe los assets; `go build` confirma que el CSS se sigue empaquetando |

## Implementación paso a paso

1. En `:root`:
   - cambiar `--color-primary` y `--color-primary-hover`;
   - añadir `--color-primary-ink: #8a6100;` justo después de `--color-primary-hover`.
2. Cambiar `.brand-dot`, `.add-mark` y `border-color`/`color` del hover del toggle a `var(--color-primary-ink)`.
3. Sustituir los literales de `.nav-link`, `.nav-link:hover` y el fondo del hover del toggle.
4. En `.button-primary`:
   - `border-color: var(--color-primary-ink)`;
   - `color: var(--color-text)`.

   En `.button-primary:hover`: `border-color: var(--color-primary-ink)`.
5. Verificar que la salida de `grep -nE '#(b73526|992b1e|f7e9e4|a03123|f2ded6)' public/assets/app.css` está vacía.
6. Verificar que los colores semánticos y el foco siguen presentes: `grep -nE '#(a52b25|912a20|fff0eb|245131|eef5e9|d5e4ce|fcf0ec|fff9f7|245ab4)' public/assets/app.css`.
7. Comprobar que no queda texto blanco sobre amarillo: `grep -n -A3 '\.button-primary {' public/assets/app.css` no debe mostrar `#ffffff`.

## Tests y validación

No se añaden tests de Go: el cambio es solo de valores CSS y la suite no cubre estilos.

```sh
gofmt -l .            # sin salida; en un checkout con CRLF, evaluar sobre contenido LF
go vet ./...
go test -p 1 ./...    # necesita PostgreSQL y migraciones de test (soda migrate -e test)
go build ./...
```

Más los `grep` de los pasos 5–7 y el recálculo de los contrastes de la tabla con la fórmula WCAG 2.1, con un script ad hoc fuera del repo.

### Validación manual

Con la app levantada, comprobar:

- botón «Añadir» amarillo con texto oscuro y borde ocre;
- navegación activa en amarillo pálido;
- «.» y «+» en ocre;
- hover del círculo en amarillo pálido;
- foco azul visible;
- error en rojo y éxito en verde.

## Riesgos y recuperación

- **Interpretación de «tono».** Si se quería también el fondo o los neutros en amarillo, el texto de toda la interfaz necesitaría revisar el contraste. Conviene confirmarlo en la revisión del PR.
- **Contradicción con el PR #4 (paleta azul, de la issue #3).** Ambos cambian el acento de `app.css` y no pueden fusionarse los dos: habrá conflicto, y el resultado tiene que ser uno de los dos. Hay que decidir cuál se queda.
- **PR #1 (prioridades).** Añade `.priority-high` en rojo; al combinarlo, decidir si «Alta» sigue en rojo (semántico) o pasa a amarillo u ocre.
- **El amarillo sobre blanco destaca poco.** Por eso el botón lleva borde ocre. Si en la revisión manual el botón no destaca, la alternativa es un amarillo más saturado (`#f2b705`; el texto oscuro seguiría por encima de 7:1).
- **Recuperación.** Un solo archivo, sin datos ni migraciones: `git revert`.
