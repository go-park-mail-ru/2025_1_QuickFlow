package validation

const (
	TypeOutComing = "outcoming"
	TypeInComing  = "incoming"
	TypeAll       = "all"
)

var acceptedReqTypesSet = map[string]struct{}{
	TypeOutComing: {},
	TypeInComing:  {},
	TypeAll:       {},
}

func validateReqType(reqType string) bool {
	if _, ok := acceptedReqTypesSet[reqType]; !ok {
		return false
	}

	return true
}
