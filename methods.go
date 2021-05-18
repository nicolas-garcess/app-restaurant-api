type allTasks []task

var tasks = allTasks{
	{
		ID:      1,
		Name:    "Task 1",
		Content: "Content 1",
		Date:    "May 2021",
	},
	{
		ID:      2,
		Name:    "Task 2",
		Content: "Content 2",
		Date:    "June 2021",
	},
}

func getTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func createTask(w http.ResponseWriter, r *http.Request) {
	var newTask task
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Insert a valid task")
	}

	json.Unmarshal(reqBody, &newTask)
	newTask.ID = len(tasks) + 1
	tasks = append(tasks, newTask)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tasks)
}

func getTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	taskID, err := strconv.Atoi(id)
	if err != nil {
		fmt.Fprintf(w, "Invalid ID")
		return
	}

	for _, task := range tasks {
		if task.ID == taskID {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(task)
		}
	}
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	taskID, err := strconv.Atoi(id)
	if err != nil {
		fmt.Fprintf(w, "Invalid ID")
		return
	}

	for index, task := range tasks {
		if task.ID == taskID {
			tasks = append(tasks[:index], tasks[index+1:]...)
			fmt.Fprintf(w, "The task with ID %v has removed succesfully", taskID)
		}
	}
}

func updateTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	taskID, err := strconv.Atoi(id)

	if err != nil {
		fmt.Fprintf(w, "Invalid ID")
		return
	}

	var updateTask task
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Insert a valid task")
	}

	json.Unmarshal(reqBody, &updateTask)

	for index, task := range tasks {
		if task.ID == taskID {
			tasks = append(tasks[:index], tasks[index+1:]...)
			updateTask.ID = taskID
			tasks = append(tasks[:index+1], tasks[index:]...)
			tasks[index] = updateTask
			fmt.Fprintf(w, "The task with ID %v has updated succesfully", taskID)
		}
	}
	fmt.Fprintf(w, "The task with ID %v does not exist", taskID)
}