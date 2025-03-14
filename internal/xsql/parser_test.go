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

package xsql

import (
	"fmt"
	"github.com/lf-edge/ekuiper/internal/testx"
	"github.com/lf-edge/ekuiper/pkg/ast"
	"math"
	"reflect"
	"strings"
	"testing"
)

// Ensure the parser can parse strings into Statement ASTs.
func TestParser_ParseStatement(t *testing.T) {
	var tests = []struct {
		s    string
		stmt *ast.SelectStatement
		err  string
	}{
		{
			s: `SELECT name FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream},
						Name:  "name",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},
		{
			s: "SELECT `select` FROM tbl",
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.FieldRef{Name: "select", StreamName: ast.DefaultStream},
						Name:  "select",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},
		{
			s: `SELECT name FROM topic/sensor1`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream},
						Name:  "name",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "topic/sensor1"}},
			},
		},

		{
			s: `SELECT t1.name FROM topic/sensor1 AS t1`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.FieldRef{StreamName: ast.StreamName("t1"), Name: "name"},
						Name:  "name",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "topic/sensor1", Alias: "t1"}},
			},
		},
		{
			s: "SELECT t1.name FROM topic/sensor1 AS `join`",
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.FieldRef{StreamName: ast.StreamName("t1"), Name: "name"},
						Name:  "name",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "topic/sensor1", Alias: "join"}},
			},
		},

		{
			s: `SELECT name FROM topic/sensor1 AS t1`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream},
						Name:  "name",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "topic/sensor1", Alias: "t1"}},
			},
		},

		{
			s: `SELECT name FROM topic/sensor1/#`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream},
						Name:  "name",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "topic/sensor1/#"}},
			},
		},

		{
			s: `SELECT name FROM topic/sensor1/# AS t2 `,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream},
						Name:  "name",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "topic/sensor1/#", Alias: "t2"}},
			},
		},

		{
			s: `SELECT name FROM /topic/sensor1/#`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream},
						Name:  "name",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "/topic/sensor1/#"}},
			},
		},

		{
			s: `SELECT name FROM /topic/sensor1/#/`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream},
						Name:  "name",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "/topic/sensor1/#/"}},
			},
		},

		{
			s: `SELECT name FROM /topic/sensor1/+/temp1/`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream},
						Name:  "name",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "/topic/sensor1/+/temp1/"}},
			},
		},

		{
			s: `SELECT name FROM topic/sensor1/+/temp`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream},
						Name:  "name",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "topic/sensor1/+/temp"}},
			},
		},

		{
			s: `SELECT * FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.Wildcard{Token: ast.ASTERISK},
						Name:  "",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},
		{
			s: `SELECT a,b FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{Expr: &ast.FieldRef{Name: "a", StreamName: ast.DefaultStream}, Name: "a", AName: ""},
					{Expr: &ast.FieldRef{Name: "b", StreamName: ast.DefaultStream}, Name: "b", AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},
		{
			s: `SELECT a, b,c FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{Expr: &ast.FieldRef{Name: "a", StreamName: ast.DefaultStream}, Name: "a", AName: ""},
					{Expr: &ast.FieldRef{Name: "b", StreamName: ast.DefaultStream}, Name: "b", AName: ""},
					{Expr: &ast.FieldRef{Name: "c", StreamName: ast.DefaultStream}, Name: "c", AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		{
			s: `SELECT a AS alias FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields:  []ast.Field{{Expr: &ast.FieldRef{Name: "a", StreamName: ast.DefaultStream}, Name: "a", AName: "alias"}},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		{
			s: `SELECT a AS alias1, b as Alias2 FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{Expr: &ast.FieldRef{Name: "a", StreamName: ast.DefaultStream}, Name: "a", AName: "alias1"},
					{Expr: &ast.FieldRef{Name: "b", StreamName: ast.DefaultStream}, Name: "b", AName: "Alias2"},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		{
			s: `SELECT length("test") FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						AName: "",
						Name:  "length",
						Expr: &ast.Call{
							Name: "length",
							Args: []ast.Expr{&ast.StringLiteral{Val: "test"}},
						},
					},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		{
			s: `SELECT length(test) FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						AName: "",
						Name:  "length",
						Expr: &ast.Call{
							Name: "length",
							Args: []ast.Expr{&ast.FieldRef{Name: "test", StreamName: ast.DefaultStream}},
						},
					},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		{
			s: `SELECT sin(123) FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						AName: "",
						Name:  "sin",
						Expr: &ast.Call{
							Name: "sin",
							Args: []ast.Expr{&ast.IntegerLiteral{Val: 123}},
						},
					},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		{
			s: `SELECT lpad("abc", 123) FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						AName: "",
						Name:  "lpad",
						Expr: &ast.Call{
							Name: "lpad",
							Args: []ast.Expr{&ast.StringLiteral{Val: "abc"}, &ast.IntegerLiteral{Val: 123}},
						},
					},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		{
			s: `SELECT newuuid() FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						AName: "",
						Name:  "newuuid",
						Expr: &ast.Call{
							Name: "newuuid",
							Args: nil,
						},
					},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		{
			s: `SELECT indexof("abc", field1) FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						AName: "",
						Name:  "indexof",
						Expr: &ast.Call{
							Name: "indexof",
							Args: []ast.Expr{
								&ast.StringLiteral{Val: "abc"},
								&ast.FieldRef{Name: "field1", StreamName: ast.DefaultStream},
							},
						},
					},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		{
			s: `SELECT lpad(lower(test),1) FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						AName: "",
						Name:  "lpad",
						Expr: &ast.Call{
							Name: "lpad",
							Args: []ast.Expr{
								&ast.Call{
									Name: "lower",
									Args: []ast.Expr{
										&ast.FieldRef{Name: "test", StreamName: ast.DefaultStream},
									},
								},
								&ast.IntegerLiteral{Val: 1},
							},
						},
					},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		{
			s: `SELECT lpad(lower(test),1) AS field1 FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						AName: "field1",
						Name:  "lpad",
						Expr: &ast.Call{
							Name: "lpad",
							Args: []ast.Expr{
								&ast.Call{
									Name: "lower",
									Args: []ast.Expr{
										&ast.FieldRef{Name: "test", StreamName: ast.DefaultStream},
									},
								},
								&ast.IntegerLiteral{Val: 1},
							},
						},
					},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		{
			s: `SELECT length(lower("test")) FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						AName: "",
						Name:  "length",
						Expr: &ast.Call{
							Name: "length",
							Args: []ast.Expr{
								&ast.Call{
									Name: "lower",
									Args: []ast.Expr{
										&ast.StringLiteral{Val: "test"},
									},
								},
							},
						},
					},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		{
			s: `SELECT count(*) FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						AName: "",
						Name:  "count",
						Expr: &ast.Call{
							Name: "count",
							Args: []ast.Expr{&ast.Wildcard{Token: ast.ASTERISK}},
						},
					},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		{
			s:    `SELECT count(*, f1) FROM tbl`,
			stmt: nil,
			err:  `found ",", expected right paren.`,
		},

		{
			s: `SELECT deduplicate(temperature, false) FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						AName: "",
						Name:  "deduplicate",
						Expr: &ast.Call{
							Name: "deduplicate",
							Args: []ast.Expr{&ast.Wildcard{Token: ast.ASTERISK}, &ast.FieldRef{Name: "temperature", StreamName: ast.DefaultStream}, &ast.BooleanLiteral{Val: false}},
						},
					},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		{
			s: `SELECT "abc" FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields:  []ast.Field{{AName: "", Name: "", Expr: &ast.StringLiteral{Val: "abc"}}},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		{
			s: `SELECT "abc" AS field1 FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields:  []ast.Field{{AName: "field1", Name: "", Expr: &ast.StringLiteral{Val: "abc"}}},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		{
			s: `SELECT field0,   "abc" AS field1, field2 FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{AName: "", Name: "field0", Expr: &ast.FieldRef{Name: "field0", StreamName: ast.DefaultStream}},
					{AName: "field1", Name: "", Expr: &ast.StringLiteral{Val: "abc"}},
					{AName: "", Name: "field2", Expr: &ast.FieldRef{Name: "field2", StreamName: ast.DefaultStream}}},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		{
			s:    `SELECT * AS alias FROM tbl`,
			stmt: nil,
			err:  `found "AS", expected FROM.`,
		},

		{
			s:    `SELECT *, FROM tbl`,
			stmt: nil,
			err:  `found ",", expected FROM.`,
		},

		{
			s:    `SELECTname FROM tbl`,
			stmt: nil,
			err:  `Found "SELECTname", Expected SELECT.` + "\n",
		},

		{
			s: `SELECT abc FROM tbl WHERE abc > 12 `,
			stmt: &ast.SelectStatement{
				Fields:  []ast.Field{{AName: "", Name: "abc", Expr: &ast.FieldRef{Name: "abc", StreamName: ast.DefaultStream}}},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
				Condition: &ast.BinaryExpr{
					LHS: &ast.FieldRef{Name: "abc", StreamName: ast.DefaultStream},
					OP:  ast.GT,
					RHS: &ast.IntegerLiteral{Val: 12},
				},
			},
		},

		{
			s: `SELECT abc FROM tbl WHERE abc = "hello" `,
			stmt: &ast.SelectStatement{
				Fields:  []ast.Field{{AName: "", Name: "abc", Expr: &ast.FieldRef{Name: "abc", StreamName: ast.DefaultStream}}},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
				Condition: &ast.BinaryExpr{
					LHS: &ast.FieldRef{Name: "abc", StreamName: ast.DefaultStream},
					OP:  ast.EQ,
					RHS: &ast.StringLiteral{Val: "hello"},
				},
			},
		},

		{
			s: `SELECT t1.abc FROM tbl AS t1 WHERE t1.abc = "hello" `,
			stmt: &ast.SelectStatement{
				Fields:  []ast.Field{{AName: "", Name: "abc", Expr: &ast.FieldRef{StreamName: ast.StreamName("t1"), Name: "abc"}}},
				Sources: []ast.Source{&ast.Table{Name: "tbl", Alias: "t1"}},
				Condition: &ast.BinaryExpr{
					LHS: &ast.FieldRef{StreamName: ast.StreamName("t1"), Name: "abc"},
					OP:  ast.EQ,
					RHS: &ast.StringLiteral{Val: "hello"},
				},
			},
		},

		{
			s: `SELECT abc, "fff" AS fa FROM tbl WHERE fa >= 5 `,
			stmt: &ast.SelectStatement{
				Fields:  []ast.Field{{AName: "", Name: "abc", Expr: &ast.FieldRef{Name: "abc", StreamName: ast.DefaultStream}}, {AName: "fa", Name: "", Expr: &ast.StringLiteral{Val: "fff"}}},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
				Condition: &ast.BinaryExpr{
					LHS: &ast.FieldRef{Name: "fa", StreamName: ast.DefaultStream},
					OP:  ast.GTE,
					RHS: &ast.IntegerLiteral{Val: 5},
				},
			},
		},

		{
			s: `SELECT field2 FROM tbl WHERE field2 != 5 `,
			stmt: &ast.SelectStatement{
				Fields:  []ast.Field{{AName: "", Name: "field2", Expr: &ast.FieldRef{Name: "field2", StreamName: ast.DefaultStream}}},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
				Condition: &ast.BinaryExpr{
					LHS: &ast.FieldRef{Name: "field2", StreamName: ast.DefaultStream},
					OP:  ast.NEQ,
					RHS: &ast.IntegerLiteral{Val: 5},
				},
			},
		},

		{
			s: `SELECT field2 FROM tbl WHERE field2 !   = 5 `, //Add space char in expression
			stmt: &ast.SelectStatement{
				Fields:  []ast.Field{{AName: "", Name: "field2", Expr: &ast.FieldRef{Name: "field2", StreamName: ast.DefaultStream}}},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
				Condition: &ast.BinaryExpr{
					LHS: &ast.FieldRef{Name: "field2", StreamName: ast.DefaultStream},
					OP:  ast.NEQ,
					RHS: &ast.IntegerLiteral{Val: 5},
				},
			},
		},

		{
			s:    `SELECT *f FROM tbl`,
			stmt: nil,
			err:  `found "f", expected FROM.`,
		},

		////TODO
		//{
		//	s: `SELECT *from FROM tbl`,
		//	stmt: nil,
		//	err: `found "f", expected FROM.`,
		//},

		{
			s: `SELECT abc+2 FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						AName: "",
						Name:  "",
						Expr: &ast.BinaryExpr{
							LHS: &ast.FieldRef{Name: "abc", StreamName: ast.DefaultStream},
							OP:  ast.ADD,
							RHS: &ast.IntegerLiteral{Val: 2},
						},
					},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		{
			s: `SELECT t1.abc+2 FROM tbl AS t1`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						AName: "",
						Name:  "",
						Expr: &ast.BinaryExpr{
							LHS: &ast.FieldRef{StreamName: ast.StreamName("t1"), Name: "abc"},
							OP:  ast.ADD,
							RHS: &ast.IntegerLiteral{Val: 2},
						},
					},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl", Alias: "t1"}},
			},
		},

		{
			s: `SELECT abc + "hello" FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						AName: "",
						Name:  "",
						Expr: &ast.BinaryExpr{
							LHS: &ast.FieldRef{Name: "abc", StreamName: ast.DefaultStream},
							OP:  ast.ADD,
							RHS: &ast.StringLiteral{Val: "hello"},
						},
					},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		{
			s: `SELECT abc*2 + 3 FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						AName: "",
						Name:  "",
						Expr: &ast.BinaryExpr{
							LHS: &ast.BinaryExpr{
								LHS: &ast.FieldRef{Name: "abc", StreamName: ast.DefaultStream},
								OP:  ast.MUL,
								RHS: &ast.IntegerLiteral{Val: 2},
							},
							OP:  ast.ADD,
							RHS: &ast.IntegerLiteral{Val: 3},
						},
					},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		{
			s: `SELECT ln(abc*2 + 3) FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						AName: "",
						Name:  "ln",
						Expr: &ast.Call{
							Name: "ln",
							Args: []ast.Expr{
								&ast.BinaryExpr{
									LHS: &ast.BinaryExpr{
										LHS: &ast.FieldRef{Name: "abc", StreamName: ast.DefaultStream},
										OP:  ast.MUL,
										RHS: &ast.IntegerLiteral{Val: 2},
									},
									OP:  ast.ADD,
									RHS: &ast.IntegerLiteral{Val: 3},
								},
							},
						},
					},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		{
			s: `SELECT ln(t1.abc*2 + 3) FROM tbl AS t1`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						AName: "",
						Name:  "ln",
						Expr: &ast.Call{
							Name: "ln",
							Args: []ast.Expr{
								&ast.BinaryExpr{
									LHS: &ast.BinaryExpr{
										LHS: &ast.FieldRef{StreamName: ast.StreamName("t1"), Name: "abc"},
										OP:  ast.MUL,
										RHS: &ast.IntegerLiteral{Val: 2},
									},
									OP:  ast.ADD,
									RHS: &ast.IntegerLiteral{Val: 3},
								},
							},
						},
					},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl", Alias: "t1"}},
			},
		},

		{
			s: `SELECT lpad("param2", abc*2 + 3) FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						AName: "",
						Name:  "lpad",
						Expr: &ast.Call{
							Name: "lpad",
							Args: []ast.Expr{
								&ast.StringLiteral{Val: "param2"},
								&ast.BinaryExpr{
									LHS: &ast.BinaryExpr{
										LHS: &ast.FieldRef{Name: "abc", StreamName: ast.DefaultStream},
										OP:  ast.MUL,
										RHS: &ast.IntegerLiteral{Val: 2},
									},
									OP:  ast.ADD,
									RHS: &ast.IntegerLiteral{Val: 3},
								},
							},
						},
					},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		{
			s: `SELECT 0.2 FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						AName: "",
						Name:  "",
						Expr:  &ast.NumberLiteral{Val: 0.2},
					},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		{
			s: `SELECT .2 FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						AName: "",
						Name:  "",
						Expr:  &ast.NumberLiteral{Val: 0.2},
					},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		{
			s: `SELECT sin(.2) FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						AName: "",
						Name:  "sin",
						Expr: &ast.Call{
							Name: "sin",
							Args: []ast.Expr{&ast.NumberLiteral{Val: 0.2}},
						},
					},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		{
			s: `SELECT power(.2, 4) FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						AName: "",
						Name:  "power",
						Expr: &ast.Call{
							Name: "power",
							Args: []ast.Expr{&ast.NumberLiteral{Val: 0.2}, &ast.IntegerLiteral{Val: 4}},
						},
					},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		{
			s: `SELECT power(.2, 4) AS f1 FROM tbl WHERE f1 > 2.2`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						AName: "f1",
						Name:  "power",
						Expr: &ast.Call{
							Name: "power",
							Args: []ast.Expr{&ast.NumberLiteral{Val: 0.2}, &ast.IntegerLiteral{Val: 4}},
						},
					},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
				Condition: &ast.BinaryExpr{
					LHS: &ast.FieldRef{Name: "f1", StreamName: ast.DefaultStream},
					OP:  ast.GT,
					RHS: &ast.NumberLiteral{Val: 2.2},
				},
			},
		},

		{
			s: `SELECT deviceId, name FROM topic/sensor1 WHERE deviceId=1 AND name = "dname"`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{Expr: &ast.FieldRef{Name: "deviceId", StreamName: ast.DefaultStream}, Name: "deviceId", AName: ""},
					{Expr: &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream}, Name: "name", AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "topic/sensor1"}},
				Condition: &ast.BinaryExpr{
					LHS: &ast.BinaryExpr{LHS: &ast.FieldRef{Name: "deviceId", StreamName: ast.DefaultStream}, OP: ast.EQ, RHS: &ast.IntegerLiteral{Val: 1}},
					OP:  ast.AND,
					RHS: &ast.BinaryExpr{LHS: &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream}, OP: ast.EQ, RHS: &ast.StringLiteral{Val: "dname"}},
				},
			},
		},

		{
			s: `SELECT deviceId, name FROM topic/sensor1 AS t1 WHERE t1.deviceId=1 AND t1.name = "dname"`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{Expr: &ast.FieldRef{Name: "deviceId", StreamName: ast.DefaultStream}, Name: "deviceId", AName: ""},
					{Expr: &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream}, Name: "name", AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "topic/sensor1", Alias: "t1"}},
				Condition: &ast.BinaryExpr{
					LHS: &ast.BinaryExpr{LHS: &ast.FieldRef{StreamName: ast.StreamName("t1"), Name: "deviceId"}, OP: ast.EQ, RHS: &ast.IntegerLiteral{Val: 1}},
					OP:  ast.AND,
					RHS: &ast.BinaryExpr{LHS: &ast.FieldRef{StreamName: ast.StreamName("t1"), Name: "name"}, OP: ast.EQ, RHS: &ast.StringLiteral{Val: "dname"}},
				},
			},
		},

		{
			s: `SELECT temp AS t, name FROM topic/sensor1 WHERE t> = 20.5 OR name = "dname"`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{Expr: &ast.FieldRef{Name: "temp", StreamName: ast.DefaultStream}, Name: "temp", AName: "t"},
					{Expr: &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream}, Name: "name", AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "topic/sensor1"}},
				Condition: &ast.BinaryExpr{
					LHS: &ast.BinaryExpr{LHS: &ast.FieldRef{Name: "t", StreamName: ast.DefaultStream}, OP: ast.GTE, RHS: &ast.NumberLiteral{Val: 20.5}},
					OP:  ast.OR,
					RHS: &ast.BinaryExpr{LHS: &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream}, OP: ast.EQ, RHS: &ast.StringLiteral{Val: "dname"}},
				},
			},
		},

		{
			s: `SELECT temp AS t, name FROM topic/sensor1 WHERE name = "dname" GROUP BY name`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{Expr: &ast.FieldRef{Name: "temp", StreamName: ast.DefaultStream}, Name: "temp", AName: "t"},
					{Expr: &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream}, Name: "name", AName: ""},
				},
				Sources:    []ast.Source{&ast.Table{Name: "topic/sensor1"}},
				Condition:  &ast.BinaryExpr{LHS: &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream}, OP: ast.EQ, RHS: &ast.StringLiteral{Val: "dname"}},
				Dimensions: ast.Dimensions{ast.Dimension{Expr: &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream}}},
			},
		},

		{
			s: `SELECT temp AS t, name FROM topic/sensor1 WHERE name = "dname" GROUP BY name HAVING count(name) > 3`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{Expr: &ast.FieldRef{Name: "temp", StreamName: ast.DefaultStream}, Name: "temp", AName: "t"},
					{Expr: &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream}, Name: "name", AName: ""},
				},
				Sources:    []ast.Source{&ast.Table{Name: "topic/sensor1"}},
				Condition:  &ast.BinaryExpr{LHS: &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream}, OP: ast.EQ, RHS: &ast.StringLiteral{Val: "dname"}},
				Dimensions: ast.Dimensions{ast.Dimension{Expr: &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream}}},
				Having:     &ast.BinaryExpr{LHS: &ast.Call{Name: "count", Args: []ast.Expr{&ast.FieldRef{StreamName: ast.DefaultStream, Name: "name"}}}, OP: ast.GT, RHS: &ast.IntegerLiteral{Val: 3}},
			},
		},

		{
			s: `SELECT temp AS t, name FROM topic/sensor1 WHERE name = "dname" HAVING count(name) > 3`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{Expr: &ast.FieldRef{Name: "temp", StreamName: ast.DefaultStream}, Name: "temp", AName: "t"},
					{Expr: &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream}, Name: "name", AName: ""},
				},
				Sources:   []ast.Source{&ast.Table{Name: "topic/sensor1"}},
				Condition: &ast.BinaryExpr{LHS: &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream}, OP: ast.EQ, RHS: &ast.StringLiteral{Val: "dname"}},
				Having:    &ast.BinaryExpr{LHS: &ast.Call{Name: "count", Args: []ast.Expr{&ast.FieldRef{StreamName: ast.DefaultStream, Name: "name"}}}, OP: ast.GT, RHS: &ast.IntegerLiteral{Val: 3}},
			},
		},

		{
			s:    `SELECT id,AVG(data) FROM t GROUP BY SUM(data)>10`,
			stmt: nil,
			err:  "Not allowed to call aggregate functions in GROUP BY clause.",
		},

		{
			s:    `SELECT temp AS t, name FROM topic/sensor1 WHERE count(name) = 3`,
			stmt: nil,
			err:  "Not allowed to call aggregate functions in WHERE clause.",
		},

		{
			s: `SELECT s1.temp AS t, name FROM topic/sensor1 AS s1 WHERE t = "dname" GROUP BY s1.temp`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{Expr: &ast.FieldRef{StreamName: "s1", Name: "temp"}, Name: "temp", AName: "t"},
					{Expr: &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream}, Name: "name", AName: ""},
				},
				Sources:    []ast.Source{&ast.Table{Name: "topic/sensor1", Alias: "s1"}},
				Condition:  &ast.BinaryExpr{LHS: &ast.FieldRef{Name: "t", StreamName: ast.DefaultStream}, OP: ast.EQ, RHS: &ast.StringLiteral{Val: "dname"}},
				Dimensions: ast.Dimensions{ast.Dimension{Expr: &ast.FieldRef{StreamName: "s1", Name: "temp"}}},
			},
		},

		{
			s: `SELECT temp AS t, name FROM topic/sensor1 WHERE name = "dname" GROUP BY lpad(name,1)`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{Expr: &ast.FieldRef{Name: "temp", StreamName: ast.DefaultStream}, Name: "temp", AName: "t"},
					{Expr: &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream}, Name: "name", AName: ""},
				},
				Sources:   []ast.Source{&ast.Table{Name: "topic/sensor1"}},
				Condition: &ast.BinaryExpr{LHS: &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream}, OP: ast.EQ, RHS: &ast.StringLiteral{Val: "dname"}},
				Dimensions: ast.Dimensions{ast.Dimension{
					Expr: &ast.Call{Name: "lpad", Args: []ast.Expr{&ast.FieldRef{Name: "name", StreamName: ast.DefaultStream}, &ast.IntegerLiteral{Val: 1}}},
				},
				},
			},
		},

		{
			s: `SELECT temp AS t, name FROM topic/sensor1 AS s1 WHERE name = "dname" GROUP BY lpad(s1.name,1)`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{Expr: &ast.FieldRef{Name: "temp", StreamName: ast.DefaultStream}, Name: "temp", AName: "t"},
					{Expr: &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream}, Name: "name", AName: ""},
				},
				Sources:   []ast.Source{&ast.Table{Name: "topic/sensor1", Alias: "s1"}},
				Condition: &ast.BinaryExpr{LHS: &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream}, OP: ast.EQ, RHS: &ast.StringLiteral{Val: "dname"}},
				Dimensions: ast.Dimensions{ast.Dimension{
					Expr: &ast.Call{Name: "lpad", Args: []ast.Expr{&ast.FieldRef{StreamName: ast.StreamName("s1"), Name: "name"}, &ast.IntegerLiteral{Val: 1}}},
				},
				},
			},
		},

		{
			s: `SELECT temp AS t, name FROM topic/sensor1 WHERE name = "dname" GROUP BY lpad(name,1) ORDER BY name`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{Expr: &ast.FieldRef{Name: "temp", StreamName: ast.DefaultStream}, Name: "temp", AName: "t"},
					{Expr: &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream}, Name: "name", AName: ""},
				},
				Sources:   []ast.Source{&ast.Table{Name: "topic/sensor1"}},
				Condition: &ast.BinaryExpr{LHS: &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream}, OP: ast.EQ, RHS: &ast.StringLiteral{Val: "dname"}},
				Dimensions: ast.Dimensions{
					ast.Dimension{
						Expr: &ast.Call{Name: "lpad", Args: []ast.Expr{
							&ast.FieldRef{Name: "name", StreamName: ast.DefaultStream},
							&ast.IntegerLiteral{Val: 1}},
						},
					},
				},
				SortFields: []ast.SortField{{Name: "name", Ascending: true}},
			},
		},

		{
			s: `SELECT temp AS t, name FROM topic/sensor1 AS s1 WHERE s1.name = "dname" GROUP BY lpad(s1.name,1) ORDER BY s1.name`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{Expr: &ast.FieldRef{Name: "temp", StreamName: ast.DefaultStream}, Name: "temp", AName: "t"},
					{Expr: &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream}, Name: "name", AName: ""},
				},
				Sources:   []ast.Source{&ast.Table{Name: "topic/sensor1", Alias: "s1"}},
				Condition: &ast.BinaryExpr{LHS: &ast.FieldRef{StreamName: ast.StreamName("s1"), Name: "name"}, OP: ast.EQ, RHS: &ast.StringLiteral{Val: "dname"}},
				Dimensions: ast.Dimensions{
					ast.Dimension{
						Expr: &ast.Call{Name: "lpad", Args: []ast.Expr{
							&ast.FieldRef{StreamName: ast.StreamName("s1"), Name: "name"},
							&ast.IntegerLiteral{Val: 1}},
						},
					},
				},
				SortFields: []ast.SortField{{Name: "s1\007name", Ascending: true}},
			},
		},

		{
			s: `SELECT temp AS t, name FROM topic/sensor1 WHERE name = "dname" GROUP BY lpad(name,1) ORDER BY name DESC`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{Expr: &ast.FieldRef{Name: "temp", StreamName: ast.DefaultStream}, Name: "temp", AName: "t"},
					{Expr: &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream}, Name: "name", AName: ""},
				},
				Sources:   []ast.Source{&ast.Table{Name: "topic/sensor1"}},
				Condition: &ast.BinaryExpr{LHS: &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream}, OP: ast.EQ, RHS: &ast.StringLiteral{Val: "dname"}},
				Dimensions: ast.Dimensions{
					ast.Dimension{
						Expr: &ast.Call{Name: "lpad", Args: []ast.Expr{
							&ast.FieldRef{Name: "name", StreamName: ast.DefaultStream},
							&ast.IntegerLiteral{Val: 1}},
						},
					},
				},
				SortFields: []ast.SortField{{Name: "name", Ascending: false}},
			},
		},

		{
			s: `SELECT * FROM topic/sensor1 ORDER BY name DESC`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.Wildcard{Token: ast.ASTERISK},
						Name:  "",
						AName: ""},
				},
				Sources:    []ast.Source{&ast.Table{Name: "topic/sensor1"}},
				SortFields: []ast.SortField{{Name: "name", Ascending: false}},
			},
		},

		{
			s: `SELECT * FROM topic/sensor1 ORDER BY name DESC, name2 ASC`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.Wildcard{Token: ast.ASTERISK},
						Name:  "",
						AName: ""},
				},
				Sources:    []ast.Source{&ast.Table{Name: "topic/sensor1"}},
				SortFields: []ast.SortField{{Name: "name", Ascending: false}, {Name: "name2", Ascending: true}},
			},
		},

		{
			s: `SELECT * FROM topic/sensor1 GROUP BY name, name2,power(name3,1.8) ORDER BY name DESC, name2 ASC`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.Wildcard{Token: ast.ASTERISK},
						Name:  "",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "topic/sensor1"}},
				Dimensions: ast.Dimensions{
					ast.Dimension{Expr: &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream}},
					ast.Dimension{Expr: &ast.FieldRef{Name: "name2", StreamName: ast.DefaultStream}},
					ast.Dimension{
						Expr: &ast.Call{Name: "power", Args: []ast.Expr{
							&ast.FieldRef{Name: "name3", StreamName: ast.DefaultStream},
							&ast.NumberLiteral{Val: 1.8}},
						},
					},
				},
				SortFields: []ast.SortField{{Name: "name", Ascending: false}, {Name: "name2", Ascending: true}},
			},
		},

		//{
		//	s: `SELECT .2sd FROM tbl`,
		//	stmt: &SelectStatement{
		//		Fields:    []Field{
		//			Field{
		//				AName:"",
		//				Expr: &NumberLiteral{Val: 0.2},
		//			},
		//		},
		//		TableName: "tbl",
		//	},
		//},

		{
			s: `SELECT name FROM tbl/*SELECT comment FROM testComments*/`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream},
						Name:  "name",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		{
			s: `/*SELECT comment FROM testComments*/SELECT name FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream},
						Name:  "name",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		{
			s: `SELECT name /*SELECT comment FROM testComments*/ FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream},
						Name:  "name",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		{
			s: `SELECT true AS f1, FALSE as f2 FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{AName: "f1", Name: "", Expr: &ast.BooleanLiteral{Val: true}},
					{AName: "f2", Name: "", Expr: &ast.BooleanLiteral{Val: false}},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		{
			s: `SELECT true AS f1 FROM tbl WHERE f2 = true`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{AName: "f1", Name: "", Expr: &ast.BooleanLiteral{Val: true}},
				},
				Sources:   []ast.Source{&ast.Table{Name: "tbl"}},
				Condition: &ast.BinaryExpr{LHS: &ast.FieldRef{Name: "f2", StreamName: ast.DefaultStream}, OP: ast.EQ, RHS: &ast.BooleanLiteral{Val: true}},
			},
		},

		{
			s: `SELECT indexof(field1, "abc") FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						AName: "",
						Name:  "indexof",
						Expr: &ast.Call{
							Name: "indexof",
							Args: []ast.Expr{&ast.FieldRef{Name: "field1", StreamName: ast.DefaultStream}, &ast.StringLiteral{Val: "abc"}},
						},
					},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		//The negative value expression support.
		{
			s: `SELECT -3 AS t1 FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.IntegerLiteral{Val: -3},
						Name:  "",
						AName: "t1"},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		{
			s: `SELECT - 3 AS t1 FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.IntegerLiteral{Val: -3},
						Name:  "",
						AName: "t1"},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		{
			s: `SELECT -. 3 AS t1 FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.NumberLiteral{Val: -.3},
						Name:  "",
						AName: "t1"},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		{
			s: `SELECT -3.3 AS t1 FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.NumberLiteral{Val: -3.3},
						Name:  "",
						AName: "t1"},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},

		{
			s:    `SELECT sample(-.3,) FROM tbl`,
			stmt: nil,
			err:  "function sample not found",
		},

		{
			s:    `select timestamp() as tp from demo`,
			stmt: nil,
			err:  "found \"TIMESTAMP\", expected expression.",
		},

		{
			s: `select tstamp() as tp from demo`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr: &ast.Call{
							Name: "tstamp",
							Args: nil,
						},
						Name:  "tstamp",
						AName: "tp"},
				},
				Sources: []ast.Source{&ast.Table{Name: "demo"}},
			},
			err: "",
		},

		{
			s:    "SELECT `half FROM tb",
			stmt: nil,
			err:  `found "EOF", expected FROM.`,
		},
		{
			s: "SELECT `space var` FROM tbl",
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.FieldRef{Name: "space var", StreamName: ast.DefaultStream},
						Name:  "space var",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},
		{
			s: "SELECT `中文 Chinese` FROM tbl",
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.FieldRef{Name: "中文 Chinese", StreamName: ast.DefaultStream},
						Name:  "中文 Chinese",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		}, {
			s: "SELECT CASE temperature WHEN 25 THEN \"bingo\" WHEN 30 THEN \"high\" ELSE \"low\" END as label, humidity FROM tbl",
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr: &ast.CaseExpr{
							Value: &ast.FieldRef{Name: "temperature", StreamName: ast.DefaultStream},
							WhenClauses: []*ast.WhenClause{
								{
									Expr:   &ast.IntegerLiteral{Val: 25},
									Result: &ast.StringLiteral{Val: "bingo"},
								}, {
									Expr:   &ast.IntegerLiteral{Val: 30},
									Result: &ast.StringLiteral{Val: "high"},
								},
							},
							ElseClause: &ast.StringLiteral{Val: "low"},
						},
						Name:  "",
						AName: "label",
					}, {
						Expr:  &ast.FieldRef{Name: "humidity", StreamName: ast.DefaultStream},
						Name:  "humidity",
						AName: "",
					},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		}, {
			s: "SELECT CASE temperature WHEN 25 THEN \"bingo\" WHEN 30 THEN \"high\" END as label, humidity FROM tbl",
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr: &ast.CaseExpr{
							Value: &ast.FieldRef{Name: "temperature", StreamName: ast.DefaultStream},
							WhenClauses: []*ast.WhenClause{
								{
									Expr:   &ast.IntegerLiteral{Val: 25},
									Result: &ast.StringLiteral{Val: "bingo"},
								}, {
									Expr:   &ast.IntegerLiteral{Val: 30},
									Result: &ast.StringLiteral{Val: "high"},
								},
							},
							ElseClause: nil,
						},
						Name:  "",
						AName: "label",
					}, {
						Expr:  &ast.FieldRef{Name: "humidity", StreamName: ast.DefaultStream},
						Name:  "humidity",
						AName: "",
					},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		}, {
			s:    "SELECT CASE temperature ELSE \"low\" END as label, humidity FROM tbl",
			stmt: nil,
			err:  "invalid CASE expression, WHEN expected before ELSE",
		}, {
			s: "SELECT CASE WHEN temperature > 30 THEN \"high\" ELSE \"low\" END as label, humidity FROM tbl",
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr: &ast.CaseExpr{
							Value: nil,
							WhenClauses: []*ast.WhenClause{
								{
									Expr: &ast.BinaryExpr{
										OP:  ast.GT,
										LHS: &ast.FieldRef{Name: "temperature", StreamName: ast.DefaultStream},
										RHS: &ast.IntegerLiteral{Val: 30},
									},
									Result: &ast.StringLiteral{Val: "high"},
								},
							},
							ElseClause: &ast.StringLiteral{Val: "low"},
						},
						Name:  "",
						AName: "label",
					}, {
						Expr:  &ast.FieldRef{Name: "humidity", StreamName: ast.DefaultStream},
						Name:  "humidity",
						AName: "",
					},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		}, {
			s:    "SELECT CASE WHEN 30 THEN \"high\" ELSE \"low\" END as label, humidity FROM tbl",
			stmt: nil,
			err:  "invalid CASE expression, WHEN expression must be a bool condition",
		}, {
			s: `SELECT echo(*) FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						AName: "",
						Name:  "echo",
						Expr: &ast.Call{
							Name: "echo",
							Args: []ast.Expr{&ast.Wildcard{Token: ast.ASTERISK}},
						},
					},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},
		{
			s: `SELECT count(*)-10 FROM demo`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						AName: "",
						Name:  "",
						Expr: &ast.BinaryExpr{
							OP: ast.SUB,
							LHS: &ast.Call{
								Name: "count",
								Args: []ast.Expr{
									&ast.Wildcard{Token: ast.ASTERISK},
								},
							},
							RHS: &ast.IntegerLiteral{Val: 10},
						},
					},
				},
				Sources: []ast.Source{&ast.Table{Name: "demo"}},
			},
		},
		{
			s:    `SELECT -abc FROM demo`,
			stmt: nil,
			err:  "found \"-\", expected expression.",
		},
		{
			s: `SELECT meta(*) FROM tbl`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						AName: "",
						Name:  "meta",
						Expr: &ast.Call{
							Name: "meta",
							Args: []ast.Expr{
								&ast.MetaRef{
									Name:       "*",
									StreamName: ast.DefaultStream,
								},
							},
						},
					},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
			},
		},
	}

	fmt.Printf("The test bucket size is %d.\n\n", len(tests))
	for i, tt := range tests {
		//fmt.Printf("Parsing SQL %q.\n", tt.s)
		stmt, err := NewParser(strings.NewReader(tt.s)).Parse()
		if !reflect.DeepEqual(tt.err, testx.Errstring(err)) {
			t.Errorf("%d. %q: error mismatch:\n  exp=%s\n  got=%s\n\n", i, tt.s, tt.err, err)
		} else if tt.err == "" && !reflect.DeepEqual(tt.stmt, stmt) {
			t.Errorf("%d. %q\n\nstmt mismatch:\n\nexp=%#v\n\ngot=%#v\n\n", i, tt.s, tt.stmt, stmt)
		}
	}
}

func TestParser_ParseWindowsExpr(t *testing.T) {
	var tests = []struct {
		s    string
		stmt *ast.SelectStatement
		err  string
	}{
		{
			s: `SELECT f1 FROM tbl GROUP BY TUMBLINGWINDOW(ss, 10)`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.FieldRef{Name: "f1", StreamName: ast.DefaultStream},
						Name:  "f1",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
				Dimensions: ast.Dimensions{
					ast.Dimension{
						Expr: &ast.Window{
							WindowType: ast.TUMBLING_WINDOW,
							Length:     &ast.IntegerLiteral{Val: 10000},
							Interval:   &ast.IntegerLiteral{Val: 0},
						},
					},
				},
			},
		},

		{
			s: `SELECT f1 FROM tbl GROUP BY HOPPINGWINDOW(mi, 5, 1)`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.FieldRef{Name: "f1", StreamName: ast.DefaultStream},
						Name:  "f1",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
				Dimensions: ast.Dimensions{
					ast.Dimension{
						Expr: &ast.Window{
							WindowType: ast.HOPPING_WINDOW,
							Length:     &ast.IntegerLiteral{Val: 3e5},
							Interval:   &ast.IntegerLiteral{Val: 6e4},
						},
					},
				},
			},
		},

		{
			s: `SELECT f1 FROM tbl GROUP BY SESSIONWINDOW(hh, 5, 1)`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.FieldRef{Name: "f1", StreamName: ast.DefaultStream},
						Name:  "f1",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
				Dimensions: ast.Dimensions{
					ast.Dimension{
						Expr: &ast.Window{
							WindowType: ast.SESSION_WINDOW,
							Length:     &ast.IntegerLiteral{Val: 1.8e7},
							Interval:   &ast.IntegerLiteral{Val: 3.6e6},
						},
					},
				},
			},
		},

		{
			s: `SELECT f1 FROM tbl GROUP BY SLIDINGWINDOW(ms, 5)`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.FieldRef{Name: "f1", StreamName: ast.DefaultStream},
						Name:  "f1",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
				Dimensions: ast.Dimensions{
					ast.Dimension{
						Expr: &ast.Window{
							WindowType: ast.SLIDING_WINDOW,
							Length:     &ast.IntegerLiteral{Val: 5},
							Interval:   &ast.IntegerLiteral{Val: 0},
						},
					},
				},
			},
		},

		{
			s:    `SELECT f1 FROM tbl GROUP BY SLIDINGWINDOW(mi, 5, 1)`,
			stmt: nil,
			err:  "The arguments for slidingwindow should be 2.\n",
		},

		{
			s:    `SELECT f1 FROM tbl GROUP BY SLIDINGWINDOW("mi", 5)`,
			stmt: nil,
			err:  "The 1st argument for slidingwindow is expecting timer literal expression. One value of [dd|hh|mi|ss|ms].\n",
		},

		{
			s: `SELECT f1 FROM tbl GROUP BY COUNTWINDOW(10)`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.FieldRef{Name: "f1", StreamName: ast.DefaultStream},
						Name:  "f1",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
				Dimensions: ast.Dimensions{
					ast.Dimension{
						Expr: &ast.Window{
							WindowType: ast.COUNT_WINDOW,
							Length:     &ast.IntegerLiteral{Val: 10},
						},
					},
				},
			},
		},

		{
			s: `SELECT f1 FROM tbl GROUP BY COUNTWINDOW(10, 5)`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.FieldRef{Name: "f1", StreamName: ast.DefaultStream},
						Name:  "f1",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "tbl"}},
				Dimensions: ast.Dimensions{
					ast.Dimension{
						Expr: &ast.Window{
							WindowType: ast.COUNT_WINDOW,
							Length:     &ast.IntegerLiteral{Val: 10},
							Interval:   &ast.IntegerLiteral{Val: 5},
						},
					},
				},
			},
		},

		{
			s:    `SELECT f1 FROM tbl GROUP BY COUNTWINDOW(3, 5)`,
			stmt: nil,
			err:  "The second parameter value 5 should be less than the first parameter 3.",
		},
		{
			s: `SELECT * FROM demo GROUP BY COUNTWINDOW(3,1) FILTER( where revenue > 100 )`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.Wildcard{Token: ast.ASTERISK},
						Name:  "",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "demo"}},
				Dimensions: ast.Dimensions{
					ast.Dimension{
						Expr: &ast.Window{
							WindowType: ast.COUNT_WINDOW,
							Length:     &ast.IntegerLiteral{Val: 3},
							Interval:   &ast.IntegerLiteral{Val: 1},
							Filter: &ast.BinaryExpr{
								LHS: &ast.FieldRef{Name: "revenue", StreamName: ast.DefaultStream},
								OP:  ast.GT,
								RHS: &ast.IntegerLiteral{Val: 100},
							},
						},
					},
				},
			},
		},
		{
			s: `SELECT * FROM demo GROUP BY department, COUNTWINDOW(3,1) FILTER( where revenue > 100 ), year`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.Wildcard{Token: ast.ASTERISK},
						Name:  "",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "demo"}},
				Dimensions: ast.Dimensions{
					ast.Dimension{Expr: &ast.FieldRef{Name: "department", StreamName: ast.DefaultStream}},
					ast.Dimension{
						Expr: &ast.Window{
							WindowType: ast.COUNT_WINDOW,
							Length:     &ast.IntegerLiteral{Val: 3},
							Interval:   &ast.IntegerLiteral{Val: 1},
							Filter: &ast.BinaryExpr{
								LHS: &ast.FieldRef{Name: "revenue", StreamName: ast.DefaultStream},
								OP:  ast.GT,
								RHS: &ast.IntegerLiteral{Val: 100},
							},
						},
					},
					ast.Dimension{Expr: &ast.FieldRef{Name: "year", StreamName: ast.DefaultStream}},
				},
			},
		},
		//to be supported
		{
			s:    `SELECT sum(f1) FILTER( where revenue > 100 ) FROM tbl GROUP BY year`,
			stmt: nil,
			err:  "found \"FILTER\", expected FROM.",
		},
		{
			s:    `SELECT * FROM demo GROUP BY COUNTWINDOW(3,1) FILTER where revenue > 100`,
			stmt: nil,
			err:  "Found \"WHERE\" after FILTER, expect parentheses.",
		},
		{
			s:    `SELECT * FROM demo GROUP BY COUNTWINDOW(3,1) where revenue > 100`,
			stmt: nil,
			err:  "found \"WHERE\", expected EOF.",
		},
	}

	fmt.Printf("The test bucket size is %d.\n\n", len(tests))
	for i, tt := range tests {
		//fmt.Printf("Parsing SQL %q.\n", tt.s)
		stmt, err := NewParser(strings.NewReader(tt.s)).Parse()
		if !reflect.DeepEqual(tt.err, testx.Errstring(err)) {
			t.Errorf("%d. %q: error mismatch:\n  exp=%s\n  got=%s\n\n", i, tt.s, tt.err, err)
		} else if tt.err == "" && !reflect.DeepEqual(tt.stmt, stmt) {
			t.Errorf("%d. %q\n\nstmt mismatch:\n\nexp=%#v\n\ngot=%#v\n\n", i, tt.s, tt.stmt, stmt)
		}
	}
}

func TestParser_ParseJsonExpr(t *testing.T) {
	var tests = []struct {
		s    string
		stmt *ast.SelectStatement
		err  string
	}{
		{
			s: `SELECT children[0] FROM demo`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr: &ast.BinaryExpr{
							LHS: &ast.FieldRef{Name: "children", StreamName: ast.DefaultStream},
							OP:  ast.SUBSET,
							RHS: &ast.IndexExpr{Index: &ast.IntegerLiteral{Val: 0}},
						},
						Name:  "",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "demo"}},
			},
		},

		{
			s: `SELECT children[0]->first FROM demo`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr: &ast.BinaryExpr{
							LHS: &ast.BinaryExpr{
								LHS: &ast.FieldRef{Name: "children", StreamName: ast.DefaultStream},
								OP:  ast.SUBSET,
								RHS: &ast.IndexExpr{Index: &ast.IntegerLiteral{Val: 0}},
							},
							OP:  ast.ARROW,
							RHS: &ast.JsonFieldRef{Name: "first"},
						},

						Name:  "",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "demo"}},
			},
		},

		{
			s: `SELECT children->first[2] FROM demo`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr: &ast.BinaryExpr{
							LHS: &ast.BinaryExpr{
								LHS: &ast.FieldRef{Name: "children", StreamName: ast.DefaultStream},
								OP:  ast.ARROW,
								RHS: &ast.JsonFieldRef{Name: "first"},
							},
							OP:  ast.SUBSET,
							RHS: &ast.IndexExpr{Index: &ast.IntegerLiteral{Val: 2}},
						},

						Name:  "",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "demo"}},
			},
		},

		{
			s: `SELECT children->first[2]->test FROM demo`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr: &ast.BinaryExpr{
							LHS: &ast.BinaryExpr{
								LHS: &ast.BinaryExpr{
									LHS: &ast.FieldRef{Name: "children", StreamName: ast.DefaultStream},
									OP:  ast.ARROW,
									RHS: &ast.JsonFieldRef{Name: "first"},
								},
								OP:  ast.SUBSET,
								RHS: &ast.IndexExpr{Index: &ast.IntegerLiteral{Val: 2}},
							},
							OP:  ast.ARROW,
							RHS: &ast.JsonFieldRef{Name: "test"},
						},

						Name:  "",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "demo"}},
			},
		},

		{
			s: `SELECT children[0:1] FROM demo`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr: &ast.BinaryExpr{
							LHS: &ast.FieldRef{Name: "children", StreamName: ast.DefaultStream},
							OP:  ast.SUBSET,
							RHS: &ast.ColonExpr{Start: &ast.IntegerLiteral{Val: 0}, End: &ast.IntegerLiteral{Val: 1}},
						},
						Name:  "",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "demo"}},
			},
		},

		{
			s: `SELECT children[:1] FROM demo`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr: &ast.BinaryExpr{
							LHS: &ast.FieldRef{Name: "children", StreamName: ast.DefaultStream},
							OP:  ast.SUBSET,
							RHS: &ast.ColonExpr{Start: &ast.IntegerLiteral{Val: 0}, End: &ast.IntegerLiteral{Val: 1}},
						},
						Name:  "",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "demo"}},
			},
		},

		{
			s: `SELECT children[:] FROM demo`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr: &ast.BinaryExpr{
							LHS: &ast.FieldRef{Name: "children", StreamName: ast.DefaultStream},
							OP:  ast.SUBSET,
							RHS: &ast.ColonExpr{Start: &ast.IntegerLiteral{Val: 0}, End: &ast.IntegerLiteral{Val: math.MinInt32}},
						},
						Name:  "",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "demo"}},
			},
		},

		{
			s: `SELECT children[2:] AS c FROM demo`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr: &ast.BinaryExpr{
							LHS: &ast.FieldRef{Name: "children", StreamName: ast.DefaultStream},
							OP:  ast.SUBSET,
							RHS: &ast.ColonExpr{Start: &ast.IntegerLiteral{Val: 2}, End: &ast.IntegerLiteral{Val: math.MinInt32}},
						},
						Name:  "",
						AName: "c"},
				},
				Sources: []ast.Source{&ast.Table{Name: "demo"}},
			},
		},

		{
			s: `SELECT children[2:]->first AS c FROM demo`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr: &ast.BinaryExpr{
							LHS: &ast.BinaryExpr{LHS: &ast.FieldRef{Name: "children", StreamName: ast.DefaultStream}, OP: ast.SUBSET, RHS: &ast.ColonExpr{Start: &ast.IntegerLiteral{Val: 2}, End: &ast.IntegerLiteral{Val: math.MinInt32}}},
							OP:  ast.ARROW,
							RHS: &ast.JsonFieldRef{Name: "first"},
						},
						Name:  "",
						AName: "c"},
				},
				Sources: []ast.Source{&ast.Table{Name: "demo"}},
			},
		},

		{
			s: `SELECT demo.* FROM demo`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.FieldRef{StreamName: ast.StreamName("demo"), Name: "*"},
						Name:  "*",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "demo"}},
			},
		},

		{
			s: `SELECT demo.children[2:]->first AS c FROM demo`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr: &ast.BinaryExpr{
							LHS: &ast.BinaryExpr{LHS: &ast.FieldRef{StreamName: ast.StreamName("demo"), Name: "children"}, OP: ast.SUBSET, RHS: &ast.ColonExpr{Start: &ast.IntegerLiteral{Val: 2}, End: &ast.IntegerLiteral{Val: math.MinInt32}}},
							OP:  ast.ARROW,
							RHS: &ast.JsonFieldRef{Name: "first"},
						},
						Name:  "",
						AName: "c"},
				},
				Sources: []ast.Source{&ast.Table{Name: "demo"}},
			},
		},

		{
			s: `SELECT lower(demo.children[2:]->first) AS c FROM demo`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr: &ast.Call{
							Name: "lower",
							Args: []ast.Expr{
								&ast.BinaryExpr{
									LHS: &ast.BinaryExpr{LHS: &ast.FieldRef{StreamName: ast.StreamName("demo"), Name: "children"}, OP: ast.SUBSET, RHS: &ast.ColonExpr{Start: &ast.IntegerLiteral{Val: 2}, End: &ast.IntegerLiteral{Val: math.MinInt32}}},
									OP:  ast.ARROW,
									RHS: &ast.JsonFieldRef{Name: "first"},
								},
							},
						},
						Name:  "lower",
						AName: "c"},
				},
				Sources: []ast.Source{&ast.Table{Name: "demo"}},
			},
		},

		{
			s: `SELECT children[:1] FROM demo WHERE abc[0] > 12`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr: &ast.BinaryExpr{
							LHS: &ast.FieldRef{Name: "children", StreamName: ast.DefaultStream},
							OP:  ast.SUBSET,
							RHS: &ast.ColonExpr{Start: &ast.IntegerLiteral{Val: 0}, End: &ast.IntegerLiteral{Val: 1}},
						},
						Name:  "",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "demo"}},
				Condition: &ast.BinaryExpr{
					LHS: &ast.BinaryExpr{
						LHS: &ast.FieldRef{Name: "abc", StreamName: ast.DefaultStream},
						OP:  ast.SUBSET,
						RHS: &ast.IndexExpr{Index: &ast.IntegerLiteral{Val: 0}},
					},
					OP:  ast.GT,
					RHS: &ast.IntegerLiteral{Val: 12},
				},
			},
		},

		{
			s:    `SELECT demo.children.first AS c FROM demo`,
			stmt: nil,
			err:  "Too many field names. Please use -> to reference keys in struct.\n",
		},
		{
			s: `SELECT children[index] FROM demo`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr: &ast.BinaryExpr{
							LHS: &ast.FieldRef{Name: "children", StreamName: ast.DefaultStream},
							OP:  ast.SUBSET,
							RHS: &ast.IndexExpr{Index: &ast.FieldRef{Name: "index", StreamName: ast.DefaultStream}},
						},
						Name:  "",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "demo"}},
			},
		},
	}

	fmt.Printf("The test bucket size is %d.\n\n", len(tests))
	for i, tt := range tests {
		stmt, err := NewParser(strings.NewReader(tt.s)).Parse()
		if !reflect.DeepEqual(tt.err, testx.Errstring(err)) {
			t.Errorf("%d. %q: error mismatch:\n  exp=%s\n  got=%s\n\n", i, tt.s, tt.err, err)
		} else if tt.err == "" && !reflect.DeepEqual(tt.stmt, stmt) {
			t.Errorf("%d. %q\n\nstmt mismatch:\n\nexp=%#v\n\ngot=%#v\n\n", i, tt.s, tt.stmt, stmt)
		}
	}
}

func TestParser_ParseJoins(t *testing.T) {
	var tests = []struct {
		s    string
		stmt *ast.SelectStatement
		err  string
	}{
		{
			s: `SELECT * FROM topic/sensor1 LEFT JOIN topic1 ON f=k`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.Wildcard{Token: ast.ASTERISK},
						Name:  "",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "topic/sensor1"}},
				Joins: []ast.Join{
					{
						Name: "topic1", Alias: "", JoinType: ast.LEFT_JOIN, Expr: &ast.BinaryExpr{
							LHS: &ast.FieldRef{Name: "f", StreamName: ast.DefaultStream},
							OP:  ast.EQ,
							RHS: &ast.FieldRef{Name: "k", StreamName: ast.DefaultStream},
						},
					},
				},
			},
		},

		{
			s: `SELECT * FROM topic/sensor1 AS t1 INNER JOIN topic1 AS t2 ON f=k`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.Wildcard{Token: ast.ASTERISK},
						Name:  "",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "topic/sensor1", Alias: "t1"}},
				Joins: []ast.Join{
					{
						Name: "topic1", Alias: "t2", JoinType: ast.INNER_JOIN, Expr: &ast.BinaryExpr{
							LHS: &ast.FieldRef{Name: "f", StreamName: ast.DefaultStream},
							OP:  ast.EQ,
							RHS: &ast.FieldRef{Name: "k", StreamName: ast.DefaultStream},
						},
					},
				},
			},
		},

		{
			s: `SELECT * FROM topic/sensor1 AS t1 LEFT JOIN topic1/sensor2 AS t2 ON f=k`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.Wildcard{Token: ast.ASTERISK},
						Name:  "",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "topic/sensor1", Alias: "t1"}},
				Joins: []ast.Join{
					{
						Name: "topic1/sensor2", Alias: "t2", JoinType: ast.LEFT_JOIN, Expr: &ast.BinaryExpr{
							LHS: &ast.FieldRef{Name: "f", StreamName: ast.DefaultStream},
							OP:  ast.EQ,
							RHS: &ast.FieldRef{Name: "k", StreamName: ast.DefaultStream},
						},
					},
				},
			},
		},

		{
			s: `SELECT t1.name FROM topic/sensor1 AS t1 LEFT JOIN topic1/sensor2 AS t2 ON f=k`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.FieldRef{StreamName: ast.StreamName("t1"), Name: "name"},
						Name:  "name",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "topic/sensor1", Alias: "t1"}},
				Joins: []ast.Join{
					{
						Name: "topic1/sensor2", Alias: "t2", JoinType: ast.LEFT_JOIN, Expr: &ast.BinaryExpr{
							LHS: &ast.FieldRef{Name: "f", StreamName: ast.DefaultStream},
							OP:  ast.EQ,
							RHS: &ast.FieldRef{Name: "k", StreamName: ast.DefaultStream},
						},
					},
				},
			},
		},

		{
			s: `SELECT t1.name FROM topic/sensor1 AS t1 LEFT JOIN topic1/sensor2 AS t2 ON t1.f=t2.k`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.FieldRef{StreamName: ast.StreamName("t1"), Name: "name"},
						Name:  "name",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "topic/sensor1", Alias: "t1"}},
				Joins: []ast.Join{
					{
						Name: "topic1/sensor2", Alias: "t2", JoinType: ast.LEFT_JOIN, Expr: &ast.BinaryExpr{
							LHS: &ast.FieldRef{StreamName: ast.StreamName("t1"), Name: "f"},
							OP:  ast.EQ,
							RHS: &ast.FieldRef{StreamName: ast.StreamName("t2"), Name: "k"},
						},
					},
				},
			},
		},

		{
			s: `SELECT t1.name FROM topic/sensor1 AS t1 RIGHT JOIN topic1/sensor2 AS t2 ON t1.f=t2.k`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.FieldRef{StreamName: ast.StreamName("t1"), Name: "name"},
						Name:  "name",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "topic/sensor1", Alias: "t1"}},
				Joins: []ast.Join{
					{
						Name: "topic1/sensor2", Alias: "t2", JoinType: ast.RIGHT_JOIN, Expr: &ast.BinaryExpr{
							LHS: &ast.FieldRef{StreamName: ast.StreamName("t1"), Name: "f"},
							OP:  ast.EQ,
							RHS: &ast.FieldRef{StreamName: ast.StreamName("t2"), Name: "k"},
						},
					},
				},
			},
		},

		{
			s: `SELECT t1.name FROM topic/sensor1 AS t1 FULL JOIN topic1/sensor2 AS t2 ON t1.f=t2.k`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.FieldRef{StreamName: ast.StreamName("t1"), Name: "name"},
						Name:  "name",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "topic/sensor1", Alias: "t1"}},
				Joins: []ast.Join{
					{
						Name: "topic1/sensor2", Alias: "t2", JoinType: ast.FULL_JOIN, Expr: &ast.BinaryExpr{
							LHS: &ast.FieldRef{StreamName: ast.StreamName("t1"), Name: "f"},
							OP:  ast.EQ,
							RHS: &ast.FieldRef{StreamName: ast.StreamName("t2"), Name: "k"},
						},
					},
				},
			},
		},

		{
			s:    `SELECT t1.name FROM topic/sensor1 AS t1 CROSS JOIN topic1/sensor2 AS t2 ON t1.f=t2.k`,
			stmt: nil,
			err:  "On expression is not required for cross join type.\n",
		},

		{
			s: `SELECT t1.name FROM topic/sensor1 AS t1 CROSS JOIN topic1/sensor2 AS t2`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.FieldRef{StreamName: ast.StreamName("t1"), Name: "name"},
						Name:  "name",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "topic/sensor1", Alias: "t1"}},
				Joins: []ast.Join{
					{
						Name: "topic1/sensor2", Alias: "t2", JoinType: ast.CROSS_JOIN, Expr: nil,
					},
				},
			},
		},

		{
			s: `SELECT demo.*, demo2.* FROM demo LEFT JOIN demo2 on demo.f1 = demo2.f2`,
			stmt: &ast.SelectStatement{
				Fields: []ast.Field{
					{
						Expr:  &ast.FieldRef{StreamName: ast.StreamName("demo"), Name: "*"},
						Name:  "*",
						AName: ""},
					{
						Expr:  &ast.FieldRef{StreamName: ast.StreamName("demo2"), Name: "*"},
						Name:  "*",
						AName: ""},
				},
				Sources: []ast.Source{&ast.Table{Name: "demo"}},
				Joins: []ast.Join{
					{
						Name: "demo2", Alias: "", JoinType: ast.LEFT_JOIN, Expr: &ast.BinaryExpr{
							LHS: &ast.FieldRef{StreamName: ast.StreamName("demo"), Name: "f1"},
							OP:  ast.EQ,
							RHS: &ast.FieldRef{StreamName: ast.StreamName("demo2"), Name: "f2"},
						},
					},
				},
			},
		},
	}

	fmt.Printf("The test bucket size is %d.\n\n", len(tests))
	for i, tt := range tests {
		stmt, err := NewParser(strings.NewReader(tt.s)).Parse()
		if !reflect.DeepEqual(tt.err, testx.Errstring(err)) {
			t.Errorf("%d. %q: error mismatch:\n  exp=%s\n  got=%s\n\n", i, tt.s, tt.err, err)
		} else if tt.err == "" && !reflect.DeepEqual(tt.stmt, stmt) {
			t.Errorf("%d. %q\n\nstmt mismatch:\n\nexp=%#v\n\ngot=%#v\n\n", i, tt.s, tt.stmt, stmt)
		}
	}
}

func TestParser_ParseStatements(t *testing.T) {
	var tests = []struct {
		s     string
		stmts []ast.SelectStatement
		err   string
	}{
		{s: "SELECT name FROM tbl;\nSELECT name FROM topic/sensor1\n",
			stmts: []ast.SelectStatement{
				{
					Fields: []ast.Field{
						{
							Expr:  &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream},
							Name:  "name",
							AName: ""},
					},
					Sources: []ast.Source{&ast.Table{Name: "tbl"}},
				},
				{
					Fields: []ast.Field{
						{
							Expr:  &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream},
							Name:  "name",
							AName: ""},
					},
					Sources: []ast.Source{&ast.Table{Name: "topic/sensor1"}},
				},
			},
		},
		{s: "SELECT name FROM tbl;\nSELECT name FROM topic/sensor1;\n--SELECT comment FROM topic/comment",
			stmts: []ast.SelectStatement{
				{
					Fields: []ast.Field{
						{
							Expr:  &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream},
							Name:  "name",
							AName: ""},
					},
					Sources: []ast.Source{&ast.Table{Name: "tbl"}},
				},
				{
					Fields: []ast.Field{
						{
							Expr:  &ast.FieldRef{Name: "name", StreamName: ast.DefaultStream},
							Name:  "name",
							AName: ""},
					},
					Sources: []ast.Source{&ast.Table{Name: "topic/sensor1"}},
				},
			},
		},
	}

	fmt.Printf("The test bucket size is %d.\n\n", len(tests))
	for i, tt := range tests {
		stmts, err := NewParser(strings.NewReader(tt.s)).ParseQueries()
		if !reflect.DeepEqual(tt.err, testx.Errstring(err)) {
			t.Errorf("%d. %q: error mismatch:\n  exp=%s\n  got=%s\n\n", i, tt.s, tt.err, err)
		} else if tt.err == "" && !reflect.DeepEqual(tt.stmts, stmts) {
			t.Errorf("%d. %q\n\nstmt mismatch:\n\nexp=%#v\n\ngot=%#v\n\n", i, tt.s, tt.stmts, stmts)
		}
	}
}
