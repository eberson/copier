package copier_test

import (
	"errors"

	"github.com/stretchr/testify/assert"

	"github.com/eberson/copier"

	"reflect"
	"testing"
	"time"
)

type User struct {
	Name     string
	Birthday *time.Time
	Nickname string
	Role     string
	Age      int32
	FakeAge  *int32
	Notes    []string
	flags    []byte
}

func (user User) DoubleAge() int32 {
	return 2 * user.Age
}

type Employee struct {
	Name      string
	Birthday  *time.Time
	Nickname  *string
	Age       int64
	FakeAge   int
	EmployeID int64
	DoubleAge int32
	SuperRule string
	Notes     []string
	flags     []byte
}

func (employee *Employee) Role(role string) {
	employee.SuperRule = "Super " + role
}

func checkEmployee(employee Employee, user User, t *testing.T, testCase string) {
	if employee.Name != user.Name {
		t.Errorf("%v: Name haven't been copied correctly.", testCase)
	}
	if employee.Nickname == nil || *employee.Nickname != user.Nickname {
		t.Errorf("%v: NickName haven't been copied correctly.", testCase)
	}
	if employee.Birthday == nil && user.Birthday != nil {
		t.Errorf("%v: Birthday haven't been copied correctly.", testCase)
	}
	if employee.Birthday != nil && user.Birthday == nil {
		t.Errorf("%v: Birthday haven't been copied correctly.", testCase)
	}
	if employee.Birthday != nil && user.Birthday != nil &&
		!employee.Birthday.Equal(*(user.Birthday)) {
		t.Errorf("%v: Birthday haven't been copied correctly.", testCase)
	}
	if employee.Age != int64(user.Age) {
		t.Errorf("%v: Age haven't been copied correctly.", testCase)
	}
	if user.FakeAge != nil && employee.FakeAge != int(*user.FakeAge) {
		t.Errorf("%v: FakeAge haven't been copied correctly.", testCase)
	}
	if employee.DoubleAge != user.DoubleAge() {
		t.Errorf("%v: Copy from method doesn't work", testCase)
	}
	if employee.SuperRule != "Super "+user.Role {
		t.Errorf("%v: Copy to method doesn't work", testCase)
	}
	if !reflect.DeepEqual(employee.Notes, user.Notes) {
		t.Errorf("%v: Copy from slice doeork", testCase)
	}
}

func TestCopySameStructWithPointerField(t *testing.T) {
	var fakeAge int32 = 12
	var currentTime = time.Now()
	user := &User{Birthday: &currentTime, Name: "Jinzhu", Nickname: "jinzhu", Age: 18, FakeAge: &fakeAge, Role: "Admin", Notes: []string{"hello world", "welcome"}, flags: []byte{'x'}}
	newUser := &User{}
	copier.Copy(newUser, user)
	if user.Birthday == newUser.Birthday {
		t.Errorf("TestCopySameStructWithPointerField: copy Birthday failed since they need to have different address")
	}

	if user.FakeAge == newUser.FakeAge {
		t.Errorf("TestCopySameStructWithPointerField: copy FakeAge failed since they need to have different address")
	}
}

func TestCopyStruct(t *testing.T) {
	var fakeAge int32 = 12
	user := User{Name: "Jinzhu", Nickname: "jinzhu", Age: 18, FakeAge: &fakeAge, Role: "Admin", Notes: []string{"hello world", "welcome"}, flags: []byte{'x'}}
	employee := Employee{}

	if err := copier.Copy(employee, &user); err == nil {
		t.Errorf("Copy to unaddressable value should get error")
	}

	copier.Copy(&employee, &user)
	checkEmployee(employee, user, t, "Copy From Ptr To Ptr")

	employee2 := Employee{}
	copier.Copy(&employee2, user)
	checkEmployee(employee2, user, t, "Copy From Struct To Ptr")

	employee3 := Employee{}
	ptrToUser := &user
	copier.Copy(&employee3, &ptrToUser)
	checkEmployee(employee3, user, t, "Copy From Double Ptr To Ptr")

	employee4 := &Employee{}
	copier.Copy(&employee4, user)
	checkEmployee(*employee4, user, t, "Copy From Ptr To Double Ptr")
}

