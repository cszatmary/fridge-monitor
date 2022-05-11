package apierror

import (
	"fmt"
	"strings"
)

type Error interface {
	error
	Code() Code
	Message() string
}

type DetailedError interface {
	Error
	Details() string
}

type Code uint8

const (
	CodeUnknown Code = iota
	CodeDatabase
	CodeRecordNotFound
	CodeInvalidParameter
)

func (c Code) String() string {
	switch c {
	case CodeDatabase:
		return "err_database"
	case CodeRecordNotFound:
		return "err_record_not_found"
	case CodeInvalidParameter:
		return "err_invalid_parameter"
	default:
		return "err_unknown"
	}
}

type Op string

func (op Op) String() string {
	if op == "" {
		return "unknown"
	}
	return string(op)
}

type standardError struct {
	code Code
	msg  string
	op   Op
	err  error
}

func New(code Code, msg string, op Op) error {
	return &standardError{code, msg, op, nil}
}

func Wrap(err error, code Code, msg string, op Op) error {
	return &standardError{code, msg, op, err}
}

func (se *standardError) Error() string {
	s := fmt.Sprintf("%s: %s", se.code, se.msg)
	if se.err != nil {
		s = fmt.Sprintf("%s: %v", s, se.err)
	}
	return s
}

func (se *standardError) Code() Code {
	return se.code
}

func (se *standardError) Message() string {
	return se.msg
}

func (se *standardError) Details() string {
	var sb strings.Builder
	sb.WriteString(se.op.String())
	sb.WriteString(": ")
	sb.WriteString(se.code.String())
	sb.WriteString(": ")
	sb.WriteString(se.msg)
	if se.err != nil {
		if prevErr, ok := se.err.(DetailedError); ok {
			sb.WriteString(":\n\t")
			sb.WriteString(prevErr.Details())
		} else {
			sb.WriteString(": ")
			sb.WriteString(se.err.Error())
		}
	}
	return sb.String()
}

func (se *standardError) Unwrap() error {
	return se.err
}
