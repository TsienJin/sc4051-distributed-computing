package test_response

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"server/internal/rpc/response"
)

type ResponseValidator func(response *response.Response) error

// PacketMustPassAll is a function to aggregate multiple validators together
func PacketMustPassAll(rv ...ResponseValidator) ResponseValidator {
	return func(r *response.Response) error {
		for _, v := range rv {
			if err := v(r); err != nil {
				return err
			}
		}
		return nil
	}
}

// PacketMustPassAny will "pass" if any of the provided validators pass
func PacketMustPassAny(rv ...ResponseValidator) ResponseValidator {
	return func(r *response.Response) error {
		var err error
		failCount := 0
		for i, v := range rv {
			if err := v(r); err != nil {
				err = fmt.Errorf("[%v], %v", i, err)
				failCount++
			}
		}
		if failCount == len(rv) {
			return err
		}
		return nil
	}
}

// PacketMustNot is the equivalent to a negation of the ResponseValidator
func PacketMustNot(rv ResponseValidator) ResponseValidator {
	return func(r *response.Response) error {
		if err := rv(r); err == nil {
			return errors.New(fmt.Sprintf("response validator %v was supposed to fail", runtime.FuncForPC(reflect.ValueOf(rv).Pointer()).Name()))
		}
		return nil
	}
}

func BeStatus(status response.StatusCode) ResponseValidator {
	return func(r *response.Response) error {
		if r.StatusCode != status {
			return errors.New(fmt.Sprintf("status code does not match, E: %v, R: %v", status, r.StatusCode))
		}
		return nil
	}
}
