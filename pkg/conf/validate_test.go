package conf

import (
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestValidate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		res, err := validate([]byte(schema.BasicSchemaJSON), []byte(`{"maxReposToSearch":123}`))
		if err != nil {
			t.Fatal(err)
		}
		if len(res.Errors()) != 0 {
			t.Errorf("errors: %v", res.Errors())
		}
	})

	t.Run("invalid", func(t *testing.T) {
		res, err := validate([]byte(schema.BasicSchemaJSON), []byte(`{"a":1}`))
		if err != nil {
			t.Fatal(err)
		}
		if len(res.Errors()) == 0 {
			t.Error("want invalid")
		}
	})
}

func TestValidateCustomCore(t *testing.T) {
	raw := `{"auth.providers":[{"type":"asdf"}]}`
	wantErr := "tagged union type must have a"

	problems, err := validateCustomCoreRaw([]byte(raw))

	if err != nil {
		if !strings.Contains(err.Error(), wantErr) {
			t.Fatal(err)
		}
	}

	if len(problems) > 0 {
		t.Fatalf("unexpected problems: %v", problems)
	}
}

func TestValidateCustomBasic(t *testing.T) {
	tests := map[string]struct {
		raw         string
		wantProblem string
	}{
		// username is optional; password and token are disjointly required
		"bitbucketserver no auth": {
			raw:         `{"bitbucketServer":[{}]}`,
			wantProblem: "specify either a token or a username/password",
		},
		"bitbucketserver password and token": {
			raw:         `{"bitbucketServer":[{"password":"p","token":"t"}]}`,
			wantProblem: "specify either a token or a username/password",
		},
		"bitbucketserver username and token": {
			raw: `{"bitbucketServer":[{"username":"u","token":"t"}]}`,
		},
		"bitbucketserver username and password": {
			raw: `{"bitbucketServer":[{"username":"u","password":"p"}]}`,
		},
		"bitbucketserver password": {
			raw: `{"bitbucketServer":[{"password":"p"}]}`,
		},
		"bitbucketserver token": {
			raw: `{"bitbucketServer":[{"token":"t"}]}`,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			problems, err := validateCustomBasicRaw([]byte(test.raw))
			if err != nil {
				t.Fatalf("got unexpected error: %v", err)
			}

			if test.wantProblem == "" {
				if len(problems) > 0 {
					t.Fatalf("unexpected problems: %v", problems)
				}
				return
			}
			for _, p := range problems {
				if strings.Contains(p, test.wantProblem) {
					return
				}
			}
			t.Fatalf("could not find problem %q in %v", test.wantProblem, problems)
		})
	}
}
