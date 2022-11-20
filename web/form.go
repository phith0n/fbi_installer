package web

type Sender3DSForm struct {
	Address string `json:"address" form:"address" binding:"required"`
	Name    string `json:"name" form:"name" binding:"required"`
}

type GameFile struct {
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	ModTime int64  `json:"mod_time"`
}
