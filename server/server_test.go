package server

import "testing"

func Test_getServiceName(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{name: "1", args: args{s: "*main.UserServer"}, want: "main/UserServer"},
		{name: "2", args: args{s: "main.UserServer"}, want: "main/UserServer"},
		{name: "3", args: args{s: "user.like.UserServer"}, want: "user/like/UserServer"},
		{name: "4", args: args{s: "*user.like.UserServer"}, want: "user/like/UserServer"},
		{name: "5", args: args{s: "UserServer"}, want: "UserServer"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getServiceName(tt.args.s); got != tt.want {
				t.Errorf("getServiceName() = %v, want %v", got, tt.want)
			}
		})
	}
}
