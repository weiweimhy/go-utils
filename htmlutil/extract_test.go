package htmlutil

import (
	"reflect"
	"testing"
)

func TestExtractTextByTagWithAttrRequiresAttributePresence(t *testing.T) {
	htmlContent := `<div data-state="">empty</div><div>missing</div><div data-state="ready">ready</div>`

	got, err := ExtractTextByTagWithAttr(htmlContent, "div", "data-state", "")
	if err != nil {
		t.Fatalf("ExtractTextByTagWithAttr() error = %v", err)
	}
	if want := []string{"empty", "ready"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("ExtractTextByTagWithAttr() = %#v, want %#v", got, want)
	}
}

func TestExtractTextByTagWithAttrMatchesExactValue(t *testing.T) {
	htmlContent := `<p role="note">first</p><p role="alert">second</p>`

	got, err := ExtractTextByTagWithAttr(htmlContent, "p", "role", "alert")
	if err != nil {
		t.Fatalf("ExtractTextByTagWithAttr() error = %v", err)
	}
	if want := []string{"second"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("ExtractTextByTagWithAttr() = %#v, want %#v", got, want)
	}
}
