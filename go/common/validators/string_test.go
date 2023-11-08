package validators

import "testing"

func TestLength(t *testing.T) {
	tests := []struct {
		name string
		s    string
		min  int
		max  int
		msg  string
		want bool
	}{
		{
			name: "valid",
			s:    "hello",
			min:  2,
			max:  10,
			msg:  "invalid",
			want: true,
		},
		{
			name: "invalid",
			s:    "hello",
			min:  10,
			max:  20,
			msg:  "invalid",
			want: false,
		},
		{
			name: "empty",
			s:    "",
			min:  2,
			max:  10,
			msg:  "invalid",
			want: false,
		},
		{
			name: "valid",
			s:    "hello",
			min:  2,
			max:  10,
			msg:  "invalid",
			want: true,
		},
		{
			name: "valid",
			s:    "hello",
			min:  2,
			max:  10,
			msg:  "invalid",
			want: true,
		},
		{
			name: "valid",
			s:    "hello",
			min:  2,
			max:  10,
			msg:  "invalid",
			want: true,
		},
		{
			name: "valid",
			s:    "hello",
			min:  2,
			max:  10,
			msg:  "invalid",
			want: true,
		},
		{
			name: "valid",
			s:    "hello",
			min:  2,
			max:  10,
			msg:  "invalid",
			want: true,
		},
		{
			name: "valid",
			s:    "hello",
			min:  2,
			max:  10,
			msg:  "invalid",
			want: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, got := Length(test.s, test.min, test.max, test.msg)
			if got != test.want {
				t.Errorf("Length() = %v, want %v", got, test.want)
			}
		})
	}
}
