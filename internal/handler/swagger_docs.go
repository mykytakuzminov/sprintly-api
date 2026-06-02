package handler

// ─────────────────────────────────────────────
// AUTH
// ─────────────────────────────────────────────

// Login godoc
// @Summary     Login
// @Description Authenticates a user with email and password. Returns a pair of JWT tokens.
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body domain.LoginInput true "Login credentials"
// @Success     200 {object} domain.AuthTokens "Tokens issued successfully"
// @Failure     400 "Invalid request body"
// @Failure     403 {object} handler.ErrorResponse "Wrong password"
// @Failure     404 {object} handler.ErrorResponse "User not found"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /auth/login [post]
func swaggerLogin() {}

// Logout godoc
// @Summary     Logout
// @Description Invalidates the provided refresh token.
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body domain.LogoutInput true "Refresh token to revoke"
// @Success     204 "Logged out successfully"
// @Failure     400 "Invalid request body"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /auth/logout [post]
func swaggerLogout() {}

// Refresh godoc
// @Summary     Refresh access token
// @Description Issues a new access token using a valid refresh token stored in Redis.
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body domain.RefreshInput true "Refresh token"
// @Success     200 {object} handler.RefreshResponse "New access token"
// @Failure     400 "Invalid request body"
// @Failure     401 {object} handler.ErrorResponse "Refresh token not found or expired"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /auth/refresh [post]
func swaggerRefresh() {}

// ─────────────────────────────────────────────
// USERS
// ─────────────────────────────────────────────

// Register godoc
// @Summary     Register a new user
// @Description Creates a new user account. Email must be unique.
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       body body domain.RegisterInput true "Registration data"
// @Success     201 {object} handler.UserResponse "User created successfully"
// @Failure     400 "Invalid request body or validation error"
// @Failure     409 {object} handler.ErrorResponse "Email already in use"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /users/register [post]
func swaggerRegister() {}

// Me godoc
// @Summary     Get current user
// @Description Returns the profile of the currently authenticated user.
// @Tags        users
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} handler.UserResponse "Current user data"
// @Failure     401 "Missing or invalid access token"
// @Failure     404 {object} handler.ErrorResponse "User not found"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /users/me [get]
func swaggerMe() {}

// ChangePassword godoc
// @Summary     Change password
// @Description Changes the password of the currently authenticated user. Requires the current password for verification.
// @Tags        users
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body domain.ChangePasswordInput true "Old and new passwords"
// @Success     204 "Password changed successfully"
// @Failure     400 "Invalid request body or validation error"
// @Failure     401 "Missing or invalid access token"
// @Failure     403 {object} handler.ErrorResponse "Old password is incorrect"
// @Failure     404 {object} handler.ErrorResponse "User not found"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /users/me/password [patch]
func swaggerChangePassword() {}

// ─────────────────────────────────────────────
// BOARDS
// ─────────────────────────────────────────────

// CreateBoard godoc
// @Summary     Create a board
// @Description Creates a new board owned by the authenticated user.
// @Tags        boards
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body domain.CreateBoardInput true "Board data"
// @Success     201 {object} domain.Board "Board created successfully"
// @Failure     400 "Invalid request body or validation error"
// @Failure     401 "Missing or invalid access token"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /boards [post]
func swaggerCreateBoard() {}

// GetAllBoards godoc
// @Summary     Get all boards
// @Description Returns all boards owned by the authenticated user.
// @Tags        boards
// @Produce     json
// @Security    BearerAuth
// @Success     200 {array} domain.Board "List of boards"
// @Failure     401 "Missing or invalid access token"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /boards [get]
func swaggerGetAllBoards() {}

// GetBoardByID godoc
// @Summary     Get board by ID
// @Description Returns a specific board by its unique identifier.
// @Tags        boards
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "Board ID (UUID)"
// @Success     200 {object} domain.Board "Board data"
// @Failure     400 "Invalid board ID format"
// @Failure     401 "Missing or invalid access token"
// @Failure     404 {object} handler.ErrorResponse "Board not found"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /boards/{id} [get]
func swaggerGetBoardByID() {}

