package partitions

import "testing"

func Test_buildQuery(t *testing.T) {
	type args struct {
		table string
		year  string
		month string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "should build valid query",
			args: args{table: "table_name", year: "2023", month: "1"},
			want: "ALTER TABLE table_name ADD IF NOT EXISTS\n    PARTITION (year = '2023', month = '1')",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildQuery(tt.args.table, tt.args.year, tt.args.month)
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
