package actions

import (
	"net/url"
	"strings"
	"time"

	"todo/models"

	"github.com/gofrs/uuid"
)

func (as *ActionSuite) Test_TasksIndex_EmptyState() {
	res := as.HTML("/").Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), "No hay tareas todavía")
}

func (as *ActionSuite) Test_TasksIndex_ListsTasks() {
	as.LoadFixture("tres tareas")

	res := as.HTML("/").Get()
	as.Equal(200, res.Code)

	body := res.Body.String()
	as.Contains(body, "Comprar material de oficina")
	as.Contains(body, "Llamar al gestor")
	as.Contains(body, "Publicar la oferta de empleo")
	as.Contains(body, "Completada")
}

func (as *ActionSuite) Test_TasksCreate_Valid() {
	res := as.HTML("/tasks").Post(&models.Task{Title: "  Nueva tarea de prueba  "})
	as.Equal(303, res.Code)
	as.Equal("/", res.Header().Get("Location"))

	task := &models.Task{}
	as.NoError(as.DB.First(task))
	as.Equal("Nueva tarea de prueba", task.Title)
	as.False(task.Completed)
	as.Equal(models.PriorityMedium, task.Priority)
}

func (as *ActionSuite) Test_TasksCreate_EmptyTitle_Invalid() {
	res := as.HTML("/tasks").Post(&models.Task{Title: ""})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título es obligatorio.")

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(0, count)
}

func (as *ActionSuite) Test_TasksCreate_OnlySpaces_Invalid() {
	res := as.HTML("/tasks").Post(&models.Task{Title: "     "})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título es obligatorio.")

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(0, count)
}

func (as *ActionSuite) Test_TasksToggle_CompletesAndReopens() {
	task := &models.Task{Title: "Pendiente de completar", Priority: models.PriorityMedium}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())

	res := as.HTML("/tasks/%s/toggle", task.ID).Put(url.Values{})
	as.Equal(303, res.Code)

	as.NoError(as.DB.Reload(task))
	as.True(task.Completed)

	res = as.HTML("/tasks/%s/toggle", task.ID).Put(url.Values{})
	as.Equal(303, res.Code)

	as.NoError(as.DB.Reload(task))
	as.False(task.Completed)
}

func (as *ActionSuite) Test_TasksToggle_PostWithMethodOverride() {
	task := &models.Task{Title: "Pendiente desde el cliente", Priority: models.PriorityMedium}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())

	res := as.HTML("/tasks/%s/toggle", task.ID).Post(url.Values{"_method": {"PUT"}})
	as.Equal(303, res.Code)

	as.NoError(as.DB.Reload(task))
	as.True(task.Completed)

	res = as.HTML("/tasks/%s/toggle", task.ID).Post(url.Values{"_method": {"PUT"}})
	as.Equal(303, res.Code)

	as.NoError(as.DB.Reload(task))
	as.False(task.Completed)
}

func (as *ActionSuite) Test_TasksToggle_NotFound() {
	id := uuid.Must(uuid.NewV4())
	res := as.HTML("/tasks/%s/toggle", id).Put(url.Values{})
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksToggle_InvalidID() {
	res := as.HTML("/tasks/no-es-un-uuid/toggle").Put(url.Values{})
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksDestroy() {
	task := &models.Task{Title: "Tarea que se elimina", Priority: models.PriorityMedium}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())

	res := as.HTML("/tasks/%s", task.ID).Delete()
	as.Equal(303, res.Code)

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(0, count)
}

func (as *ActionSuite) Test_TasksDestroy_NotFound() {
	id := uuid.Must(uuid.NewV4())
	res := as.HTML("/tasks/%s", id).Delete()
	as.Equal(404, res.Code)
}

func (as *ActionSuite) taskForEditing(completed bool) *models.Task {
	task := &models.Task{Title: "Título original", Completed: completed, Priority: models.PriorityMedium}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())
	as.NoError(as.DB.Reload(task))
	return task
}

func (as *ActionSuite) Test_TasksIndex_EditLink() {
	task := as.taskForEditing(false)
	res := as.HTML("/").Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), "/tasks/"+task.ID.String()+"/edit")
	as.Contains(res.Body.String(), "Editar tarea: Título original")
}

func (as *ActionSuite) Test_TasksEdit_PrefillsTitle() {
	task := as.taskForEditing(true)
	res := as.HTML("/tasks/%s/edit", task.ID).Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), `value="Título original"`)
	as.Contains(res.Body.String(), `value="PUT"`)
	as.Contains(res.Body.String(), "Guardar cambios")
	as.Contains(res.Body.String(), "Cancelar")
	as.NoError(as.DB.Reload(task))
	as.True(task.Completed)
}

