package http

type paginationQuery struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

type relationQuery struct {
	ParentID string `form:"parent_id" binding:"required"`
	ChildID  string `form:"child_id" binding:"required"`
}
