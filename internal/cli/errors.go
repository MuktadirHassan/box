package cli

import "errors"

var (
	ErrPurgeRequired     = errors.New("refusing to delete box without --purge")
	ErrSetupConfirmation = errors.New("review the resolved configuration and rerun with --yes to save it")
)
