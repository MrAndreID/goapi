package user

import (
	"testing"

	"github.com/MrAndreID/goapi/v2/internal/entity"
	"github.com/google/uuid"
)

func TestCreateDataValidate(t *testing.T) {
	valid := CreateData{Name: "Andre", Emails: []string{"andre@gmail.com"}}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid create data to pass validation, got %v", err)
	}

	invalid1 := CreateData{Name: "", Emails: []string{""}}
	if err := invalid1.Validate(); err == nil {
		t.Fatal("expected invalid create data to fail validation")
	}

	invalid2 := CreateData{Name: "Andre [ID]", Emails: []string{"not-an-email"}}
	if err := invalid2.Validate(); err == nil {
		t.Fatal("expected invalid create data to fail validation")
	}
}

func TestReadDataValidate(t *testing.T) {
	valid := ReadData{PaginatorRequest: entity.PaginatorRequest{Page: "1", Limit: "10", OrderBy: "name", SortBy: "asc", Search: "andre", DisableCalculateTotal: "false"}, ID: uuid.NewString()}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid read data to pass validation, got %v", err)
	}

	invalid := ReadData{PaginatorRequest: entity.PaginatorRequest{Page: "x", Limit: "y", OrderBy: "unknown", SortBy: "side", Search: "bad'", DisableCalculateTotal: "maybe"}, ID: "invalid-uuid"}
	if err := invalid.Validate(); err == nil {
		t.Fatal("expected invalid read data to fail validation")
	}
}

func TestUpdateDataValidate(t *testing.T) {
	valid := UpdateData{ID: uuid.NewString(), Name: "Andre", Emails: []string{"andre@gmail.com"}}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid update data to pass validation, got %v", err)
	}

	invalid := UpdateData{ID: "invalid-id", Name: "bad\"", Emails: []string{"not-an-email"}}
	if err := invalid.Validate(); err == nil {
		t.Fatal("expected invalid update data to fail validation")
	}
}

func TestDeleteDataValidate(t *testing.T) {
	valid := DeleteData{ID: uuid.NewString()}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid delete data to pass validation, got %v", err)
	}

	invalid := DeleteData{ID: "invalid-id"}
	if err := invalid.Validate(); err == nil {
		t.Fatal("expected invalid delete data to fail validation")
	}
}
