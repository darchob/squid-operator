package squid

const (
	noDiffError = "no diff with incomming rules and currents"
)

func IsRulesExistError(err error) bool {
	return err.Error() != noDiffError
}
