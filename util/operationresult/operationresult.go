package operationresult

var (
	SuccessfullyRemoved = New("SUCCESSFULLY_REMOVED", "the item successfully removed")
	SuccessfullyUpdated = New("SUCCESSFULLY_UPDATED", "the item successfuly updated")
	PasswordSuccessfullyChanged = New("PASSWORD_SUCCESSFULLY_CHANGE", "user password successfully changed")
)

type OperationResult struct {
	Message     string    `json:"message"`
	Description string    `json:"description"`
}

func New(message, description string) *OperationResult {
	return &OperationResult{message, description}
}