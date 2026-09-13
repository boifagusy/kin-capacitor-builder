package generator

import "errors"

var (
    ErrInvalidNode    = errors.New("INVALID_NODE_PROJECT")
    ErrLaravelDetected = errors.New("LARAVEL_DETECTED")
    ErrUnknownProject = errors.New("UNKNOWN_PROJECT")
    ErrMissingBuildOut = errors.New("MISSING_BUILD_OUTPUT")
    ErrMissingIndex   = errors.New("MISSING_INDEX_HTML")
)