func TestCopyFromStructToSlice(t *testing.T) {
	user := User{Name: "Jinzhu", Nickname: "jinzhu", Age: 18, Role: "Admin", Notes: []string{"hello world"}}
	employees := []Employee{}

	if err := copier.Copy(employees, &user); err != nil && len(employees) != 0 {
		t.Errorf("Copy to unaddressable value should get error")
	}

	if copier.Copy(&employees, &user); len(employees) != 1 {
		t.Errorf("Should only have one elem when copy struct to slice")
	} else {
		checkEmployee(employees[0], user, t, "Copy From Struct To Slice Ptr")
	}

	employees2 := &[]Employee{}
	if copier.Copy(&employees2, user); len(*employees2) != 1 {
		t.Errorf("Should only have one elem when copy struct to slice")
	} else {
		checkEmployee((*employees2)[0], user, t, "Copy From Struct To Double Slice Ptr")
	}

	employees3 := []*Employee{}
	if copier.Copy(&employees3, user); len(employees3) != 1 {
		t.Errorf("Should only have one elem when copy struct to slice")
	} else {
		checkEmployee(*(employees3[0]), user, t, "Copy From Struct To Ptr Slice Ptr")
	}

	employees4 := &[]*Employee{}
	if copier.Copy(&employees4, user); len(*employees4) != 1 {
		t.Errorf("Should only have one elem when copy struct to slice")
	} else {
		checkEmployee(*((*employees4)[0]), user, t, "Copy From Struct To Double Ptr Slice Ptr")
	}
}

func TestCopyFromSliceToSlice(t *testing.T) {
	users := []User{User{Name: "Jinzhu", Nickname: "jinzhu", Age: 18, Role: "Admin", Notes: []string{"hello world"}}, User{Name: "Jinzhu2", Nickname: "jinzhu", Age: 22, Role: "Dev", Notes: []string{"hello world", "hello"}}}
	employees := []Employee{}

	if copier.Copy(&employees, users); len(employees) != 2 {
		t.Errorf("Should have two elems when copy slice to slice")
	} else {
		checkEmployee(employees[0], users[0], t, "Copy From Slice To Slice Ptr @ 1")
		checkEmployee(employees[1], users[1], t, "Copy From Slice To Slice Ptr @ 2")
	}

	employees2 := &[]Employee{}
	if copier.Copy(&employees2, &users); len(*employees2) != 2 {
		t.Errorf("Should have two elems when copy slice to slice")
	} else {
		checkEmployee((*employees2)[0], users[0], t, "Copy From Slice Ptr To Double Slice Ptr @ 1")
		checkEmployee((*employees2)[1], users[1], t, "Copy From Slice Ptr To Double Slice Ptr @ 2")
	}

	employees3 := []*Employee{}
	if copier.Copy(&employees3, users); len(employees3) != 2 {
		t.Errorf("Should have two elems when copy slice to slice")
	} else {
		checkEmployee(*(employees3[0]), users[0], t, "Copy From Slice To Ptr Slice Ptr @ 1")
		checkEmployee(*(employees3[1]), users[1], t, "Copy From Slice To Ptr Slice Ptr @ 2")
	}

	employees4 := &[]*Employee{}
	if copier.Copy(&employees4, users); len(*employees4) != 2 {
		t.Errorf("Should have two elems when copy slice to slice")
	} else {
		checkEmployee(*((*employees4)[0]), users[0], t, "Copy From Slice Ptr To Double Ptr Slice Ptr @ 1")
		checkEmployee(*((*employees4)[1]), users[1], t, "Copy From Slice Ptr To Double Ptr Slice Ptr @ 2")
	}
}

func TestEmbedded(t *testing.T) {
	type Base struct {
		BaseField1 int
		BaseField2 int
	}

	type Embed struct {
		EmbedField1 int
		EmbedField2 int
		Base
	}

	base := Base{}
	embeded := Embed{}
	embeded.BaseField1 = 1
	embeded.BaseField2 = 2
	embeded.EmbedField1 = 3
	embeded.EmbedField2 = 4

	copier.Copy(&base, &embeded)

	if base.BaseField1 != 1 {
		t.Error("Embedded fields not copied")
	}
}

type structSameName1 struct {
	A string
	B int64
	C time.Time
}

type structSameName2 struct {
	A string
	B time.Time
	C int64
}

func TestCopyFieldsWithSameNameButDifferentTypes(t *testing.T) {
	obj1 := structSameName1{A: "123", B: 2, C: time.Now()}
	obj2 := &structSameName2{}
	err := copier.Copy(obj2, &obj1)
	if err != nil {
		t.Error("Should not raise error")
	}

	if obj2.A != obj1.A {
		t.Errorf("Field A should be copied")
	}
}

type ScannerValue struct {
	V int
}

