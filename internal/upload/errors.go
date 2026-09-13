package upload

import "errors"

var (
    ErrInvalidPath      = errors.New("INVALID_PATH")
    ErrPathTraversal    = errors.New("PATH_TRAVERSAL")
    ErrAbsolutePath     = errors.New("ABSOLUTE_PATH")
    ErrSymlink          = errors.New("SYMLINK")
    ErrFileTooLarge     = errors.New("FILE_TOO_LARGE")
    ErrTotalTooLarge    = errors.New("TOTAL_SIZE_EXCEEDED")
    ErrTooManyFiles     = errors.New("FILE_COUNT_EXCEEDED")
    ErrCompressionBomb  = errors.New("COMPRESSION_BOMB")
    ErrNestedArchive    = errors.New("NESTED_ARCHIVE")
    ErrInvalidZIP       = errors.New("INVALID_ZIP")
)
