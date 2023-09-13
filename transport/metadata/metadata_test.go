package metadata

import (
	"context"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		mds []map[string]string
	}
	tests := []struct {
		name string
		args args
		want Metadata
	}{
		{
			name: "hello",
			args: args{[]map[string]string{{"hello": "goctopus"}, {"hello2": "go-goctopus"}}},
			want: Metadata{"hello": "goctopus", "hello2": "go-goctopus"},
		},
		{
			name: "hi",
			args: args{[]map[string]string{{"hi": "goctopus"}, {"hi2": "go-goctopus"}}},
			want: Metadata{"hi": "goctopus", "hi2": "go-goctopus"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.mds...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMetadata_Get(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		m    Metadata
		args args
		want string
	}{
		{
			name: "goctopus",
			m:    Metadata{"goctopus": "value", "env": "dev"},
			args: args{key: "goctopus"},
			want: "value",
		},
		{
			name: "env",
			m:    Metadata{"goctopus": "value", "env": "dev"},
			args: args{key: "env"},
			want: "dev",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.Get(tt.args.key); got != tt.want {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMetadata_Set(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name string
		m    Metadata
		args args
		want Metadata
	}{
		{
			name: "goctopus",
			m:    Metadata{},
			args: args{key: "hello", value: "goctopus"},
			want: Metadata{"hello": "goctopus"},
		},
		{
			name: "env",
			m:    Metadata{"hello": "goctopus"},
			args: args{key: "env", value: "pro"},
			want: Metadata{"hello": "goctopus", "env": "pro"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.m.Set(tt.args.key, tt.args.value)
			if !reflect.DeepEqual(tt.m, tt.want) {
				t.Errorf("Set() = %v, want %v", tt.m, tt.want)
			}
		})
	}
}

func TestClientContext(t *testing.T) {
	type args struct {
		ctx context.Context
		md  Metadata
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "goctopus",
			args: args{context.Background(), Metadata{"hello": "goctopus", "goctopus": "https://go-goctopus.dev"}},
		},
		{
			name: "hello",
			args: args{context.Background(), Metadata{"hello": "goctopus", "hello2": "https://go-goctopus.dev"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewClientContext(tt.args.ctx, tt.args.md)
			m, ok := FromClientContext(ctx)
			if !ok {
				t.Errorf("FromClientContext() = %v, want %v", ok, true)
			}

			if !reflect.DeepEqual(m, tt.args.md) {
				t.Errorf("meta = %v, want %v", m, tt.args.md)
			}
		})
	}
}

func TestServerContext(t *testing.T) {
	type args struct {
		ctx context.Context
		md  Metadata
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "goctopus",
			args: args{context.Background(), Metadata{"hello": "goctopus", "goctopus": "https://go-goctopus.dev"}},
		},
		{
			name: "hello",
			args: args{context.Background(), Metadata{"hello": "goctopus", "hello2": "https://go-goctopus.dev"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewServerContext(tt.args.ctx, tt.args.md)
			m, ok := FromServerContext(ctx)
			if !ok {
				t.Errorf("FromServerContext() = %v, want %v", ok, true)
			}

			if !reflect.DeepEqual(m, tt.args.md) {
				t.Errorf("meta = %v, want %v", m, tt.args.md)
			}
		})
	}
}

func TestAppendToClientContext(t *testing.T) {
	type args struct {
		md Metadata
		kv []string
	}
	tests := []struct {
		name string
		args args
		want Metadata
	}{
		{
			name: "goctopus",
			args: args{Metadata{}, []string{"hello", "goctopus", "env", "dev"}},
			want: Metadata{"hello": "goctopus", "env": "dev"},
		},
		{
			name: "hello",
			args: args{Metadata{"hi": "https://go-goctopus.dev/"}, []string{"hello", "goctopus", "env", "dev"}},
			want: Metadata{"hello": "goctopus", "env": "dev", "hi": "https://go-goctopus.dev/"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewClientContext(context.Background(), tt.args.md)
			ctx = AppendToClientContext(ctx, tt.args.kv...)
			md, ok := FromClientContext(ctx)
			if !ok {
				t.Errorf("FromServerContext() = %v, want %v", ok, true)
			}
			if !reflect.DeepEqual(md, tt.want) {
				t.Errorf("metadata = %v, want %v", md, tt.want)
			}
		})
	}
}

func TestMergeToClientContext(t *testing.T) {
	type args struct {
		md       Metadata
		appendMd Metadata
	}
	tests := []struct {
		name string
		args args
		want Metadata
	}{
		{
			name: "goctopus",
			args: args{Metadata{}, Metadata{"hello": "goctopus", "env": "dev"}},
			want: Metadata{"hello": "goctopus", "env": "dev"},
		},
		{
			name: "hello",
			args: args{Metadata{"hi": "https://go-goctopus.dev/"}, Metadata{"hello": "goctopus", "env": "dev"}},
			want: Metadata{"hello": "goctopus", "env": "dev", "hi": "https://go-goctopus.dev/"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewClientContext(context.Background(), tt.args.md)
			ctx = MergeToClientContext(ctx, tt.args.appendMd)
			md, ok := FromClientContext(ctx)
			if !ok {
				t.Errorf("FromServerContext() = %v, want %v", ok, true)
			}
			if !reflect.DeepEqual(md, tt.want) {
				t.Errorf("metadata = %v, want %v", md, tt.want)
			}
		})
	}
}

func TestMetadata_Range(t *testing.T) {
	md := Metadata{"goctopus": "goctopus", "https://go-goctopus.dev/": "https://go-goctopus.dev/", "go-goctopus": "go-goctopus"}
	tmp := Metadata{}
	md.Range(func(k, v string) bool {
		if k == "https://go-goctopus.dev/" || k == "goctopus" {
			tmp[k] = v
		}
		return true
	})
	if !reflect.DeepEqual(tmp, Metadata{"https://go-goctopus.dev/": "https://go-goctopus.dev/", "goctopus": "goctopus"}) {
		t.Errorf("metadata = %v, want %v", tmp, Metadata{"goctopus": "goctopus"})
	}
}

func TestMetadata_Clone(t *testing.T) {
	tests := []struct {
		name string
		m    Metadata
		want Metadata
	}{
		{
			name: "goctopus",
			m:    Metadata{"goctopus": "goctopus", "https://go-goctopus.dev/": "https://go-goctopus.dev/", "go-goctopus": "go-goctopus"},
			want: Metadata{"goctopus": "goctopus", "https://go-goctopus.dev/": "https://go-goctopus.dev/", "go-goctopus": "go-goctopus"},
		},
		{
			name: "go",
			m:    Metadata{"language": "golang"},
			want: Metadata{"language": "golang"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.m.Clone()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Clone() = %v, want %v", got, tt.want)
			}
			got["goctopus"] = "go"
			if reflect.DeepEqual(got, tt.want) {
				t.Errorf("want got != want got %v want %v", got, tt.want)
			}
		})
	}
}