func (s *ScannerValue) Scan(src interface{}) error {
	return errors.New("I failed")
}

type ScannerStruct struct {
	V *ScannerValue
}

type ScannerStructTo struct {
	V *ScannerValue
}

func TestScanner(t *testing.T) {
	s := &ScannerStruct{
		V: &ScannerValue{
			V: 12,
		},
	}

	s2 := &ScannerStructTo{}

	err := copier.Copy(s2, s)
	if err != nil {
		t.Error("Should not raise error")
	}

	if s.V.V != s2.V.V {
		t.Errorf("Field V should be copied")
	}
}

func TestMapCopyValues(t *testing.T) {
	type MyStruct struct {
		Name  string
		Items map[string]int
	}

	source := &MyStruct{
		Name: "something",
		Items: map[string]int{
			"a": 0,
			"b": 1,
			"c": 2,
		},
	}

	target := &MyStruct{}

	copier.Copy(target, source)

	delete(target.Items, "b")

	assert := assert.New(t)

	assert.Equal(2, len(target.Items))
	assert.Equal(3, len(source.Items))

}

func TestMapCopyComplexValues(t *testing.T) {
	assert := assert.New(t)

	type Student struct {
		Name string
		Age  int
	}

	type Class struct {
		Name     string
		Students map[string]Student
	}

	source := &Class{
		Name: "something",
		Students: map[string]Student{
			"a": {
				Name: "Student A",
				Age:  15,
			},
			"b": {
				Name: "Student B",
				Age:  16,
			},
		},
	}

	target := &Class{}

	copier.Copy(target, source)

	student := target.Students["a"]
	student.Name = "Student X"
	target.Students["a"] = student

	assert.NotEqual(source.Students["a"].Name, target.Students["a"].Name)

	delete(target.Students, "b")

	assert.Equal(1, len(target.Students))
	assert.Equal(2, len(source.Students))

}

func TestCopyNestedSlices(t *testing.T) {
	type MyStruct struct {
		Name  string
		Items []int
	}

	source := &MyStruct{
		Name:  "something",
		Items: []int{0, 1},
	}

	target := &MyStruct{}

	copier.Copy(target, source)

	target.Items[0] = 10
	target.Items[1] = 11

	assert := assert.New(t)

	assert.NotEqual(source.Items[0], target.Items[0])
	assert.NotEqual(source.Items[1], target.Items[1])
}

func TestMismatchedStructToSimple(t *testing.T) {
	// Types don't match.  Nothing should be copied to the mismatched field, but it should not panic.
	type From struct {
		Tmp  time.Time
		Safe int64
	}

	type To struct {
		Tmp  string
		Safe int64
	}

	from := From{}
	from.Tmp = time.Now()
	from.Safe = 1
	to := To{}
	copier.Copy(&to, &from)

	if to.Tmp != "" {
		t.Error("Simple string field populated from complex data type incorrectly.")
	}
	if to.Safe != 1 {
		t.Error("Simple integer field did not copy correctly.")
	}
}

type A struct {
	S       *string
	Control string
}

type B struct {
	S       string
	Control string
}

type C struct {
	S       *string
	Control string
}

func TestEmptyValue(t *testing.T) {
	b := &B{"", "foo"}
	a := &A{}
	copier.Copy(a, b)

	if a.Control != "foo" {
		t.Error("Incorrectly copied string")
	} else if a.S != nil {
		t.Error("Copied empty value field to pointer")
	}
}

func TestNilPtoNiLP(t *testing.T) {
	a := &A{}
	c := &C{}
	copier.Copy(a, c)
	if a.S != nil {
		t.Error("Did not copy nil pointer correctly")
	}
}

func TestNilMap(t *testing.T) {
	type Z struct {
		M map[string]int
		S string
	}
	type Y struct {
		M map[string]int
		S string
	}
	z := &Z{nil, "foo"}
	y := &Y{}

	copier.Copy(y, z)
	if y.S != "foo" {
		t.Error("Unable to copy string")
	} else if y.M != nil {
		t.Error("Uncorrectly copied zero value of map")
	}

}

func TestNilSlice(t *testing.T) {
	type Z struct {
		Slice []string
		S     string
	}
	type Y struct {
		Slice []string
		S     string
	}
	z := &Z{nil, "foo"}
	y := &Y{}

	copier.Copy(y, z)
	if y.S != "foo" {
		t.Error("Unable to copy string")
	} else if y.Slice != nil {
		t.Error("Uncorrectly copied zero value of slice")
	}
}
