package sqlkata

import "errors"

// ErrVariableNotFound is returned by FindVariable when name is not defined
// on the query or any parent. Use errors.Is(err, ErrVariableNotFound).
var ErrVariableNotFound = errors.New("variable not found")
