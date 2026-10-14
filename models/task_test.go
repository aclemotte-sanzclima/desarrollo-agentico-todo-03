package models

import "strings"

func (ms *ModelSuite) Test_Task_Create_Valid() {
	task := &Task{Title: "Comprar café", Priority: PriorityMedium}

	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	ms.NotZero(task.ID)
	ms.False(task.Completed)

	count, err := ms.DB.Count(&Task{})
	ms.NoError(err)
	ms.Equal(1, count)
}

func (ms *ModelSuite) Test_Task_Title_Required() {
	task := &Task{Title: ""}

	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.True(verrs.HasAny())
	ms.Contains(verrs.Get("title"), "El título es obligatorio.")

	count, err := ms.DB.Count(&Task{})
	ms.NoError(err)
	ms.Equal(0, count)
}

func (ms *ModelSuite) Test_Task_Title_OnlySpaces_Invalid() {
	task := &Task{Title: "   \t  "}

	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.True(verrs.HasAny())
	ms.Contains(verrs.Get("title"), "El título es obligatorio.")
}

func (ms *ModelSuite) Test_Task_Title_MaxLength() {
	task := &Task{Title: strings.Repeat("a", TitleMaxLength), Priority: PriorityMedium}

	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.False(verrs.HasAny())
}

func (ms *ModelSuite) Test_Task_Title_TooLong_Invalid() {
	task := &Task{Title: strings.Repeat("a", TitleMaxLength+1)}

	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.True(verrs.HasAny())
	ms.NotEmpty(verrs.Get("title"))
}

func (ms *ModelSuite) Test_Task_Toggle_Persists() {
	task := &Task{Title: "Revisar el correo", Priority: PriorityMedium}
	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	task.Completed = true
	ms.NoError(ms.DB.Update(task))

	reloaded := &Task{}
	ms.NoError(ms.DB.Find(reloaded, task.ID))
	ms.True(reloaded.Completed)
}

func (ms *ModelSuite) Test_Task_Priority_ValidValues() {
	for _, priority := range []int{PriorityLow, PriorityMedium, PriorityHigh} {
		task := &Task{Title: "Tarea con prioridad", Priority: priority}
		verrs, err := ms.DB.ValidateAndCreate(task)
		ms.NoError(err)
		ms.False(verrs.HasAny())

		reloaded := &Task{}
		ms.NoError(ms.DB.Find(reloaded, task.ID))
		ms.Equal(priority, reloaded.Priority)
	}
}

func (ms *ModelSuite) Test_Task_Priority_Invalid() {
	for _, priority := range []int{0, 4, -1} {
		task := &Task{Title: "Tarea sin prioridad válida", Priority: priority}
		verrs, err := ms.DB.ValidateAndCreate(task)
		ms.NoError(err)
		ms.True(verrs.HasAny())
		ms.Contains(verrs.Get("priority"), "La prioridad no es válida.")
	}

	count, err := ms.DB.Count(&Task{})
	ms.NoError(err)
	ms.Equal(0, count)
}

func (ms *ModelSuite) Test_Task_PriorityLabel() {
	ms.Equal("Baja", Task{Priority: PriorityLow}.PriorityLabel())
	ms.Equal("Media", Task{Priority: PriorityMedium}.PriorityLabel())
	ms.Equal("Alta", Task{Priority: PriorityHigh}.PriorityLabel())
	ms.Equal("", Task{Priority: 9}.PriorityLabel())
}

func (ms *ModelSuite) Test_Task_Priority_DBDefault() {
	ms.NoError(ms.DB.RawQuery(
		"INSERT INTO tasks (id, title, completed, created_at, updated_at) VALUES ('5b6c2f0e-3f1a-4c8e-9a51-2d7e1c0b9f42', 'Sin prioridad', false, now(), now())",
	).Exec())

	task := &Task{}
	ms.NoError(ms.DB.First(task))
	ms.Equal(PriorityMedium, task.Priority)
}
