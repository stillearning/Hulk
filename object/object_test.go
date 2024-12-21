package object

import (
	"testing"
)

func TestHashKeysString(t *testing.T) {
	hello1 := &String{Value: "Hello World"}
	hello2 := &String{Value: "Hello World"}
	johnny1 := &String{Value: "johnny"}
	johnny2 := &String{Value: "johnny"}

	if hello1.HashKey() != hello2.HashKey() {
		t.Fatalf("string with same value have different hash keys")
	}
	if johnny1.HashKey() != johnny2.HashKey() {
		t.Fatalf("string with same value have different hash keys")
	}
	if hello1.HashKey() == johnny1.HashKey() {
		t.Fatalf("different string have same hash keys")
	}
}

func TestHashKeysBoolean(t *testing.T) {
	trueobj1 := &Boolean{Value: true}
	trueobj2 := &Boolean{Value: true}
	falseobj1 := &Boolean{Value: false}
	falseobj2 := &Boolean{Value: false}

	if trueobj1.HashKey() != trueobj2.HashKey() {
		t.Fatalf("string with same value have different hash keys")
	}
	if falseobj1.HashKey() != falseobj2.HashKey() {
		t.Fatalf("string with same value have different hash keys")
	}
	if trueobj1.HashKey() == falseobj1.HashKey() {
		t.Fatalf("different string have same hash keys")
	}
}

func TestHashKeysInteger(t *testing.T) {
	ninebj1 := &Integer{Value: 92}
	nineobj2 := &Integer{Value: 92}
	twoobj1 := &Integer{Value: 2}
	twoobj2 := &Integer{Value: 2}

	if ninebj1.HashKey() != nineobj2.HashKey() {
		t.Fatalf("string with same value have different hash keys")
	}
	if twoobj1.HashKey() != twoobj2.HashKey() {
		t.Fatalf("string with same value have different hash keys")
	}
	if ninebj1.HashKey() == twoobj1.HashKey() {
		t.Fatalf("different string have same hash keys")
	}
}
