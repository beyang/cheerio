package cheerio

import (
	"github.com/kr/pretty"
	"reflect"
	"testing"
)

func TestParseRequirements(t *testing.T) {
	expReqs := []*Requirement{
		{
			Name:       "dep1",
			Constraint: "==",
			Version:    "2.3.2",
		},
		{
			Name:       "dep2",
			Constraint: ">=",
			Version:    "1.0",
		},
		{
			Name:       "dep3",
			Constraint: "",
			Version:    "",
		},
		{
			Name:       "dep4",
			Constraint: "",
			Version:    "",
		},
		{
			Name:       "dep5",
			Constraint: "==",
			Version:    "2.3.2",
		},
		{
			Name:       "dep6",
			Constraint: ">=",
			Version:    "7",
		},
		{
			Name:       "dep7",
			Constraint: "==",
			Version:    "10",
		},
		{
			Name:       "dep8.subdep",
			Constraint: "==",
			Version:    "1.2.3",
		},
		{
			Name:       "dep9",
			Constraint: ">",
			Version:    "1",
		},
		{
			Name:       "dep9",
			Constraint: ">",
			Version:    "1",
		},
		{
			Name:       "dep10",
			Constraint: "==",
			Version:    "1",
		},
		{
			Name:       "dep10",
			Constraint: "",
			Version:    "",
		},
	}
	reqs, err := ParseRequirements(`dep1==2.3.2
dep2>=1.0
dep3
          dep4
 dep5 == 2.3.2
dep6>= 7

[this-is-a-heading]

dep7 ==10
dep8.subdep==1.2.3
dep9>1
dep9 > 1
dep10[extradep]==1
dep10[extradep]
`)

	if err != nil {
		t.Errorf("Error parsing requirements: %s", err)
	} else if !reflect.DeepEqual(reqs, expReqs) {
		t.Errorf("Requirements do not match: %v", pretty.Diff(reqs, expReqs))
	}
}
