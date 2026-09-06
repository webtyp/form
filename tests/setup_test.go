package form_test

import "webtyp.com/model"

import (
	"webtyp.com/fmt"
	"webtyp.com/form"
	"webtyp.com/input"
)

// testIDGen is a deterministic test double for model.IDGenerator — form must
// never construct its own generator, so every test passes one explicitly.
type testIDGen struct{ n int }

func (g *testIDGen) NewID() string {
	g.n++
	return "test-id-" + fmt.Convert(g.n).String()
}

// User is a sample struct for testing data binding.
type User struct {
	Name     string
	Email    string
	Password string
	Gender   string
	Role     string
	Address  string
}

func (u *User) Schema() []model.Field {
	return []model.Field{
		{Name: "Name", Type: input.Text(), NotNull: true},
		{Name: "Email", Type: input.Email(), NotNull: true},
		{Name: "Password", Type: input.Password(), NotNull: true},
		{Name: "Gender", Type: input.Gender()},
		{Name: "Role", Type: input.Text()},
		{Name: "Address", Type: input.Address()},
	}
}

func (u *User) Values() []any {
	return []any{u.Name, u.Email, u.Password, u.Gender, u.Role, u.Address}
}

func (u *User) Pointers() []any {
	return []any{&u.Name, &u.Email, &u.Password, &u.Gender, &u.Role, &u.Address}
}

func (u *User) FormName() string {
	return "user"
}

// createTestForm is a helper to create a form for testing.
func createTestForm() *form.Form {
	u := &User{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "secretpassword",
		Gender:   "m",
		Role:     "admin",
		Address:  "123 Main St",
	}
	f, err := form.New("test-parent", u, &testIDGen{})
	if err != nil {
		panic(err)
	}
	return f
}
