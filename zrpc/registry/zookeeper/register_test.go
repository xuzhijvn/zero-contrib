package zookeeper

import (
	"testing"

	"github.com/go-zookeeper/zk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// 模拟zk.Conn接口
type MockZkConn struct {
	mock.Mock
}

func (m *MockZkConn) Create(path string, data []byte, flags int32, acl []zk.ACL) (string, error) {
	args := m.Called(path, data, flags, acl)
	return args.String(0), args.Error(1)
}

func (m *MockZkConn) Delete(path string, version int32) error {
	args := m.Called(path, version)
	return args.Error(0)
}

func (m *MockZkConn) Exists(path string) (bool, *zk.Stat, error) {
	args := m.Called(path)
	return args.Bool(0), args.Get(1).(*zk.Stat), args.Error(2)
}

func (m *MockZkConn) Close() {
	m.Called()
}

// 测试createPath函数
func TestCreatePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		exists  bool
		wantErr bool
	}{
		{
			name:    "路径已存在",
			path:    "/services",
			exists:  true,
			wantErr: false,
		},
		{
			name:    "路径不存在，创建成功",
			path:    "/services",
			exists:  false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := new(MockZkConn)
			mockConn.On("Exists", tt.path).Return(tt.exists, &zk.Stat{}, nil)

			if !tt.exists {
				mockConn.On("Create", tt.path, []byte{}, int32(0), zk.WorldACL(zk.PermAll)).Return(tt.path, nil)
			}

			err := createPath(mockConn, tt.path)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockConn.AssertExpectations(t)
		})
	}
}

// 测试figureOutListenOn函数
func TestFigureOutListenOn(t *testing.T) {
	tests := []struct {
		name     string
		listenOn string
		want     string
	}{
		{
			name:     "指定IP",
			listenOn: "192.168.1.1:8080",
			want:     "192.168.1.1:8080",
		},
		{
			name:     "空IP",
			listenOn: ":8080",
			want:     ":8080", // 在测试环境中无法获取内部IP，所以预期结果不变
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := figureOutListenOn(tt.listenOn)
			assert.Equal(t, tt.want, got)
		})
	}
}
