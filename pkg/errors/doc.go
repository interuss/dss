// Package errors provides types and functions supporting consistent error
// handling.
//
// Errors within the DSS are generally either plain Go errors, or else
// errors augmented with the stacktrace package which attaches information to
// errors as they propagate up the call stack.  Whenever returning an error due
// to a lower-level error, always stacktrace.Propagate it rather than returning
// it directly.  Also, when creating new errors, use stacktrace.NewError rather
// than errors.New so that information about the codebase location where the
// error was generated can be obtained by a debugger.
//
// When an error will likely result in a response with a given code if that
// error were bubbled up to the request handler, attach the code to the error
// with stacktrace.PropagateWithCode or stacktrace.NewErrorWithCode.  Higher-
// level calling functions can always override this code if appropriate.  The
// recognized codes are enumerated in errors.go of this package.
//
// Just before an error is ultimately returned by a request handler, the
// interceptor in errors.go logs the full details of the error and then
// replaces it with a simple error containing an ID that may be used to look up
// the full details in the logs.
package errors
