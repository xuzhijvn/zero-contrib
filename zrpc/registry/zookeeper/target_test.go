package zookeeper

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseURL(t *testing.T) {
	tests := []struct {
		name    string
		rawURL  string
		want    *target
		wantErr bool
	}{
		{
			name:   "基本URL",
			rawURL: "zookeeper://127.0.0.1:2181/service-name",
			want: &target{
				Hosts:       []string{"127.0.0.1:2181"},
				BasePath:    "/services",
				ServiceName: "service-name",
				Timeout:     15 * time.Second,
			},
			wantErr: false,
		},
		{
			name:   "带查询参数的URL",
			rawURL: "zookeeper://127.0.0.1:2181,127.0.0.2:2181/service-name?timeout=30&basePath=/my-services",
			want: &target{
				Hosts:       []string{"127.0.0.1:2181", "127.0.0.2:2181"},
				BasePath:    "/my-services",
				ServiceName: "service-name",
				Timeout:     30 * time.Second,
			},
			wantErr: false,
		},
		{
			name:    "无主机URL",
			rawURL:  "zookeeper:///service-name",
			wantErr: true,
		},
		{
			name:    "无服务名URL",
			rawURL:  "zookeeper://127.0.0.1:2181/",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.rawURL)
			assert.NoError(t, err)

			got, err := parseURL(*u)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want.BasePath, got.BasePath)
			assert.Equal(t, tt.want.ServiceName, got.ServiceName)
			assert.Equal(t, tt.want.Timeout, got.Timeout)
			assert.Equal(t, tt.want.Hosts, got.Hosts)
		})
	}
}

func TestParseHosts(t *testing.T) {
	tests := []struct {
		name  string
		hosts string
		want  []string
	}{
		{
			name:  "单主机无端口",
			hosts: "localhost",
			want:  []string{"localhost:2181"},
		},
		{
			name:  "单主机带端口",
			hosts: "localhost:2182",
			want:  []string{"localhost:2182"},
		},
		{
			name:  "多主机混合",
			hosts: "localhost:2181,127.0.0.1,127.0.0.2:2183",
			want:  []string{"localhost:2181", "127.0.0.1:2181", "127.0.0.2:2183"},
		},
		{
			name:  "带空格",
			hosts: "localhost:2181, 127.0.0.1, 127.0.0.2:2183",
			want:  []string{"localhost:2181", "127.0.0.1:2181", "127.0.0.2:2183"},
		},
		{
			name:  "空主机",
			hosts: "",
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseHosts(tt.hosts)
			assert.Equal(t, tt.want, got)
		})
	}
}
