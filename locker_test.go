package locker

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/afero"
)

func lockedFs(t *testing.T, path string) afero.Fs {
	fs := afero.NewMemMapFs()
	f, err := fs.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	_, err = f.WriteString("1")
	if err != nil {
		t.Fatal(err)
	}

	return fs
}

func unlockedFs(_ *testing.T) afero.Fs {
	fs := afero.NewMemMapFs()
	return fs
}

func TestLocker_Lock(t *testing.T) {
	type fields struct {
		fs      afero.Fs
		path    string
		content interface{}
	}
	type args struct {
		p string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:   "round-trip",
			fields: fields{fs: unlockedFs(t), path: "/lock", content: 1},
		},
		{
			name:    "locked",
			fields:  fields{fs: lockedFs(t, "/lock"), path: "/lock", content: 1},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := _new(tt.fields.fs, tt.fields.path, tt.fields.content)
			err := l.Lock()
			if (err != nil) != tt.wantErr {
				t.Errorf("Lock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestLocker_Unlock(t *testing.T) {
	type fields struct {
		fs      afero.Fs
		path    string
		content interface{}
	}
	tests := []struct {
		name     string
		fields   fields
		wantErr  bool
		wantFile bool
	}{
		{
			name: "locked",
			fields: fields{
				fs:      lockedFs(t, "/lock"),
				path:    "/lock",
				content: 1,
			},
			wantFile: false,
		},
		{
			name: "not locked",
			fields: fields{
				fs:      unlockedFs(t),
				path:    "/lock",
				content: 1,
			},
			wantErr:  false,
			wantFile: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := Locker{
				fs:      tt.fields.fs,
				path:    tt.fields.path,
				content: tt.fields.content,
			}
			if err := l.Unlock(); (err != nil) != tt.wantErr {
				t.Errorf("Unlock() error = %v, wantErr %v", err, tt.wantErr)
			}
			if haveFile := l.exists(); haveFile != tt.wantFile {
				t.Errorf("Unlock() wantFile = %t, haveFile: %t", tt.wantFile, haveFile)
			}
		})
	}
}

func TestLocker_ShutdownContext(t *testing.T) {
	type fields struct {
		fs      afero.Fs
		path    string
		content interface{}
	}
	tests := []struct {
		name     string
		fields   fields
		wantFile bool
	}{
		{
			name: "locked",
			fields: fields{
				fs:      lockedFs(t, "/lock"),
				path:    "/lock",
				content: 1,
			},
			wantFile: false,
		},
		{
			name: "not locked",
			fields: fields{
				fs:      unlockedFs(t),
				path:    "/lock",
				content: 1,
			},
			wantFile: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := Locker{
				fs:      tt.fields.fs,
				path:    tt.fields.path,
				content: tt.fields.content,
			}
			ctx, cancel := context.WithCancel(context.Background())
			l.ShutdownContext(ctx)
			cancel()
			time.Sleep(50*time.Millisecond)
			if haveFile := l.exists(); haveFile != tt.wantFile {
				t.Errorf("Unlock() wantFile = %t, haveFile: %t", tt.wantFile, haveFile)
			}
		})
	}
}
func TestLocker_Read(t *testing.T) {
	type fields struct {
		fs      afero.Fs
		path    string
		content interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		want    int
		wantErr bool
	}{
		{
			name: "Read",
			fields: fields{
				fs:      lockedFs(t, "/lock"),
				path:    "/lock",
				content: 1,
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "Read fail",
			fields: fields{
				fs:      unlockedFs(t),
				path:    "/lock",
				content: 1,
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := Locker{
				fs:      tt.fields.fs,
				path:    tt.fields.path,
				content: tt.fields.content,
			}
			var got int
			err := l.Read(&got)
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Read() got = %v, want %v", got, tt.want)
			}
		})
	}
}
