package core

type EnableStatus int
type DeleteStatus int

const (
	EnableStatusTrue  EnableStatus = 0x01
	EnableStatusFalse EnableStatus = 0xff
	HasBeenDeleted    DeleteStatus = 0x01
	HasNotDeleted     DeleteStatus = 0x00
)
