package squid

const (
	DuplicateError = "found duplicate rule : "
)

func IsRulesExistError(err error) bool {
	return err.Error() != DuplicateError
}
