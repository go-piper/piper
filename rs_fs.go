// Copyright (c) 2022 Vincent Cheung (coolingfall@gmail.com).
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package piper

import (
	"embed"
	"os"
	"syscall"
	"time"

	"github.com/spf13/afero"
)

// resourceFs defines a readonly afero.Fs which is used to read config in resources
// of application. The os filesystem has high priority to read. The embedded
// filesystem will be read if os filesystem not found.
type resourceFs struct {
	osFs  afero.Fs
	memFs afero.Fs
}

// memFs will be used to store embedded resources
var memFs = afero.NewMemMapFs()

// newResourceFs creates a new instance resourceFs.
func newResourceFs(embed embed.FS) afero.Fs {
	return &resourceFs{
		osFs:  afero.NewOsFs(),
		memFs: afero.NewReadOnlyFs(memFs),
	}

}

func (*resourceFs) Create(_ string) (afero.File, error) {
	return nil, syscall.EPERM
}

func (*resourceFs) Mkdir(_ string, _ os.FileMode) error {
	return syscall.EPERM
}

func (*resourceFs) MkdirAll(_ string, _ os.FileMode) error {
	return syscall.EPERM
}

func (fs *resourceFs) Open(name string) (afero.File, error) {
	file, err := fs.osFs.Open(name)
	if err == nil {
		return file, nil
	}

	return fs.memFs.Open(name)
}

func (fs *resourceFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	file, err := fs.osFs.OpenFile(name, flag, perm)
	if err == nil {
		return file, nil
	}

	return fs.memFs.OpenFile(name, flag, perm)
}

func (*resourceFs) Remove(_ string) error {
	return syscall.EPERM
}

func (*resourceFs) RemoveAll(_ string) error {
	return syscall.EPERM
}

func (*resourceFs) Rename(_, _ string) error {
	return syscall.EPERM
}

func (fs *resourceFs) Stat(name string) (os.FileInfo, error) {
	fileInfo, err := fs.osFs.Stat(name)
	if err == nil {
		return fileInfo, nil
	}

	return fs.memFs.Stat(name)
}

func (*resourceFs) Name() string {
	return "ResourceFs"
}

func (*resourceFs) Chmod(_ string, _ os.FileMode) error {
	return syscall.EPERM
}

func (*resourceFs) Chown(_ string, _, _ int) error {
	return syscall.EPERM
}

func (*resourceFs) Chtimes(_ string, _ time.Time, _ time.Time) error {
	return syscall.EPERM
}
