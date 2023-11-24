package common

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"reflect"
	"testing"
)

func TestNewDDB(t *testing.T) {
	client := NewDDB()

	if nil == client {
		t.Error("client is nil")
	}
}

func TestAtomicExpr(t *testing.T) {
	type args struct {
		a int
	}
	tests := []struct {
		name string
		args args
		want map[string]types.AttributeValue
	}{
		{
			name: "test 1",
			args: args{a: 1},
			want: map[string]types.AttributeValue{
				":c": &types.AttributeValueMemberN{Value: "1"},
			},
		},
		{
			name: "test long",
			args: args{a: 2179834},
			want: map[string]types.AttributeValue{
				":c": &types.AttributeValueMemberN{Value: "2179834"},
			},
		},
		{
			name: "test 0",
			args: args{a: 0},
			want: nil,
		},
		{
			name: "test -1",
			args: args{a: -1},
			want: map[string]types.AttributeValue{
				":c": &types.AttributeValueMemberN{Value: "-1"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AtomicExpr(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AtomicExpr() = %v, want %v", got, tt.want)
			}
		})
	}
}
