package state

import (
	"context"
	"fmt"
	"strconv"
)

func GetValueAsString(ctx context.Context, r Reader, arg Argument) (string, error) {
	switch arg.Type {
	case ArgumentTypeString:
		return r.GetString(ctx, arg)
	case ArgumentTypeInt64:
		val, err := r.GetInt64(ctx, arg)
		if err != nil {
			return "", err
		}

		return strconv.FormatInt(val, 10), nil
	case ArgumentTypeFloat64:
		val, err := r.GetFloat64(ctx, arg)
		if err != nil {
			return "", err
		}

		return strconv.FormatFloat(val, 'f', 8, 64), nil
	case ArgumentTypeBool:
		val, err := r.GetBool(ctx, arg)
		if err != nil {
			return "", err
		}

		return strconv.FormatBool(val), nil
	case ArgumentTypeFile:
		file, err := r.GetFile(ctx, arg)
		if err != nil {
			return "", err
		}

		return file.Name(), nil

	case ArgumentTypeUnpackagedFS, ArgumentTypeFS:
		return r.GetDirectoryString(ctx, arg)

	default:
	}

	return "", fmt.Errorf("unsupported or unrecognized argument type: %s", arg.Type)
}

func ArgListContains(args Arguments, arg Argument) bool {
	for _, v := range args {
		if v == arg {
			return true
		}
	}

	return false
}
