// Package kong
package kong

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckErrorOfMissingChildren(t *testing.T) {
	var (
		missing []string
		err     error
	)

	defer func() {
		if e := recover(); e != nil {
			require.Equal(t, "fail=' error is not a type Err", e.(error).Error())
		}
	}()
	// Not error
	err = checkErrorOfMissingChildren(missing)
	require.NoError(t, err)
	// One
	missing = append(missing, "1")
	err = checkErrorOfMissingChildren(missing)
	require.Error(t, err)
	// Checking the error object, a Errors().Expected error should be returned
	if err.(Err).Anchor() != Errors().Expected("").Anchor() {
		t.Fatalf("Incorrect error object, expected Errors().Expected")
	}
	if err.(Err).Anchor() == Errors().ExpectedOneOf("").Anchor() {
		t.Fatalf("Incorrect error object, expected Errors().Expected")
	}
	// Less or equal five elements
	missing = append(missing, "2", "3", "4", "5")
	err = checkErrorOfMissingChildren(missing)
	require.Error(t, err)
	// Checking the error object, a Errors().ExpectedOneOf error should be returned
	if err.(Err).Anchor() != Errors().ExpectedOneOf("").Anchor() {
		t.Fatalf("Incorrect error object, expected Errors().ExpectedOneOf")
	}
	if err.(Err).Anchor() == Errors().Expected("").Anchor() {
		t.Fatalf("Incorrect error object, expected Errors().ExpectedOneOf")
	}
	// Check not present constant onyOther
	require.NotContains(t, err.Error(), onyOther)
	// More five elements
	missing = append(missing, "6")
	err = checkErrorOfMissingChildren(missing)
	require.Error(t, err)
	// And specified error
	if err.(Err).Anchor() != Errors().ExpectedOneOf("").Anchor() {
		t.Fatalf("Incorrect error object, expected Errors().ExpectedOneOf")
	}
	if err.(Err).Anchor() == Errors().Expected("").Anchor() {
		t.Fatalf("Incorrect error object, expected Errors().ExpectedOneOf")
	}
	// Check constant onyOther
	require.Contains(t, err.Error(), onyOther)
}
