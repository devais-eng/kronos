package util

import (
	"devais.it/kronos/internal/pkg/constants"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"github.com/sirupsen/logrus"
	"strings"
)

type VersionAlgorithm int

const (
	VersionAlgorithmSha1 VersionAlgorithm = iota
	VersionAlgorithmMd5
	VersionAlgorithmUuid
)

func (v VersionAlgorithm) MarshalText() ([]byte, error) {
	switch v {
	case VersionAlgorithmSha1:
		return []byte("sha1"), nil
	case VersionAlgorithmMd5:
		return []byte("md5"), nil
	case VersionAlgorithmUuid:
		return []byte("uuid"), nil
	default:
		return nil, eris.New("Invalid version algorithm")
	}
}

func (v VersionAlgorithm) String() string {
	text, err := v.MarshalText()
	if err != nil {
		return "Unknown"
	}
	return string(text)
}

func VersionAlgorithmFromString(str string) (VersionAlgorithm, error) {
	switch strings.ToLower(str) {
	case "sha1":
		return VersionAlgorithmSha1, nil
	case "md5":
		return VersionAlgorithmMd5, nil
	case "uuid":
		return VersionAlgorithmUuid, nil
	default:
		return -1, eris.Errorf("Unknown version algorithm: '%s'", str)
	}
}

func BytesChecksum(bytes []byte, algo VersionAlgorithm) (string, error) {
	var space uuid.UUID
	var checksum uuid.UUID

	if algo == VersionAlgorithmSha1 {
		checksum = uuid.NewSHA1(space, bytes)
	} else if algo == VersionAlgorithmMd5 {
		checksum = uuid.NewMD5(space, bytes)
	} else if algo == VersionAlgorithmUuid {
		checksum = uuid.New()
	} else {
		logrus.Panicf("Unknown version algorithm: %d", algo)
	}

	return checksum.String(), nil
}

func GenerateVersionChecksum(entity interface{}, algo VersionAlgorithm) (string, error) {
	jsonMap, err := StructToJSONMap(entity)
	if err != nil {
		return "", err
	}

	// Remove meta fields from the map
	for _, field := range constants.MetaFields {
		delete(jsonMap, field)
	}

	jsonBytes, err := json.Marshal(jsonMap)
	if err != nil {
		return "", err
	}

	return BytesChecksum(jsonBytes, algo)
}
