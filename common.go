package yarx

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/cel-go/checker/decls"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

var (
	ErrReverseNotSupported = errors.New("reverse type is not supported yet")
	ErrRequestNotSupported = errors.New("request variable is not supported yet")
)

type CelContext struct {
	eval       map[string]interface{}
	vafDefines map[string]*expr.Type
}

func (c *CelContext) UnmarshalJSON(bytes []byte) error {
	m := make(map[string]map[string]interface{})
	err := json.Unmarshal(bytes, &m)
	if err != nil {
		return err
	}
	c.eval = m["eval"]
	c.vafDefines = make(map[string]*expr.Type)
	for key, ptInter := range m["defines"] {
		ptFloat, ok := ptInter.(float64)
		if !ok {
			return fmt.Errorf("type convertion to float64 failed, %v", ptInter)
		}
		pt := expr.Type_PrimitiveType(ptFloat)
		var t *expr.Type
		switch pt {
		case expr.Type_BYTES:
			t = decls.Bytes
			if c.eval[key] != nil {
				if val, ok :=  c.eval[key].(string); ok {
					base64.StdEncoding.De
				}
			}
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
		default:
			return fmt.Errorf("error type of primitiveType, %v", pt)
		}
		c.vafDefines[key] = t
	}
	return nil
}

func (c *CelContext) MarshalJSON() ([]byte, error) {
	m := map[string]map[string]interface{}{
		"eval":    c.eval,
		"defines": {},
	}
	for key, t := range c.vafDefines {
		if t.GetPrimitive() == expr.Type_PRIMITIVE_TYPE_UNSPECIFIED {
			return nil, fmt.Errorf("unsupported types of %s", t)
		}
		m["defines"][key] = t.GetPrimitive()
	}
	return json.Marshal(m)
}

func NewCelContext() *CelContext {
	return &CelContext{
		eval:       make(map[string]interface{}),
		vafDefines: make(map[string]*expr.Type),
	}
}

var variableRegex = regexp.MustCompile(`{{([a-zA-Z0-9_]+)}}`)

func variableToRegexp(template string, varContext map[string]interface{}, withRapper bool, fixNewline bool) (*regexp.Regexp, string, error) {
	var replacedStr string
	if !strings.Contains(template, "{{") {
		replacedStr = template
		template = regexp.QuoteMeta(template)
		if withRapper {
			template = "^" + template + "$"
		}
	} else {
		replaceMap := make(map[string]string)
		for i, arr := range variableRegex.FindAllStringSubmatch(template, -1) {
			if val, ok := varContext[arr[1]]; ok {
				switch v := val.(type) {
				case []byte, byte:
					template = strings.ReplaceAll(template, arr[0], fmt.Sprintf("%s", v))
				default:
					template = strings.ReplaceAll(template, arr[0], fmt.Sprintf("%v", v))
				}
			} else {
				namedGroup := fmt.Sprintf(`(?P<%s>\w+)`, arr[1])
				// 这里不能直接替换，因为数据中的 ? . () 之类的不应该被视为正则, 这里用占位符先弄一下，后面转义后再替换
				// 这个占位符既不能是正则中的字符，也不能是 url 需要转移的字符
				placeholder := fmt.Sprintf(`---variable-%d---`, i)
				replaceMap[placeholder] = namedGroup
				template = strings.Replace(template, arr[0], placeholder, 1)
			}
		}
		replacedStr = template
		template = regexp.QuoteMeta(template)
		for k, v := range replaceMap {
			template = strings.Replace(template, k, v, 1)
		}
	}

	if withRapper {
		template = "^" + template + "$"
	}
	if fixNewline {
		template = strings.Replace(template, "\n", ".{1,2}?", -1)
	}
	template = "(?s)" + template
	re, err := regexp.Compile(template)
	return re, replacedStr, err
}

func SortedURI(u *url.URL) string {
	result := u.Opaque
	if result == "" {
		result = u.EscapedPath()
		if result == "" {
			result = "/"
		}
	} else {
		if strings.HasPrefix(result, "//") {
			result = u.Scheme + ":" + result
		}
	}
	p, err := url.PathUnescape(result)
	if err == nil {
		result = p
	}
	query := SortedQuery(u.RawQuery)
	if query != "" {
		result = result + "?" + query
	}
	return result
}

func SortedQueryKey(query string) string {
	var queryKeys []string
	for query != "" {
		key := query
		if i := strings.IndexAny(key, "&"); i >= 0 {
			key, query = key[:i], key[i+1:]
		} else {
			query = ""
		}
		if key == "" {
			continue
		}
		if i := strings.Index(key, "="); i >= 0 {
			key = key[:i]
		} else {
			continue
		}
		queryKeys = append(queryKeys, key)
	}
	sort.Strings(queryKeys)
	return strings.Join(queryKeys, "#")
}

func SortedQuery(query string) string {
	var m = make(map[string]string)
	var keys []string
	for query != "" {
		key := query
		if i := strings.IndexAny(key, "&"); i >= 0 {
			key, query = key[:i], key[i+1:]
		} else {
			query = ""
		}
		if key == "" {
			continue
		}
		value := ""
		if i := strings.Index(key, "="); i >= 0 {
			key, value = key[:i], key[i+1:]
		}
		m[key] = value
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var buf strings.Builder
	for _, key := range keys {
		value := m[key]
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(key)
		buf.WriteByte('=')
		buf.WriteString(value)
	}
	return buf.String()
}
