// Copyright 2021 EMQ Technologies Co., Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package function

import (
	"fmt"
	"strings"
)

func aggCall(name string, args []interface{}) (interface{}, bool) {
	lowerName := strings.ToLower(name)
	switch lowerName {
	case "avg":
		arg0 := args[0].([]interface{})
		c := getCount(arg0)
		if c > 0 {
			v := getFirstValidArg(arg0)
			switch v.(type) {
			case int, int64:
				if r, err := sliceIntTotal(arg0); err != nil {
					return err, false
				} else {
					return r / c, true
				}
			case float64:
				if r, err := sliceFloatTotal(arg0); err != nil {
					return err, false
				} else {
					return r / float64(c), true
				}
			case nil:
				return nil, true
			default:
				return fmt.Errorf("run avg function error: found invalid arg %[1]T(%[1]v)", v), false
			}
		}
		return 0, true
	case "count":
		arg0 := args[0].([]interface{})
		return getCount(arg0), true
	case "max":
		arg0 := args[0].([]interface{})
		if len(arg0) > 0 {
			v := getFirstValidArg(arg0)
			switch t := v.(type) {
			case int:
				if r, err := sliceIntMax(arg0, t); err != nil {
					return err, false
				} else {
					return r, true
				}
			case int64:
				if r, err := sliceIntMax(arg0, int(t)); err != nil {
					return err, false
				} else {
					return r, true
				}
			case float64:
				if r, err := sliceFloatMax(arg0, t); err != nil {
					return err, false
				} else {
					return r, true
				}
			case string:
				if r, err := sliceStringMax(arg0, t); err != nil {
					return err, false
				} else {
					return r, true
				}
			case nil:
				return nil, true
			default:
				return fmt.Errorf("run max function error: found invalid arg %[1]T(%[1]v)", v), false
			}
		}
		return fmt.Errorf("run max function error: empty data"), false
	case "min":
		arg0 := args[0].([]interface{})
		if len(arg0) > 0 {
			v := getFirstValidArg(arg0)
			switch t := v.(type) {
			case int:
				if r, err := sliceIntMin(arg0, t); err != nil {
					return err, false
				} else {
					return r, true
				}
			case int64:
				if r, err := sliceIntMin(arg0, int(t)); err != nil {
					return err, false
				} else {
					return r, true
				}
			case float64:
				if r, err := sliceFloatMin(arg0, t); err != nil {
					return err, false
				} else {
					return r, true
				}
			case string:
				if r, err := sliceStringMin(arg0, t); err != nil {
					return err, false
				} else {
					return r, true
				}
			case nil:
				return nil, true
			default:
				return fmt.Errorf("run min function error: found invalid arg %[1]T(%[1]v)", v), false
			}
		}
		return fmt.Errorf("run min function error: empty data"), false
	case "sum":
		arg0 := args[0].([]interface{})
		if len(arg0) > 0 {
			v := getFirstValidArg(arg0)
			switch v.(type) {
			case int, int64:
				if r, err := sliceIntTotal(arg0); err != nil {
					return err, false
				} else {
					return r, true
				}
			case float64:
				if r, err := sliceFloatTotal(arg0); err != nil {
					return err, false
				} else {
					return r, true
				}
			case nil:
				return nil, true
			default:
				return fmt.Errorf("run sum function error: found invalid arg %[1]T(%[1]v)", v), false
			}
		}
		return 0, true
	case "collect":
		return args[0], true
	case "deduplicate":
		v1, ok1 := args[0].([]interface{})
		v2, ok2 := args[1].([]interface{})
		v3a, ok3 := args[2].([]interface{})

		if ok1 && ok2 && ok3 && len(v3a) > 0 {
			v3, ok4 := getFirstValidArg(v3a).(bool)
			if ok4 {
				if r, err := dedup(v1, v2, v3); err != nil {
					return err, false
				} else {
					return r, true
				}
			}
		}
		return fmt.Errorf("Invalid argument type found."), false
	default:
		return fmt.Errorf("Unknown aggregate function name."), false
	}
}

func getCount(s []interface{}) int {
	c := 0
	for _, v := range s {
		if v != nil {
			c++
		}
	}
	return c
}

func getFirstValidArg(s []interface{}) interface{} {
	for _, v := range s {
		if v != nil {
			return v
		}
	}
	return nil
}

func sliceIntTotal(s []interface{}) (int, error) {
	var total int
	for _, v := range s {
		if vi, ok := v.(int); ok {
			total += vi
		} else if v != nil {
			return 0, fmt.Errorf("requires int but found %[1]T(%[1]v)", v)
		}
	}
	return total, nil
}

func sliceFloatTotal(s []interface{}) (float64, error) {
	var total float64
	for _, v := range s {
		if vf, ok := v.(float64); ok {
			total += vf
		} else if v != nil {
			return 0, fmt.Errorf("requires float64 but found %[1]T(%[1]v)", v)
		}
	}
	return total, nil
}
func sliceIntMax(s []interface{}, max int) (int, error) {
	for _, v := range s {
		if vi, ok := v.(int); ok {
			if max < vi {
				max = vi
			}
		} else if v != nil {
			return 0, fmt.Errorf("requires int but found %[1]T(%[1]v)", v)
		}
	}
	return max, nil
}
func sliceFloatMax(s []interface{}, max float64) (float64, error) {
	for _, v := range s {
		if vf, ok := v.(float64); ok {
			if max < vf {
				max = vf
			}
		} else if v != nil {
			return 0, fmt.Errorf("requires float64 but found %[1]T(%[1]v)", v)
		}
	}
	return max, nil
}

func sliceStringMax(s []interface{}, max string) (string, error) {
	for _, v := range s {
		if vs, ok := v.(string); ok {
			if max < vs {
				max = vs
			}
		} else if v != nil {
			return "", fmt.Errorf("requires string but found %[1]T(%[1]v)", v)
		}
	}
	return max, nil
}
func sliceIntMin(s []interface{}, min int) (int, error) {
	for _, v := range s {
		if vi, ok := v.(int); ok {
			if min > vi {
				min = vi
			}
		} else if v != nil {
			return 0, fmt.Errorf("requires int but found %[1]T(%[1]v)", v)
		}
	}
	return min, nil
}
func sliceFloatMin(s []interface{}, min float64) (float64, error) {
	for _, v := range s {
		if vf, ok := v.(float64); ok {
			if min > vf {
				min = vf
			}
		} else if v != nil {
			return 0, fmt.Errorf("requires float64 but found %[1]T(%[1]v)", v)
		}
	}
	return min, nil
}

func sliceStringMin(s []interface{}, min string) (string, error) {
	for _, v := range s {
		if vs, ok := v.(string); ok {
			if min < vs {
				min = vs
			}
		} else if v != nil {
			return "", fmt.Errorf("requires string but found %[1]T(%[1]v)", v)
		}
	}
	return min, nil
}

func dedup(r []interface{}, col []interface{}, all bool) (interface{}, error) {
	keyset := make(map[string]bool)
	result := make([]interface{}, 0)
	for i, m := range col {
		key := fmt.Sprintf("%v", m)
		if _, ok := keyset[key]; !ok {
			if all {
				result = append(result, r[i])
			} else if i == len(col)-1 {
				result = append(result, r[i])
			}
			keyset[key] = true
		}
	}
	if !all {
		if len(result) == 0 {
			return nil, nil
		} else {
			return result[0], nil
		}
	} else {
		return result, nil
	}
}
