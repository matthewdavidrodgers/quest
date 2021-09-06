package config

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type ConfigFile struct {
	fullPath string
	fd       *os.File
}

func OpenConfigFile(filename string) (ConfigFile, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return ConfigFile{}, err
	}
	fullPath := filepath.Join(home, filename)
	fd, err := os.OpenFile(filepath.Join(home, filename), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return ConfigFile{}, err
	}

	return ConfigFile{fullPath, fd}, nil
}

func (c ConfigFile) Close() {
	c.fd.Close()
}

func (c ConfigFile) readData() ([]byte, error) {
	c.fd.Seek(0, io.SeekStart)
	data, err := io.ReadAll(c.fd)

	if err != nil {
		return nil, err
	}
	return data, nil

}

func (c ConfigFile) GetValue(key string) (string, error) {
	data, err := c.readData()
	if err != nil {
		return "", err
	}

	for _, line := range strings.Split(string(data), "\n") {
		parsed := strings.SplitN(line, "=", 2)
		if parsed[0] == key {
			return parsed[1], nil
		}
	}

	return "", nil
}

func (c ConfigFile) AppendValue(key string, value string) (string, error) {
	data, err := c.readData()
	if err != nil {
		return "", err
	}

	keyBytes := make([]byte, len(key)+2)
	copy(keyBytes[1:], []byte(key))
	keyBytes[0] = byte('\n')
	keyBytes[len(key)+1] = byte('=')

	valueStartIndex := -1
	valueEndIndex := -1
	if bytes.Equal(data[:len(keyBytes)-1], keyBytes[1:]) {
		// special case for start of file, no newline at the front
		valueStartIndex = len(keyBytes) - 1
		for valueEndIndex = valueStartIndex; valueEndIndex < len(data); valueEndIndex += 1 {
			if data[valueEndIndex] == byte('\n') {
				break
			}
		}
	} else {
		for i := 0; i < len(data)-len(keyBytes); i += 1 {
			if bytes.Equal(data[i:i+len(keyBytes)], keyBytes) {
				valueStartIndex = i + len(keyBytes)
				break
			}
		}
		if valueStartIndex != -1 {
			for valueEndIndex = valueStartIndex; valueEndIndex < len(data); valueEndIndex += 1 {
				if data[valueEndIndex] == byte('\n') {
					break
				}
			}
		}
	}

	var newValue []byte
	var newData []byte

	if valueStartIndex != -1 {
		// appending existing
		newBytesLength := len(value) + 2 // account for new "; " before new value
		newBytes := make([]byte, newBytesLength)

		newBytes[0] = byte(';')
		newBytes[1] = byte(' ')
		copy(newBytes[2:], []byte(value))

		newData = make([]byte, len(data)+newBytesLength)
		copy(newData, data[:valueEndIndex])
		copy(newData[valueEndIndex:valueEndIndex+newBytesLength], newBytes)
		copy(newData[valueEndIndex+newBytesLength:], data[valueEndIndex:])

		newValue = newData[valueStartIndex : valueEndIndex+newBytesLength]
	} else {
		// appending new entry
		var newBytes []byte
		if len(data) == 0 {
			// appending to empty file
			newBytesLength := len(value) + len(key) + 1 // account for and =
			newBytes = make([]byte, newBytesLength)

			copy(newBytes, []byte(key))
			newBytes[len(key)] = byte('=')
			copy(newBytes[len(key)+1:], []byte(value))
		} else {
			// appending to non-empty file; begin with a newline
			newBytesLength := len(value) + len(key) + 2 // account for \n and =
			newBytes = make([]byte, newBytesLength)

			newBytes[0] = byte('\n')
			copy(newBytes[1:len(key)+1], []byte(key))
			newBytes[len(key)+1] = byte('=')
			copy(newBytes[len(key)+2:], []byte(value))
		}

		newData = append(data, newBytes...)
		newValue = []byte(value)
	}

	_, err = c.fd.WriteAt(newData, 0)
	if err != nil {
		return "", err
	}
	return string(newValue), nil
}

func (c ConfigFile) RemoveValue(key string) (bool, error) {
	data, err := c.readData()
	if err != nil {
		return false, err
	}

	keyBytes := make([]byte, len(key)+2)
	keyBytes[0] = byte('\n')
	copy(keyBytes[1:], []byte(key))
	keyBytes[len(key)+1] = byte('=')

	valueStartIndex := -1
	valueEndIndex := -1
	if bytes.Equal(data[:len(keyBytes)-1], keyBytes[1:]) {
		// special case for start of file, no newline at the front
		valueStartIndex = 0
		for valueEndIndex = valueStartIndex + len(keyBytes) - 1; valueEndIndex < len(data); valueEndIndex += 1 {
			if data[valueEndIndex] == byte('\n') {
				break
			}
		}
	} else {
		for i := 0; i < len(data)-len(keyBytes); i += 1 {
			if bytes.Equal(data[i:i+len(keyBytes)], keyBytes) {
				valueStartIndex = i
				break
			}
		}
		if valueStartIndex != -1 {
			for valueEndIndex = valueStartIndex + len(keyBytes); valueEndIndex < len(data); valueEndIndex += 1 {
				if data[valueEndIndex] == byte('\n') {
					break
				}
			}
		}
	}

	if valueStartIndex == -1 {
		return false, nil // value is not present, nothing to remove
	}

	entryLength := valueEndIndex - valueStartIndex
	newData := make([]byte, len(data)-entryLength)
	copy(newData[:valueStartIndex], data[:valueStartIndex])
	copy(newData[valueStartIndex:], data[valueEndIndex:])

	_, err = c.fd.WriteAt(newData, 0)
	if err != nil {
		return false, err
	}
	if err = c.fd.Truncate(int64(len(newData))); err != nil {
		return false, err
	}

	return true, nil
}

func (c ConfigFile) GetCookie() (string, error) {
	return c.GetValue("COOKIE")
}

func (c ConfigFile) AddCookieToStore(cookie string) error {
	_, err := c.AppendValue("COOKIE", cookie)
	return err
}

func (c ConfigFile) RemoveCookie() error {
	_, err := c.RemoveValue("COOKIE")
	return err
}