func (as *ActionSuite) Test_TasksEdit_NotFound() {
	res := as.HTML("/tasks/%s/edit", uuid.Must(uuid.NewV4())).Get()
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksEdit_InvalidID() {
	res := as.HTML("/tasks/no-es-un-uuid/edit").Get()
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksEdit_EscapesTitle() {
	task := &models.Task{Title: `<img src=x onerror="alert(1)">`, Priority: models.PriorityMedium}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())
	res := as.HTML("/tasks/%s/edit", task.ID).Get()
	as.Equal(200, res.Code)
	as.NotContains(res.Body.String(), `<img src=x`)
	as.Contains(res.Body.String(), "&lt;img")
}

func (as *ActionSuite) Test_TasksUpdate_PersistsTitleOnly() {
	task := as.taskForEditing(true)
	id, createdAt, updatedAt := task.ID, task.CreatedAt, task.UpdatedAt
	res := as.HTML("/tasks/%s", id).Put(url.Values{
		"title": {"  Título actualizado  "}, "completed": {"false"},
		"id": {uuid.Must(uuid.NewV4()).String()}, "created_at": {"2000-01-01T00:00:00Z"},
	})
	as.Equal(303, res.Code)
	as.Equal("/", res.Header().Get("Location"))
	as.NoError(as.DB.Reload(task))
	as.Equal("Título actualizado", task.Title)
	as.Equal(models.PriorityMedium, task.Priority)
	as.Equal(id, task.ID)
	as.True(task.Completed)
	as.True(task.CreatedAt.Equal(createdAt))
	as.False(task.UpdatedAt.Before(updatedAt))
	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(1, count)
}

func (as *ActionSuite) Test_TasksUpdate_PendingTaskRemainsPending() {
	task := as.taskForEditing(false)
	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {"Nuevo título"}, "completed": {"true"}})
	as.Equal(303, res.Code)
	as.NoError(as.DB.Reload(task))
	as.Equal("Nuevo título", task.Title)
	as.False(task.Completed)
}

func (as *ActionSuite) Test_TasksUpdate_EmptyTitle() {
	task := as.taskForEditing(false)
	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {""}})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título es obligatorio.")
	as.Contains(res.Body.String(), `aria-describedby="title-error"`)
	as.NoError(as.DB.Reload(task))
	as.Equal("Título original", task.Title)
}

func (as *ActionSuite) Test_TasksUpdate_OnlySpaces() {
	task := as.taskForEditing(true)
	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {"   "}})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título es obligatorio.")
	as.NoError(as.DB.Reload(task))
	as.Equal("Título original", task.Title)
	as.True(task.Completed)
}

func (as *ActionSuite) Test_TasksUpdate_TitleTooLong() {
	task := as.taskForEditing(false)
	title := strings.Repeat("á", models.TitleMaxLength+1)
	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {title}})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título no puede superar los 200 caracteres.")
	as.Contains(res.Body.String(), title)
	as.NoError(as.DB.Reload(task))
	as.Equal("Título original", task.Title)
}

func (as *ActionSuite) Test_TasksUpdate_TitleAtLimit() {
	task := as.taskForEditing(false)
	title := strings.Repeat("á", models.TitleMaxLength)
	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {title}})
	as.Equal(303, res.Code)
	as.NoError(as.DB.Reload(task))
	as.Equal(title, task.Title)
}

func (as *ActionSuite) Test_TasksUpdate_NotFound() {
	res := as.HTML("/tasks/%s", uuid.Must(uuid.NewV4())).Put(url.Values{"title": {"Título nuevo"}})
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksUpdate_InvalidID() {
	res := as.HTML("/tasks/no-es-un-uuid").Put(url.Values{"title": {"Título nuevo"}})
	as.Equal(404, res.Code)
}

// taskWithPriority crea una tarea con la prioridad y la fecha de creación indicadas.
func (as *ActionSuite) taskWithPriority(title string, priority int, createdAt time.Time) *models.Task {
	task := &models.Task{Title: title, Priority: priority}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())
	as.NoError(as.DB.RawQuery("UPDATE tasks SET created_at = ? WHERE id = ?", createdAt, task.ID).Exec())
	return task
}

func (as *ActionSuite) assertOrder(body string, titles ...string) {
	last := -1
	for _, title := range titles {
		i := strings.Index(body, title)
		as.True(i > last, "%q debería aparecer después de la tarea anterior", title)
		last = i
	}
}

