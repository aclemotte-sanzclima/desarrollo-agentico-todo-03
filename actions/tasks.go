package actions

import (
	"net/http"
	"strconv"
	"strings"

	"todo/models"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
)

// taskOrders mapea el parámetro sort a cláusulas ORDER BY fijas (lista blanca).
var taskOrders = map[string]string{
	"priority": "priority desc, created_at asc, id asc",
}

const defaultTaskOrder = "created_at asc, id asc"

// listTasks carga las tareas en el orden pedido y devuelve el sort normalizado.
func listTasks(tx *pop.Connection, sort string) (models.Tasks, string, error) {
	order, ok := taskOrders[sort]
	if !ok {
		order, sort = defaultTaskOrder, ""
	}

	tasks := models.Tasks{}
	if err := tx.Order(order).All(&tasks); err != nil {
		return nil, "", err
	}
	return tasks, sort, nil
}

// parsePriority convierte el valor del formulario; vacío conserva fallback y
// un valor no numérico devuelve -1 para que la validación lo rechace.
func parsePriority(raw string, fallback int) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback
	}
	priority, err := strconv.Atoi(raw)
	if err != nil {
		return -1
	}
	return priority
}

// TasksIndex muestra la lista de tareas y el formulario de creación.
func TasksIndex(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	tasks, sort, err := listTasks(tx, c.Param("sort"))
	if err != nil {
		return err
	}

	c.Set("tasks", tasks)
	c.Set("sort", sort)
	c.Set("priorities", models.PriorityOptions)
	c.Set("task", &models.Task{Priority: models.PriorityMedium})
	c.Set("errors", validate.NewErrors())
	return c.Render(http.StatusOK, r.HTML("tasks/index.plush.html"))
}

// TasksCreate crea una tarea nueva a partir del formulario.
func TasksCreate(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	if err := c.Bind(task); err != nil {
		return err
	}
	task.Title = strings.TrimSpace(task.Title)
	task.Priority = parsePriority(c.Param("priority"), models.PriorityMedium)
	if task.Priority == 0 {
		task.Priority = models.PriorityMedium
	}

	verrs, err := tx.ValidateAndCreate(task)
	if err != nil {
		return err
	}

	if verrs.HasAny() {
		tasks, sort, err := listTasks(tx, "")
		if err != nil {
			return err
		}

		c.Set("tasks", tasks)
		c.Set("sort", sort)
		c.Set("priorities", models.PriorityOptions)
		c.Set("task", task)
		c.Set("errors", verrs)
		return c.Render(http.StatusUnprocessableEntity, r.HTML("tasks/index.plush.html"))
	}

	c.Flash().Add("success", "Tarea creada.")
	return c.Redirect(http.StatusSeeOther, "/")
}

// TasksEdit muestra el formulario de edición de una tarea existente.
func TasksEdit(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	if err := tx.Find(task, c.Param("task_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	c.Set("task", task)
	c.Set("priorities", models.PriorityOptions)
	c.Set("errors", validate.NewErrors())
	return c.Render(http.StatusOK, r.HTML("tasks/edit.plush.html"))
}

// TasksUpdate actualiza el título y la prioridad, conservando los demás datos de la tarea.
func TasksUpdate(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	if err := tx.Find(task, c.Param("task_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	input := struct {
		Title string `form:"title" json:"title"`
	}{}
	if err := c.Bind(&input); err != nil {
		return err
	}
	task.Title = strings.TrimSpace(input.Title)
	task.Priority = parsePriority(c.Param("priority"), task.Priority)

	verrs, err := tx.ValidateAndUpdate(task)
	if err != nil {
		return err
	}
	if verrs.HasAny() {
		c.Set("task", task)
		c.Set("priorities", models.PriorityOptions)
		c.Set("errors", verrs)
		return c.Render(http.StatusUnprocessableEntity, r.HTML("tasks/edit.plush.html"))
	}

	c.Flash().Add("success", "Tarea actualizada.")
	return c.Redirect(http.StatusSeeOther, "/")
}

// TasksToggle alterna una tarea entre completada y pendiente.
func TasksToggle(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	if err := tx.Find(task, c.Param("task_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	task.Completed = !task.Completed
	if err := tx.Update(task); err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/")
}

// TasksDestroy elimina una tarea.
func TasksDestroy(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	if err := tx.Find(task, c.Param("task_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if err := tx.Destroy(task); err != nil {
		return err
	}

	c.Flash().Add("success", "Tarea eliminada.")
	return c.Redirect(http.StatusSeeOther, "/")
}
