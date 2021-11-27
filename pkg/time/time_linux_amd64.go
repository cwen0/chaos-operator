// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package time

import (
	"bytes"
	"debug/elf"
	"embed"
	"encoding/binary"
	"os"
	"strings"

	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

//go:embed fakeclock/*.o
var fakeclock embed.FS

type FakeImage struct {
	content []byte
	offset  map[string]int
}

var fakeImages = map[string]FakeImage{}

func init() {
	// in this function, we will load fake image from `fakeclock/*.o`
	entries, err := fakeclock.ReadDir("fakeclock")
	if err != nil {
		log.Error(err, "readdir from embedded fs")
		os.Exit(1)
	}

	for _, entry := range entries {
		if entry.Name() == ".embed.o" {
			// skip the .embed.o file, as it's used to remove the error of `go fmt`
			continue
		}
		path := "fakeclock/" + entry.Name()
		object, err := fakeclock.ReadFile(path)
		if err != nil {
			log.Error(err, "read file from embedded fs", "path", path)
			os.Exit(1)
		}

		elfFile, err := elf.NewFile(bytes.NewReader(object))
		if err != nil {
			log.Error(err, "parse elf", "path", path)
			os.Exit(1)
		}

		syms, err := elfFile.Symbols()
		if err != nil {
			log.Error(err, "get symbols")
			os.Exit(1)
		}

		fakeImage := FakeImage{
			offset: make(map[string]int),
		}
		for _, r := range elfFile.Sections {
			if r.Type == elf.SHT_PROGBITS && r.Name == ".text" {
				fakeImage.content, err = r.Data()
				if err != nil {
					log.Error(err, "read text section")
					os.Exit(1)
				}

				break
			}
		}

		for _, r := range elfFile.Sections {
			if r.Type == elf.SHT_RELA && r.Name == ".rela.text" {
				rela_section, err := r.Data()
				if err != nil {
					log.Error(err, "read rela section")
					os.Exit(1)
				}
				rela_section_reader := bytes.NewReader(rela_section)

				var rela elf.Rela64
				for rela_section_reader.Len() > 0 {
					binary.Read(rela_section_reader, elfFile.ByteOrder, &rela)

					symNo := rela.Info >> 32
					if symNo == 0 || symNo > uint64(len(syms)) {
						continue
					}

					// The relocation of a X86 image is like:
					// Relocation section '.rela.text' at offset 0x288 contains 3 entries:
					// Offset          Info           Type           Sym. Value    Sym. Name + Addend
					// 000000000016  000900000002 R_X86_64_PC32     0000000000000000 CLOCK_IDS_MASK - 4
					// 00000000001f  000a00000002 R_X86_64_PC32     0000000000000008 TV_NSEC_DELTA - 4
					// 00000000002a  000b00000002 R_X86_64_PC32     0000000000000010 TV_SEC_DELTA - 4
					//
					// For example, we need to write the offset of `CLOCK_IDS_MASK` - 4 in 0x16 of the section
					// If we want to put the `CLOCK_IDS_MASK` at the end of the section, it will be
					// len(fakeImage.content) - 4 - 0x16

					sym := &syms[symNo-1]
					fakeImage.offset[sym.Name] = len(fakeImage.content)
					targetOffset := uint32(len(fakeImage.content)) - uint32(rela.Off) + uint32(rela.Addend)
					elfFile.ByteOrder.PutUint32(fakeImage.content[rela.Off:rela.Off+4], targetOffset)

					// TODO: support other length besides uint64 (which is 8 bytes)
					fakeImage.content = append(fakeImage.content, make([]byte, 8)...)
				}

				break
			}
		}

		fakeImages[strings.TrimSuffix(entry.Name(), ".o")] = fakeImage
	}
}

// ModifyTime modifies time of target process
func ModifyTime(pid int, deltaSec int64, deltaNsec int64, clockIdsMask uint64) error {
	// Mock point to return error in unit test
	if err := mock.On("ModifyTimeError"); err != nil {
		if e, ok := err.(error); ok {
			return e
		}
		if ignore, ok := err.(bool); ok && ignore {
			return nil
		}
	}
	return NewTimeSkew(deltaSec, deltaNsec, clockIdsMask).Inject(pid)
}
