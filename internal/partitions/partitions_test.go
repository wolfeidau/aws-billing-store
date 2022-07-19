package partitions

import "testing"

func Test_buildQuery(t *testing.T) {
	type args struct {
		table string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "should build valid query",
			args: args{table: "table_name"},
			want: "ALTER TABLE table_name ADD IF NOT EXISTS\nPARTITION (year=?, month=?)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildQuery(tt.args.table)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("buildQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}
