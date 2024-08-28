package functions

var (
	IsBroadcastMode        = make(map[int64]bool)
	MaterialStep           = make(map[int64]string)
	InProcessSubReq        = make(map[int64]bool)
	InProcessAdminReq      = make(map[int64]bool)
	IsDeleteSubscriberMode = false
	IsDeleteAdminMode      = false
	IsDeleteMaterialMode   = false
	IsDeleteSubjectMode    = false
)