func (as *ActionSuite) Test_TasksCreate_WithPriority() {
	res := as.HTML("/tasks").Post(url.Values{"title": {"Urgente"}, "priority": {"3"}})
	as.Equal(303, res.Code)

	task := &models.Task{}
	as.NoError(as.DB.First(task))
	as.Equal("Urgente", task.Title)
	as.Equal(models.PriorityHigh, task.Priority)
}

func (as *ActionSuite) Test_TasksCreate_InvalidPriority() {
	for _, priority := range []string{"7", "abc"} {
		res := as.HTML("/tasks").Post(url.Values{"title": {"Tarea"}, "priority": {priority}})
		as.Equal(422, res.Code)
		as.Contains(res.Body.String(), "La prioridad no es válida.")
		as.Contains(res.Body.String(), `aria-describedby="priority-error"`)
	}

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(0, count)
}

func (as *ActionSuite) Test_TasksIndex_ShowsPriority() {
	base := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	as.taskWithPriority("Revisar contrato", models.PriorityHigh, base)
	as.taskWithPriority("Ordenar escritorio", models.PriorityLow, base.Add(time.Minute))

	res := as.HTML("/").Get()
	as.Equal(200, res.Code)
	body := res.Body.String()
	as.Contains(body, `class="task-priority priority-high"`)
	as.Contains(body, `class="task-priority priority-low"`)
	as.Contains(body, "Alta</span>")
	as.Contains(body, "Baja</span>")
}

func (as *ActionSuite) Test_TasksIndex_NewTaskDefaultsToMedium() {
	res := as.HTML("/").Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), `name="priority"`)
	as.Contains(res.Body.String(), `value="2" selected`)
}

func (as *ActionSuite) Test_TasksIndex_SortByPriority() {
	base := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	as.taskWithPriority("Ordenar escritorio", models.PriorityLow, base)
	as.taskWithPriority("Revisar contrato", models.PriorityHigh, base.Add(time.Minute))
	as.taskWithPriority("Responder correos", models.PriorityMedium, base.Add(2*time.Minute))

	res := as.HTML("/?sort=priority").Get()
	as.Equal(200, res.Code)
	body := res.Body.String()
	as.assertOrder(body, "Revisar contrato", "Responder correos", "Ordenar escritorio")
	as.Contains(body, `<a href="/?sort=priority" aria-current="true">`)
}

func (as *ActionSuite) Test_TasksIndex_SortByPriority_TieBreakByCreation() {
	base := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	as.taskWithPriority("Segunda urgente", models.PriorityHigh, base.Add(time.Hour))
	as.taskWithPriority("Primera urgente", models.PriorityHigh, base)

	res := as.HTML("/?sort=priority").Get()
	as.Equal(200, res.Code)
	as.assertOrder(res.Body.String(), "Primera urgente", "Segunda urgente")
}

func (as *ActionSuite) Test_TasksIndex_DefaultOrderUnchanged() {
	base := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	as.taskWithPriority("Ordenar escritorio", models.PriorityLow, base)
	as.taskWithPriority("Revisar contrato", models.PriorityHigh, base.Add(time.Minute))
	as.taskWithPriority("Responder correos", models.PriorityMedium, base.Add(2*time.Minute))

	for _, path := range []string{"/", "/?sort=foo%27%3B%20drop%20table%20tasks%3B--"} {
		res := as.HTML("%s", path).Get()
		as.Equal(200, res.Code)
		body := res.Body.String()
		as.assertOrder(body, "Ordenar escritorio", "Revisar contrato", "Responder correos")
		as.Contains(body, `<a href="/" aria-current="true">`)
	}

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(3, count)
}

func (as *ActionSuite) Test_TasksEdit_ShowsCurrentPriority() {
	task := as.taskWithPriority("Revisar contrato", models.PriorityHigh, time.Now())
	res := as.HTML("/tasks/%s/edit", task.ID).Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), `value="3" selected`)
	as.NotContains(res.Body.String(), `value="2" selected`)
}

func (as *ActionSuite) Test_TasksUpdate_ChangesPriority() {
	task := as.taskForEditing(false)
	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {"Título original"}, "priority": {"1"}})
	as.Equal(303, res.Code)
	as.NoError(as.DB.Reload(task))
	as.Equal(models.PriorityLow, task.Priority)
}

func (as *ActionSuite) Test_TasksUpdate_InvalidPriority() {
	task := as.taskForEditing(false)
	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {"Otro título"}, "priority": {"9"}})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "La prioridad no es válida.")
	as.NoError(as.DB.Reload(task))
	as.Equal("Título original", task.Title)
	as.Equal(models.PriorityMedium, task.Priority)
}
