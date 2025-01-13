package identityaccess_test

import (
	"errors"
	"testing"

	"github.com/jm96441n/movieswithfriends/identityaccess"
	"github.com/jm96441n/movieswithfriends/testhelpers"
)

func TestSignupReq_Validate(t *testing.T) {
	testCases := map[string]struct {
		req  *identityaccess.SignupReq
		want error
	}{
		"validRequestNoPartyID": {
			req: &identityaccess.SignupReq{
				Email:     "email@email.com",
				Password:  "1Password",
				FirstName: "FirstName",
				LastName:  "Lastname",
			},
			want: nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			tt.Parallel()
			got := tc.req.Validate()
			testhelpers.Assert(t, errors.Is(got, tc.want), "expected %v, got %v", got, tc.want)
		})
	}
}
