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

package physicalmachine

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SshTunnel struct {
	config *ssh.ClientConfig
	host   string
	port   int
	client *ssh.Client
}

func NewSshTunnel(ip string, port int, user, privateKeyFile string) *SshTunnel {
	config := ssh.ClientConfig{
		Timeout: 5 * time.Minute,
		User:    user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeysCallback(func() ([]ssh.Signer, error) {
				key, err := ioutil.ReadFile(privateKeyFile)
				if err != nil {
					return nil, fmt.Errorf("ssh key file read failed: %v", err)
				}

				signer, err := ssh.ParsePrivateKey(key)
				if err != nil {
					return nil, fmt.Errorf("ssh key signer failed: %v", err)
				}
				return []ssh.Signer{signer}, nil
			}),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	return &SshTunnel{
		config: &config,
		host:   ip,
		port:   port,
	}
}

func (s *SshTunnel) Open() error {
	// first, try ssh with public keys
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", s.host, s.port), s.config)
	if err != nil {
		// if failed, try ssh with password
		fmt.Printf("please input the password: ")
		password := ""
		fmt.Scanf("%s\n", &password)

		s.config.Auth = []ssh.AuthMethod{ssh.Password(password)}
		conn, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", s.host, s.port), s.config)
		if err != nil {
			return err
		}
	}
	s.client = conn
	return nil
}

func (s *SshTunnel) Close() error {
	if s.client == nil {
		return nil
	}
	return s.client.Close()
}

func (s *SshTunnel) SFTP(filename string, data []byte) error {
	if s.client == nil {
		return errors.New("tunnel is not opened")
	}

	// open an SFTP session over an existing ssh connection.
	client, err := sftp.NewClient(s.client)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.MkdirAll(filepath.Dir(filename)); err != nil {
		return err
	}

	f, err := client.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return err
	}
	return nil
}
