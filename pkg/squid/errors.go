package squid

func IsRulesExistError(err error) bool {
	return err.Error() != DuplicateError
}