// UpdateBoard godoc
// @Summary     Update a board
// @Description Updates the name and/or description of a board. Only the board owner can perform this action.
// @Tags        boards
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id   path string                  true "Board ID (UUID)"
// @Param       body body domain.UpdateBoardInput true "Updated board data"
// @Success     204 "Board updated successfully"
// @Failure     400 "Invalid board ID format or request body"
// @Failure     401 "Missing or invalid access token"
// @Failure     403 {object} handler.ErrorResponse "Caller is not the board owner"
// @Failure     404 {object} handler.ErrorResponse "Board not found"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /boards/{id} [patch]
func swaggerUpdateBoard() {}

// DeleteBoard godoc
// @Summary     Delete a board
// @Description Deletes a board by ID. Only the board owner can perform this action.
// @Tags        boards
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "Board ID (UUID)"
// @Success     204 "Board deleted successfully"
// @Failure     400 "Invalid board ID format"
// @Failure     401 "Missing or invalid access token"
// @Failure     403 {object} handler.ErrorResponse "Caller is not the board owner"
// @Failure     404 {object} handler.ErrorResponse "Board not found"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /boards/{id} [delete]
func swaggerDeleteBoard() {}

// ─────────────────────────────────────────────
// COLUMNS
// ─────────────────────────────────────────────

// CreateColumn godoc
// @Summary     Create a column
// @Description Creates a new column inside a board. Only the board owner can add columns.
// @Tags        columns
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       boardID path string                   true "Board ID (UUID)"
// @Param       body    body domain.CreateColumnInput true "Column data"
// @Success     201 {object} domain.Column "Column created successfully"
// @Failure     400 "Invalid board ID format, request body, or validation error"
// @Failure     401 "Missing or invalid access token"
// @Failure     403 {object} handler.ErrorResponse "Caller is not the board owner"
// @Failure     404 {object} handler.ErrorResponse "Board not found"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /boards/{boardID}/columns [post]
func swaggerCreateColumn() {}

// GetAllColumns godoc
// @Summary     Get all columns of a board
// @Description Returns all columns belonging to the specified board.
// @Tags        columns
// @Produce     json
// @Security    BearerAuth
// @Param       boardID path string true "Board ID (UUID)"
// @Success     200 {array} domain.Column "List of columns"
// @Failure     400 "Invalid board ID format"
// @Failure     401 "Missing or invalid access token"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /boards/{boardID}/columns [get]
func swaggerGetAllColumns() {}

// GetColumnByID godoc
// @Summary     Get column by ID
// @Description Returns a specific column by its unique identifier.
// @Tags        columns
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "Column ID (UUID)"
// @Success     200 {object} domain.Column "Column data"
// @Failure     400 "Invalid column ID format"
// @Failure     401 "Missing or invalid access token"
// @Failure     404 {object} handler.ErrorResponse "Column not found"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /columns/{id} [get]
func swaggerGetColumnByID() {}

// UpdateColumn godoc
// @Summary     Update a column
// @Description Updates the name and/or position of a column. Only the owner of the parent board can perform this action.
// @Tags        columns
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id   path string                   true "Column ID (UUID)"
// @Param       body body domain.UpdateColumnInput true "Updated column data"
// @Success     204 "Column updated successfully"
// @Failure     400 "Invalid column ID format, request body, or validation error"
// @Failure     401 "Missing or invalid access token"
// @Failure     403 {object} handler.ErrorResponse "Caller is not the board owner"
// @Failure     404 {object} handler.ErrorResponse "Column not found"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /columns/{id} [patch]
func swaggerUpdateColumn() {}

// DeleteColumn godoc
// @Summary     Delete a column
// @Description Deletes a column by ID. Only the owner of the parent board can perform this action.
// @Tags        columns
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "Column ID (UUID)"
// @Success     204 "Column deleted successfully"
// @Failure     400 "Invalid column ID format"
// @Failure     401 "Missing or invalid access token"
// @Failure     403 {object} handler.ErrorResponse "Caller is not the board owner"
// @Failure     404 {object} handler.ErrorResponse "Column not found"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /columns/{id} [delete]
func swaggerDeleteColumn() {}

