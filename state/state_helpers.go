package state

import (
	"fmt"
	"strconv"
)

func GetValueAsString(r Reader, arg Argument) (string, error) {
	switch arg.Type {
	case ArgumentTypeString:
		return r.GetString(arg)
	case ArgumentTypeInt64:
		val, err := r.GetInt64(arg)
		if err != nil {
			return "", err
		}

		return strconv.FormatInt(val, 10), nil
	case ArgumentTypeFloat64:
		val, err := r.GetFloat64(arg)
		if err != nil {
			return "", err
		}

		return strconv.FormatFloat(val, 'f', 8, 64), nil
	case ArgumentTypeBool:
		val, err := r.GetBool(arg)
		if err != nil {
			return "", err
		}

		return strconv.FormatBool(val), nil
	case ArgumentTypeFile:
		file, err := r.GetFile(arg)
		if err != nil {
			return "", err
		}

		return file.Name(), nil

	case ArgumentTypeUnpackagedFS, ArgumentTypeFS:
		return r.GetDirectoryString(arg)

	default:
	}

	return "", fmt.Errorf("unsupported or unrecognized argument type: %s", arg.Type)
}
