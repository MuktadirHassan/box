package cli

import "errors"

var ErrPurgeRequired = errors.New("refusing to delete box without --purge")
