package cloud

import (
	"testing"

	. "main/aikido_types"
	"main/ipc/protos"

	"github.com/stretchr/testify/assert"
)

func TestGetAttackDetailsUser(t *testing.T) {
	server := &ServerData{Users: map[string]User{"1": {ID: "1", Name: "Alice"}}}

	for _, tc := range []struct {
		name     string
		userId   string
		expected *RequestUser
	}{
		{"no user", "", nil},
		{"user in the users map", "1", &RequestUser{ID: "1", Name: "Alice"}},
		{"user not in the users map yet", "2", &RequestUser{ID: "2"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			details := GetAttackDetails(server, &protos.Attack{UserId: tc.userId})

			assert.Equal(t, tc.expected, details.User)
		})
	}
}
