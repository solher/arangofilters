package filters

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fp = newFilterProcessor("")

var offsetFilter = &Filter{
	Offset: 1,
}

var limitFilter = &Filter{
	Limit: 2,
}

var offsetLimitFilter = &Filter{
	Offset: 3,
	Limit:  4,
}

var sortFilter = &Filter{
	Sort: []string{"firstName ASC", "lastName dESc", "age"},
}

// integers are converted to float64 because that is what the json unmarshaller do
var basicWhereFilter = &Filter{
	Where: []map[string]interface{}{{
		"password":     "qwertyuiop",
		"age":          float64(22),
		"money":        3000.55,
		"awesome":      true,
		"notAwesome":   false,
		"graduated":    []interface{}{float64(2010), float64(2015)},
		"avg":          []interface{}{15.5, 13.24},
		"birthPlace":   []interface{}{"Chalon", "Macon"},
		"bools":        []interface{}{true, false},
		"strWithQuote": "O'Hare",
	}},
}

var orWhereFilter = &Filter{
	Where: []map[string]interface{}{{
		"oR": []interface{}{
			map[string]interface{}{"lastName": map[string]interface{}{"eq": "O'Connor"}},
			map[string]interface{}{"age": map[string]interface{}{"gt": float64(23)}},
			map[string]interface{}{"age": map[string]interface{}{"lt": float64(26)}},
		}},
	},
}

var andWhereFilter = &Filter{
	Where: []map[string]interface{}{{
		"and": []interface{}{
			map[string]interface{}{"firstName": map[string]interface{}{"neq": "Toto"}},
			map[string]interface{}{"money": 200.5},
		}},
	},
}

var notWhereFilter = &Filter{
	Where: []map[string]interface{}{
		{"not": map[string]interface{}{"firstName": "D'Arcy"}},
		{"nOt": map[string]interface{}{
			"or": []interface{}{
				map[string]interface{}{"lastName": "Herfray"},
				map[string]interface{}{"money": map[string]interface{}{"gte": float64(0)}},
				map[string]interface{}{"money": map[string]interface{}{"lte": 1000.5}}},
		}},
	},
}

var likeWhereFilter = &Filter{
	Where: []map[string]interface{}{
		{"like": map[string]interface{}{
			"text":             "firstName",
			"search":           "fab%",
			"case_insensitive": true,
		}},
		{"like": map[string]interface{}{
			"text":             "lastName",
			"search":           "Her%",
			"case_insensitive": false,
		}},
	},
}

func newAssertRequire(t *testing.T) (*assert.Assertions, *require.Assertions) {
	a := assert.New(t)
	r := require.New(t)
	return a, r
}

// Offset and limit filters
func TestProcessOffsetFilter(t *testing.T) {
	a, r := newAssertRequire(t)
	p, err := fp.Process(offsetFilter)
	r.NoError(err)
	a.Equal("1", p.OffsetLimit)
}

func TestProcessLimitFilter(t *testing.T) {
	a, r := newAssertRequire(t)
	p, err := fp.Process(limitFilter)
	r.NoError(err)
	a.Equal("2", p.OffsetLimit)
}

func TestProcessOffsetLimitFilter(t *testing.T) {
	a, r := newAssertRequire(t)
	p, err := fp.Process(offsetLimitFilter)
	r.NoError(err)
	a.Equal("3, 4", p.OffsetLimit)
}

func TestProcessNegativeOffsetFilter(t *testing.T) {
	a, r := newAssertRequire(t)
	p, err := fp.Process(&Filter{Offset: -1})
	r.NoError(err)
	a.NotNil(p)
}

func TestProcessNegativeLimitFilter(t *testing.T) {
	a, r := newAssertRequire(t)
	p, err := fp.Process(&Filter{Limit: -1})
	r.NoError(err)
	a.NotNil(p)
}

func TestProcessSortFilter(t *testing.T) {
	a, r := newAssertRequire(t)
	p, err := fp.Process(sortFilter)
	r.NoError(err)
	a.Equal("var.firstName ASC, var.lastName DESC, var.age ASC", p.Sort)
}

func TestProcessEmptySortFilter(t *testing.T) {
	a, r := newAssertRequire(t)
	p, err := fp.Process(&Filter{Sort: []string{}})
	r.NoError(err)
	a.Equal("", p.Sort)
}

func TestProcessInvalidCharacterSortFilter(t *testing.T) {
	a, r := newAssertRequire(t)
	p, err := fp.Process(&Filter{Sort: []string{"foo, bar"}})
	r.Error(err)
	a.Nil(p)
}

func TestProcessInvalidUsingOperatorSortFilter(t *testing.T) {
	a, r := newAssertRequire(t)
	p, err := fp.Process(&Filter{Sort: []string{"INSeRT ASC"}})
	r.Error(err)
	a.Nil(p)
}

func TestProcessBasicWhereFilter(t *testing.T) {
	a, r := newAssertRequire(t)
	p, err := fp.Process(basicWhereFilter)
	r.NoError(err)
	split := strings.Split(p.Where, " && ")
	a.Equal(10, len(split))
	expected := []string{
		`var.awesome == true`,
		`var.graduated IN [2010, 2015]`,
		`var.avg IN [15.5, 13.24]`,
		`var.birthPlace IN ['Chalon', 'Macon']`,
		`var.password == 'qwertyuiop'`,
		`var.age == 22`,
		`var.money == 3000.55`,
		`var.notAwesome == false`,
		`var.bools IN [true, false]`,
		`var.strWithQuote == 'O\'Hare'`,
	}
	for _, s := range split {
		a.Contains(expected, s)
	}
}

