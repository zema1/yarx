package yarx

import (
	"encoding/json"
	"fmt"
	"github.com/google/cel-go/checker/decls"
	"github.com/stretchr/testify/require"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"testing"
)

func TestCelContext_MarshalJSON(t *testing.T) {
	assert := require.New(t)
	celCtx := NewCelContext()
	/*
			case expr.Type_BYTES:
			t = decls.Bytes
		case expr.Type_BOOL:
			t = decls.Bool
		case expr.Type_INT64:
			t = decls.Int
		case expr.Type_UINT64:
			t = decls.Uint
		case expr.Type_DOUBLE:
			t = decls.Double
		case expr.Type_STRING:
			t = decls.String
	*/
	celCtx.eval = map[string]interface{}{
		"bytes":  []byte("this is bytes"),
		"bool":   true,
		"int64":  int64(123),
		"uint64": uint64(456),
		"double": float64(7.89),
		"string": "hello string",
	}
	celCtx.vafDefines = map[string]*expr.Type{
		"bytes":  decls.Bytes,
		"bool":   decls.Bool,
		"int64":  decls.Int,
		"uint64": decls.Uint,
		"double": decls.Double,
		"string": decls.String,
	}
	data, err := json.Marshal(celCtx)
	assert.Nil(err)
	fmt.Println(string(data))
	var newCtx CelContext
	err = json.Unmarshal(data, &newCtx)
	assert.Nil(err)
	fmt.Printf("%+v", newCtx.eval)
	fmt.Printf("%+v", newCtx.vafDefines)
}
