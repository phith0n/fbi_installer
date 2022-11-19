package web

type Sender3DSForm struct {
	Address string `json:"address" form:"address" binding:"required"`
	Name    string `json:"name" form:"name" binding:"required"`
}
