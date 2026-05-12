package db

import (
	"errors"
	"strings"

	"connectrpc.com/connect"
	"github.com/datakit-dev/dtkt-integrations/browser/pkg/db/ent"
)

const entErrorPrefix = "ent: "

// ToConnectError converts an Ent error to an appropriate Connect error
func ToConnectError(err error) error {
	if err == nil {
		return nil
	} else if err, ok := err.(*connect.Error); ok {
		return err
	}

	// Handle Ent-specific errors
	switch {
	case ent.IsNotFound(err):
		return connect.NewError(connect.CodeNotFound, stripEntErrorPrefix(err))
	case ent.IsConstraintError(err):
		return connect.NewError(connect.CodeAlreadyExists, stripEntErrorPrefix(err))
	case ent.IsValidationError(err):
		return connect.NewError(connect.CodeInvalidArgument, stripEntErrorPrefix(err))
	case ent.IsNotSingular(err):
		return connect.NewError(connect.CodeFailedPrecondition, stripEntErrorPrefix(err))
	case ent.IsNotLoaded(err):
		return connect.NewError(connect.CodeFailedPrecondition, stripEntErrorPrefix(err))
	}

	// Handle SQLite errors by checking error message
	// ncruces/go-sqlite3 returns errors as strings
	errMsg := err.Error()
	if strings.Contains(errMsg, "UNIQUE constraint failed") || strings.Contains(errMsg, "constraint failed: UNIQUE") {
		return connect.NewError(connect.CodeAlreadyExists, stripEntErrorPrefix(err))
	}
	if strings.Contains(errMsg, "FOREIGN KEY constraint failed") {
		return connect.NewError(connect.CodeFailedPrecondition, stripEntErrorPrefix(err))
	}
	if strings.Contains(errMsg, "CHECK constraint failed") {
		return connect.NewError(connect.CodeInvalidArgument, stripEntErrorPrefix(err))
	}

	// Default to internal error
	return connect.NewError(connect.CodeInternal, stripEntErrorPrefix(err))
}

func stripEntErrorPrefix(err error) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()
	str, ok := strings.CutPrefix(errMsg, entErrorPrefix)
	if ok {
		return errors.New(str)
	}
	return err
}
