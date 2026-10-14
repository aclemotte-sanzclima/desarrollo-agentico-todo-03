package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

// TitleMaxLength es el número máximo de caracteres permitido en el título.
const TitleMaxLength = 200

// Prioridades posibles de una tarea.
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
	{PriorityHigh, "Alta"},
	{PriorityMedium, "Media"},
	{PriorityLow, "Baja"},
}

// Task representa una tarea de la lista.
type Task struct {
	ID        uuid.UUID `json:"id" db:"id" form:"-"`
	Title     string    `json:"title" db:"title" form:"title"`
	Completed bool      `json:"completed" db:"completed" form:"-"`
	Priority  int       `json:"priority" db:"priority" form:"-"`
	CreatedAt time.Time `json:"created_at" db:"created_at" form:"-"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" form:"-"`
}

// Tasks es una colección de Task.
type Tasks []Task

// Validate se ejecuta en cada llamada a los métodos "pop.Validate*".
func (t *Task) Validate(tx *pop.Connection) (*validate.Errors, error) {
	verrs := validate.NewErrors()

	title := strings.TrimSpace(t.Title)
	if title == "" {
		verrs.Add("title", "El título es obligatorio.")
	} else if len([]rune(title)) > TitleMaxLength {
		verrs.Add("title", fmt.Sprintf("El título no puede superar los %d caracteres.", TitleMaxLength))
	}

	if t.PriorityLabel() == "" {
		verrs.Add("priority", "La prioridad no es válida.")
	}

	return verrs, nil
}

// PriorityLabel devuelve el nombre de la prioridad o "" si no es válida.
func (t Task) PriorityLabel() string {
	for _, p := range PriorityOptions {
		if p.Value == t.Priority {
			return p.Label
		}
	}
	return ""
}

// PriorityClass devuelve la clase CSS asociada a la prioridad.
func (t Task) PriorityClass() string {
	switch t.Priority {
	case PriorityHigh:
		return "priority-high"
	case PriorityMedium:
		return "priority-medium"
	case PriorityLow:
		return "priority-low"
	}
	return ""
}
