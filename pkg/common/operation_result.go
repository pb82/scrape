package common

type OperationStatus string

const (
	OperationStatusSuccess OperationStatus = "success"
	OperationStatusFailed  OperationStatus = "failed"
)

type OperationResult struct {
	Status  OperationStatus
	Message string
}
