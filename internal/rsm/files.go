// Copyright 2017-2019 Lei Ni (nilei81@gmail.com)
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

package rsm

import (
	"os"
	"path/filepath"

	pb "github.com/lni/dragonboat/raftpb"
)

type Files struct {
	files []*pb.SnapshotFile
	idmap map[uint64]struct{}
}

func NewFileCollection() *Files {
	return &Files{
		files: make([]*pb.SnapshotFile, 0),
		idmap: make(map[uint64]struct{}),
	}
}

func (fc *Files) AddFile(fileID uint64,
	path string, metadata []byte) {
	if _, ok := fc.idmap[fileID]; ok {
		plog.Panicf("trying to add file %d again", fileID)
	}
	f := &pb.SnapshotFile{
		Filepath: path,
		FileId:   fileID,
		Metadata: metadata,
	}
	fc.files = append(fc.files, f)
	fc.idmap[fileID] = struct{}{}
}

func (fc *Files) Size() uint64 {
	return uint64(len(fc.files))
}

func (fc *Files) GetFileAt(idx uint64) *pb.SnapshotFile {
	return fc.files[idx]
}

func (fc *Files) PrepareFiles(tmpdir string,
	finaldir string) ([]*pb.SnapshotFile, error) {
	for _, file := range fc.files {
		fn := file.Filename()
		fp := filepath.Join(tmpdir, fn)
		if err := os.Link(file.Filepath, fp); err != nil {
			return nil, err
		}
		fi, err := os.Stat(fp)
		if err != nil {
			return nil, err
		}
		if fi.IsDir() {
			plog.Panicf("%s is a dir", fp)
		}
		if fi.Size() == 0 {
			plog.Panicf("empty file found, id %d",
				file.FileId)
		}
		file.Filepath = filepath.Join(finaldir, fn)
		file.FileSize = uint64(fi.Size())
	}
	return fc.files, nil
}