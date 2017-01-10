package runtime

import (
	"fmt"
	"reflect"
	"testing"
)

func TestNewCronSchedule(t *testing.T) {
	testData := []struct {
		CaseNo       int
		Expr         string
		HasError     bool
		ExpectedByte []byte
	}{
		{
			CaseNo:       1,
			Expr:         "0 1 2 3 4",
			HasError:     false,
			ExpectedByte: []byte{48, 0, 49, 0, 50, 0, 51, 0, 52, 0},
		},
		{
			CaseNo:       2,
			Expr:         "* 1 2 3 4",
			HasError:     false,
			ExpectedByte: []byte{0, 49, 0, 50, 0, 51, 0, 52, 0},
		},
		{
			CaseNo:       3,
			Expr:         "* * * * *",
			HasError:     false,
			ExpectedByte: []byte{0, 0, 0, 0, 0},
		},
		{
			CaseNo:       4,
			Expr:         "* * * 0 *",
			HasError:     true,
			ExpectedByte: []byte{0, 0, 0, 0, 0},
		},
		{
			CaseNo:       5,
			Expr:         "* * * * 0",
			HasError:     false,
			ExpectedByte: []byte{0, 0, 0, 0, 48, 0},
		},
		{
			CaseNo:       6,
			Expr:         "* * * - *",
			HasError:     true,
			ExpectedByte: []byte{48, 0, 48, 0, 49, 0, 49, 0, 48, 0},
		},
	}
	for _, td := range testData {
		gotObj, gotErr := NewCronSchedule(td.Expr)

		if td.HasError != (gotErr != nil) {
			t.Fatalf("case %v: expected error %v but got %v\n", td.CaseNo, td.HasError, gotErr != nil)
		}
		if td.HasError {
			continue
		}
		if !reflect.DeepEqual(gotObj.Bytes(), td.ExpectedByte) {
			t.Fatalf("case %v: expected `%v` got `%v`\n", td.CaseNo, td.ExpectedByte, gotObj.Bytes())
		}
	}

	_, err := NewCronSchedule("@every 1h")
	fmt.Print(err)
	if err == nil {
		t.Fatalf("expected err but got `<nil>`\n")
	}
}
