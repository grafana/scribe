package state

import "context"

type JSONState map[string]StateValueJSON

type StateValueJSON struct {
	Argument Argument `json:"argument"`
	Value    any      `json:"value"`
}

func SetValueFromJSON(ctx context.Context, w Writer, value StateValueJSON) error {
	switch value.Argument.Type {
	case ArgumentTypeString:
		return w.SetString(ctx, value.Argument, value.Value.(string))
	case ArgumentTypeInt64:
		return w.SetInt64(ctx, value.Argument, int64(value.Value.(float64)))
	case ArgumentTypeFloat64:
		return w.SetFloat64(ctx, value.Argument, value.Value.(float64))
	case ArgumentTypeBool:
		return w.SetBool(ctx, value.Argument, value.Value.(bool))
	//case ArgumentTypeSecret:
	//return w.SetSecret(value.Argument, value.Value.(bool))
	case ArgumentTypeFile:
		return w.SetFile(ctx, value.Argument, value.Value.(string))
	case ArgumentTypeFS:
		return w.SetDirectory(ctx, value.Argument, value.Value.(string))
	case ArgumentTypeUnpackagedFS:
		return w.SetDirectory(ctx, value.Argument, value.Value.(string))
	default:
	}

	return nil
}
