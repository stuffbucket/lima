package spiceclient

// SPDX-FileCopyrightText: Copyright The Lima Authors
// SPDX-License-Identifier: Apache-2.0

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestGetConnectionInfo(t *testing.T) {
	tests := []struct {
		name       string
		displayStr string
		wantHost   string
		wantPort   string
		wantUnix   string
		wantErr    bool
	}{
		{
			name:       "Basic SPICE with default port",
			displayStr: "spice",
			wantHost:   "127.0.0.1",
			wantPort:   "5900",
		},
		{
			name:       "SPICE with custom port",
			displayStr: "spice,port=5930",
			wantHost:   "127.0.0.1",
			wantPort:   "5930",
		},
		{
			name:       "SPICE with custom host and port",
			displayStr: "spice,addr=0.0.0.0,port=5931",
			wantHost:   "0.0.0.0",
			wantPort:   "5931",
		},
		{
			name:       "SPICE with password",
			displayStr: "spice,port=5900,password=secret123",
			wantHost:   "127.0.0.1",
			wantPort:   "5900",
		},
		{
			name:       "SPICE Unix socket",
			displayStr: "spice+unix:///tmp/spice.sock",
			wantUnix:   "/tmp/spice.sock",
		},
		{
			name:       "Invalid display string",
			displayStr: "vnc",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, err := GetConnectionInfo(tt.displayStr)
			if tt.wantErr {
				assert.Assert(t, err != nil)
				return
			}
			assert.NilError(t, err)

			if tt.wantUnix != "" {
				assert.Equal(t, tt.wantUnix, conn.UnixPath)
			} else {
				assert.Equal(t, tt.wantHost, conn.Host)
				assert.Equal(t, tt.wantPort, conn.Port)
			}
		})
	}
}

func TestBuildSpiceURI(t *testing.T) {
	tests := []struct {
		name    string
		conn    *Connection
		want    string
		wantErr bool
	}{
		{
			name: "TCP connection without password",
			conn: &Connection{
				Host: "127.0.0.1",
				Port: "5900",
			},
			want: "spice://127.0.0.1:5900",
		},
		{
			name: "TCP connection with password",
			conn: &Connection{
				Host:     "192.168.1.100",
				Port:     "5930",
				Password: "secret",
			},
			want: "spice://192.168.1.100:5930?password=secret",
		},
		{
			name: "Unix socket connection",
			conn: &Connection{
				UnixPath: "/var/run/spice.sock",
			},
			want: "spice+unix:///var/run/spice.sock",
		},
		{
			name: "Missing host",
			conn: &Connection{
				Port: "5900",
			},
			wantErr: true,
		},
		{
			name: "Missing port",
			conn: &Connection{
				Host: "127.0.0.1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildSpiceURI(tt.conn)
			if tt.wantErr {
				assert.Assert(t, err != nil)
				return
			}
			assert.NilError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