// ─────────────────────────────────────────────
// TASKS
// ─────────────────────────────────────────────

// CreateTask godoc
// @Summary     Create a task
// @Description Creates a new task inside a column. Only the owner of the parent board can create tasks.
// @Tags        tasks
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       columnID path string                 true "Column ID (UUID)"
// @Param       body     body domain.CreateTaskInput true "Task data"
// @Success     201 {object} domain.Task "Task created successfully"
// @Failure     400 "Invalid column ID format, request body, or validation error"
// @Failure     401 "Missing or invalid access token"
// @Failure     403 {object} handler.ErrorResponse "Caller is not the board owner"
// @Failure     404 {object} handler.ErrorResponse "Column not found"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /columns/{columnID}/tasks [post]
func swaggerCreateTask() {}

// GetAllTasksByColumnID godoc
// @Summary     Get all tasks in a column
// @Description Returns all tasks belonging to the specified column.
// @Tags        tasks
// @Produce     json
// @Security    BearerAuth
// @Param       columnID path string true "Column ID (UUID)"
// @Success     200 {array} domain.Task "List of tasks"
// @Failure     400 "Invalid column ID format"
// @Failure     401 "Missing or invalid access token"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /columns/{columnID}/tasks [get]
func swaggerGetAllTasksByColumnID() {}

// GetAllTasksByUserID godoc
// @Summary     Get all tasks created by current user
// @Description Returns all tasks where owner_id matches the authenticated user.
// @Tags        tasks
// @Produce     json
// @Security    BearerAuth
// @Success     200 {array} domain.Task "List of tasks"
// @Failure     401 "Missing or invalid access token"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /users/me/tasks [get]
func swaggerGetAllTasksByUserID() {}

// GetAllTasksByAssigneeID godoc
// @Summary     Get all tasks assigned to current user
// @Description Returns all tasks where assignee_id matches the authenticated user.
// @Tags        tasks
// @Produce     json
// @Security    BearerAuth
// @Success     200 {array} domain.Task "List of assigned tasks"
// @Failure     401 "Missing or invalid access token"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /users/me/tasks/assigned [get]
func swaggerGetAllTasksByAssigneeID() {}

// GetTaskByID godoc
// @Summary     Get task by ID
// @Description Returns a specific task by its unique identifier.
// @Tags        tasks
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "Task ID (UUID)"
// @Success     200 {object} domain.Task "Task data"
// @Failure     400 "Invalid task ID format"
// @Failure     401 "Missing or invalid access token"
// @Failure     404 {object} handler.ErrorResponse "Task not found"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /tasks/{id} [get]
func swaggerGetTaskByID() {}

// UpdateTask godoc
// @Summary     Update a task
// @Description Updates task fields including column, assignee, name, description and due date.
//
//	The column_id can be used to move the task to another column.
//	Only the task owner (board owner) can update it.
//
// @Tags        tasks
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id   path string                 true "Task ID (UUID)"
// @Param       body body domain.UpdateTaskInput true "Updated task data"
// @Success     204 "Task updated successfully"
// @Failure     400 "Invalid task ID format, request body, or validation error"
// @Failure     401 "Missing or invalid access token"
// @Failure     403 {object} handler.ErrorResponse "Caller is not the task owner"
// @Failure     404 {object} handler.ErrorResponse "Task not found"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /tasks/{id} [patch]
func swaggerUpdateTask() {}

// DeleteTask godoc
// @Summary     Delete a task
// @Description Deletes a task by ID. Only the task owner (board owner) can perform this action.
// @Tags        tasks
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "Task ID (UUID)"
// @Success     204 "Task deleted successfully"
// @Failure     400 "Invalid task ID format"
// @Failure     401 "Missing or invalid access token"
// @Failure     403 {object} handler.ErrorResponse "Caller is not the task owner"
// @Failure     404 {object} handler.ErrorResponse "Task not found"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /tasks/{id} [delete]
func swaggerDeleteTask() {}
