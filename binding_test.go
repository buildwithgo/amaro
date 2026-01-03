package amaro

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type TestUser struct {
	Name       string   `json:"name" query:"name" form:"name"`
	Age        int      `json:"age" query:"age" form:"age"`
	Admin      bool     `json:"admin" query:"admin" form:"admin"`
	Score      float64  `json:"score" query:"score" form:"score"`
	Tags       []string `json:"tags" query:"tags" form:"tags"`
	Ratings    []int    `json:"ratings" query:"ratings" form:"ratings"`
	PtrField   *int     `json:"ptr_field" query:"ptr_field" form:"ptr_field"`
	ComplexVal complex128 `json:"complex" query:"complex" form:"complex"`
}

func TestBindJSON(t *testing.T) {
	ptrVal := 123
	user := TestUser{
		Name:       "Alice",
		Age:        30,
		Admin:      true,
		Score:      99.5,
		Tags:       []string{"go", "rust"},
		Ratings:    []int{5, 4},
		PtrField:   &ptrVal,
		ComplexVal: 1 + 2i,
	}
	// JSON marshaling of complex numbers is not supported by standard library
	// So we omit it for JSON test or handle it specially if we wanted.
	// We'll skip complex for JSON test as standard json doesn't support it without custom marshaller.
	// But let's test the others.
	user.ComplexVal = 0

	body, _ := json.Marshal(user)

	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	w := httptest.NewRecorder()
	c := NewContext(w, req)

	var boundUser TestUser
	if err := c.BindJSON(&boundUser); err != nil {
		t.Fatalf("BindJSON failed: %v", err)
	}

	if boundUser.Name != user.Name {
		t.Errorf("Expected Name %v, got %v", user.Name, boundUser.Name)
	}
	if boundUser.Score != user.Score {
		t.Errorf("Expected Score %v, got %v", user.Score, boundUser.Score)
	}
	if len(boundUser.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(boundUser.Tags))
	}
}

func TestBindQuery(t *testing.T) {
	// query parameters
	// arrays in query usually: ?tags=go&tags=rust
	q := url.Values{}
	q.Set("name", "Bob")
	q.Set("age", "25")
	q.Set("admin", "false")
	q.Set("score", "88.8")
	q.Add("tags", "one")
	q.Add("tags", "two")
	q.Add("ratings", "10")
	q.Add("ratings", "20")
	q.Set("ptr_field", "456")
	q.Set("complex", "1+2i")

	req := httptest.NewRequest("GET", "/?"+q.Encode(), nil)
	w := httptest.NewRecorder()
	c := NewContext(w, req)

	var boundUser TestUser
	if err := c.BindQuery(&boundUser); err != nil {
		t.Fatalf("BindQuery failed: %v", err)
	}

	if boundUser.Name != "Bob" {
		t.Errorf("Expected Name Bob, got %s", boundUser.Name)
	}
	if boundUser.Score != 88.8 {
		t.Errorf("Expected Score 88.8, got %f", boundUser.Score)
	}
	if len(boundUser.Tags) != 2 || boundUser.Tags[0] != "one" {
		t.Errorf("Expected Tags [one, two], got %v", boundUser.Tags)
	}
	if len(boundUser.Ratings) != 2 || boundUser.Ratings[0] != 10 {
		t.Errorf("Expected Ratings [10, 20], got %v", boundUser.Ratings)
	}
	if boundUser.PtrField == nil || *boundUser.PtrField != 456 {
		t.Errorf("Expected PtrField 456, got %v", boundUser.PtrField)
	}
	if boundUser.ComplexVal != 1+2i {
		t.Errorf("Expected Complex 1+2i, got %v", boundUser.ComplexVal)
	}
}

func TestBindForm(t *testing.T) {
	form := url.Values{}
	form.Set("name", "Charlie")
	form.Set("age", "40")
	form.Set("admin", "true")
	form.Set("score", "12.34")
	form.Add("tags", "alpha")
	form.Add("tags", "beta")
	form.Add("ratings", "100")
	form.Set("ptr_field", "789")
	form.Set("complex", "3+4i")

	req := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	c := NewContext(w, req)

	var boundUser TestUser
	if err := c.BindForm(&boundUser); err != nil {
		t.Fatalf("BindForm failed: %v", err)
	}

	if boundUser.Name != "Charlie" {
		t.Errorf("Expected Name Charlie, got %s", boundUser.Name)
	}
	if boundUser.Score != 12.34 {
		t.Errorf("Expected Score 12.34, got %f", boundUser.Score)
	}
	if len(boundUser.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %v", boundUser.Tags)
	}
	if len(boundUser.Ratings) != 1 || boundUser.Ratings[0] != 100 {
		t.Errorf("Expected Ratings [100], got %v", boundUser.Ratings)
	}
	if boundUser.PtrField == nil || *boundUser.PtrField != 789 {
		t.Errorf("Expected PtrField 789, got %v", boundUser.PtrField)
	}
	if boundUser.ComplexVal != 3+4i {
		t.Errorf("Expected Complex 3+4i, got %v", boundUser.ComplexVal)
	}
}