func TestProcessOrWhereFilter(t *testing.T) {
	a, r := newAssertRequire(t)
	p, err := fp.Process(orWhereFilter)
	r.NoError(err)
	a.Equal(`(var.lastName == 'O\'Connor' || var.age > 23 || var.age < 26)`, p.Where)
}

func TestProcessAndWhereFilter(t *testing.T) {
	a, r := newAssertRequire(t)
	p, err := fp.Process(andWhereFilter)
	r.NoError(err)
	a.Equal(`(var.firstName != 'Toto' && var.money == 200.5)`, p.Where)
}

func TestProcessNotWhereFilter(t *testing.T) {
	a, r := newAssertRequire(t)
	p, err := fp.Process(notWhereFilter)
	r.NoError(err)
	split := strings.Split(p.Where, " && ")
	a.Equal(2, len(split))
	expected := []string{
		`!(var.firstName == 'D\'Arcy')`,
		`!((var.lastName == 'Herfray' || var.money >= 0 || var.money <= 1000.5))`,
	}
	for _, s := range split {
		a.Contains(expected, s)
	}
}

func TestProcessLikeWhereFilter(t *testing.T) {
	a, r := newAssertRequire(t)
	p, err := fp.Process(likeWhereFilter)
	r.NoError(err)
	split := strings.Split(p.Where, " && ")
	expected := []string{
		`LIKE(var.firstName, 'fab%', true)`,
		`LIKE(var.lastName, 'Her%')`,
	}
	for _, s := range split {
		a.Contains(expected, s)
	}
}

func TestProcessInvalidSimpleConditionTypeFilter(t *testing.T) {
	a, r := newAssertRequire(t)
	p, err := fp.Process(&Filter{Where: []map[string]interface{}{{"var.firstName": []interface{}{"foo", map[string]interface{}{"foo": "bar"}}}}})
	r.Error(err)
	a.Nil(p)
}

func TestProcessInvalidAndMapConditionTypeFilter(t *testing.T) {
	a, r := newAssertRequire(t)
	p, err := fp.Process(&Filter{Where: []map[string]interface{}{{"and": []interface{}{"foo", "bar"}}}})
	r.Error(err)
	a.Nil(p)
}

func TestProcessInvalidOrMapConditionTypeFilter(t *testing.T) {
	a, r := newAssertRequire(t)
	p, err := fp.Process(&Filter{Where: []map[string]interface{}{{"or": []interface{}{"foo", "bar"}}}})
	r.Error(err)
	a.Nil(p)
}

func TestProcessInvalidNonArrayConditionTypeFilter(t *testing.T) {
	a, r := newAssertRequire(t)
	p, err := fp.Process(&Filter{Where: []map[string]interface{}{{"and": map[string]interface{}{"var.firstName": "Fabien", "foo": "bar"}}}})
	r.Error(err)
	a.Nil(p)
}

func TestProcessInvalidArrayConditionFilter(t *testing.T) {
	a, r := newAssertRequire(t)
	p, err := fp.Process(&Filter{Where: []map[string]interface{}{{"and": []interface{}{"INSeRT"}}}})
	r.Error(err)
	a.Nil(p)
}

func TestProcessInvalidEqIntFilter(t *testing.T) {
	a, r := newAssertRequire(t)
	p, err := fp.Process(&Filter{Where: []map[string]interface{}{{"money": map[string]interface{}{"eq": 1}}}})
	r.Error(err)
	a.Nil(p)
}

func TestProcessInvalidNeqIntFilter(t *testing.T) {
	a, r := newAssertRequire(t)
	p, err := fp.Process(&Filter{Where: []map[string]interface{}{{"money": map[string]interface{}{"neq": 1}}}})
	r.Error(err)
	a.Nil(p)
}

func TestProcessInvalidGtIntFilter(t *testing.T) {
	a, r := newAssertRequire(t)
	p, err := fp.Process(&Filter{Where: []map[string]interface{}{{"money": map[string]interface{}{"gt": 1}}}})
	r.Error(err)
	a.Nil(p)
}

func TestProcessInvalidGteIntFilter(t *testing.T) {
	a, r := newAssertRequire(t)
	p, err := fp.Process(&Filter{Where: []map[string]interface{}{{"money": map[string]interface{}{"gte": 1}}}})
	r.Error(err)
	a.Nil(p)
}

func TestProcessInvalidLtIntFilter(t *testing.T) {
	a, r := newAssertRequire(t)
	p, err := fp.Process(&Filter{Where: []map[string]interface{}{{"money": map[string]interface{}{"lt": 1}}}})
	r.Error(err)
	a.Nil(p)
}

func TestProcessInvalidLteIntFilter(t *testing.T) {
	a, r := newAssertRequire(t)
	p, err := fp.Process(&Filter{Where: []map[string]interface{}{{"money": map[string]interface{}{"lte": 1}}}})
	r.Error(err)
	a.Nil(p)
}

func TestProcessInvalidNotIntFilter(t *testing.T) {
	a, r := newAssertRequire(t)
	p, err := fp.Process(&Filter{Where: []map[string]interface{}{{"not": 1}}})
	r.Error(err)
	a.Nil(p)
}

func TestProcessNilFilter(t *testing.T) {
	_, r := newAssertRequire(t)
	_, err := fp.Process(nil)
	r.NoError(err)
}

func TestEscapeString(t *testing.T) {
	a, _ := newAssertRequire(t)
	s := escapeString("O'Hare")
	a.Equal("O\\'Hare", s)
}
