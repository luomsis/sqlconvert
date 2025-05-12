package ora2pg_test

import (
	"github/luomsis/sqlconvert/ora2pg"
	"testing"
)

func TestOra2PgVisitor(t *testing.T) {
	tests := []struct {
		name string
		args struct {
			sql string
		}
		want string
	}{
		{
			name: "Test1",
			args: struct{ sql string }{
				sql: "SELECT SYSDATE FROM dual;",
			},
			want: "SELECT CURRENT_TIMESTAMP(0);",
		},
		// {
		// 	name: "Test2",
		// 	args: struct{ sql string }{
		// 		sql: "SELECT * FROM dual WHERE 1=1",
		// 	},
		// 	want: "SELECT * FROM pg_catalog.dual WHERE 1=1",
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ora2pg.Convert1(tt.args.sql)
			if got != tt.want {
				t.Errorf("Convert() = %v, want %v", got, tt.want)
			}
		})
	}
}
