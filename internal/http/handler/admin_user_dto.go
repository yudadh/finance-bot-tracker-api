package handler

type RegisterAdminUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
}

type RegisterAdminUserResponse struct {
	ID     uint64 `json:"id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type GetMeResponse struct {
	ID     uint64 `json:"id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type ListUsersQuery struct {
	PerPage  int    `form:"per_page,default=10" binding:"min=10"`
	Page   int    `form:"page,default=1" binding:"min=1"`
	Status string `form:"status" binding:"omitempty,oneof=active inactive blocked"`
	Search string `form:"search"`
}

type ListUsersResponse struct {
	ID               uint64 `json:"id"`
	TelegramID       int64  `json:"telegram_id"`
	TelegramUsername string `json:"telegram_username"`
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	LanguageCode     string `json:"language_code"`
	Timezone         string `json:"timezone"`
	Status           string `json:"status"`
}
