package fburl

import (
	"testing"
)

func TestDefaultUrl(t *testing.T) {
	url := &URL{}
	if url.String() != "http://www.facebook.com/" {
		t.Fatalf("Got unexpected default URL %s.", url.String())
	}
}

func TestCustom(t *testing.T) {
	url := &URL{
		SubDomain: "apps",
		Path:      "/fbrelll",
		Env:       Beta,
	}
	if url.String() != "http://apps.beta.facebook.com/fbrelll" {
		t.Fatalf("Got unexpected our URL %s.", url.String())
	}
}

func TestOurFix(t *testing.T) {
	url := &URL{SubDomain: "our"}
	if url.String() != "http://our.intern.facebook.com/" {
		t.Fatalf("Got unexpected our URL %s.", url.String())
	}
}
