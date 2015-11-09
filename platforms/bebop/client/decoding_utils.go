package client

import (
  "encoding/binary"
  "bytes"
  "errors"
)

// Common task: Decode a 4-byte enum, then use it to index some strings representing enum value
var enumOutOfRangeError error = errors.New("Enum value fell outside expected range.")
var enumBadSizeError error = errors.New("Wrong size binary given for a Bebop enum. Expected 4.")

func decodeEnum(raw []byte, vals []string) (string, error) {
	var evalue int
	if len(raw) != 4 {
		return "", enumBadSizeError
	}
	binary.Read(bytes.NewReader(raw), binary.LittleEndian, &evalue)
	if evalue < 0 || evalue > len(vals)-1 {
		return "", enumOutOfRangeError
	}
	return vals[evalue], nil
}

// Given a frame presumed to contain a null-ended string, return the string
// and the reamining bytes after the null
var notStringError = errors.New("Could not locate a NUL byte in presumed NUL-terminated string")

func parseNullTermedString(dataframe []byte) (string, []byte, error) {
	nul := bytes.IndexByte(dataframe, byte(0))
	if nul == -1 {
		return "", nil, notStringError
	}
	return string(dataframe[:nul]), dataframe[nul+1:], nil
}
