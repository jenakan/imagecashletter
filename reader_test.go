// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package imagecashletter

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestICLFileRead validates reading an ICL file
func TestICLFileRead(t *testing.T) {
	fd, err := os.Open(filepath.Join("test", "testdata", "BNK20180905121042882-A.icl"))
	if err != nil {
		t.Errorf("%T: %s", err, err)
	}
	defer fd.Close()
	r := NewReader(fd)
	_, err = r.Read()

	if err != nil {
		if p, ok := err.(*ParseError); ok {
			if e, ok := p.Err.(*BundleError); ok {
				if e.FieldName != "entries" {
					t.Errorf("%T: %s", e, e)
				}
			}
		} else {
			t.Errorf("%T: %s", err, err)
		}
	}

	err2 := r.File.Validate()

	if err2 != nil {
		if e, ok := err2.(*FileError); ok {
			if e.FieldName != "BundleCount" {
				t.Errorf("%T: %s", e, e)
			}
		} else {
			t.Errorf("%T: %s", err, err)
		}
	}
}

// TestICLFile validates reading an ICL file
func TestICLFile(t *testing.T) {
	fd, err := os.Open(filepath.Join("test", "testdata", "BNK20180905121042882-A.icl"))
	if err != nil {
		t.Fatalf("Can not open local file: %s: \n", err)
	}
	defer fd.Close()

	r := NewReader(fd)
	ICLFile, err := r.Read()
	if err != nil {
		t.Errorf("Issue reading file: %+v \n", err)
	}
	t.Logf("r.File.Header=%#v", r.File.Header)
	t.Logf("r.File.Control=%#v", r.File.Control)
	// ensure we have a validated file structure
	if ICLFile.Validate(); err != nil {
		t.Errorf("Could not validate entire read file: %v", err)
	}
}

// TestRecordTypeUnknown validates record type unknown
func TestRecordTypeUnknown(t *testing.T) {
	var line = "1735T231380104121042882201809051523NCitadel           Wells Fargo        US     "
	r := NewReader(strings.NewReader(line))
	_, err := r.Read()
	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FileError); ok {
			if e.FieldName != "recordType" {
				t.Errorf("%T: %s", e, e)
			}
		} else {
			t.Errorf("%T: %s", err, err)
		}
	}
}

//TestFileLineShort validates file line is short
func TestFileLineShort(t *testing.T) {
	line := "1 line is only 70 characters ........................................!"
	r := NewReader(strings.NewReader(line))
	_, err := r.Read()

	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FileError); ok {
			if e.FieldName != "RecordLength" {
				t.Errorf("%T: %s", e, e)
			}
		} else {
			t.Errorf("%T: %s", e, e)
		}
	}
}

func TestReaderCrash__parseBundleControl(t *testing.T) {
	r := &Reader{}
	if err := r.parseBundleControl(); err == nil {
		t.Error("expected error")
	}
}

// TestFileFileHeaderErr validates error flows back from the parser
func TestFileFileHeaderErr(t *testing.T) {
	fh := mockFileHeader()
	fh.ImmediateOrigin = ""
	r := NewReader(strings.NewReader(fh.String()))
	// necessary to have a file control not nil
	r.File.Control = mockFileControl()
	_, err := r.Read()
	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FieldError); ok {
			if !strings.Contains(e.Msg, msgFieldInclusion) {
				t.Errorf("%T: %s", e, e)
			}
		}
	} else {
		t.Errorf("%T: %s", err, err)
	}
}

// TestTwoFileHeaders validates one file header
func TestTwoFileHeaders(t *testing.T) {
	var line = "0135T231380104121042882201809051523NCitadel           Wells Fargo        US     "
	var twoHeaders = line + "\n" + line
	r := NewReader(strings.NewReader(twoHeaders))
	_, err := r.Read()
	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FileError); ok {
			if e.Msg != msgFileControl {
				t.Errorf("%T: %s", e, e)
			}
		} else {
			t.Errorf("%T: %s", err, err)
		}
	}
}

// TestCashLetterHeaderErr validates error flows back from the parser
func TestCashLetterHeaderErr(t *testing.T) {
	clh := mockCashLetterHeader()
	clh.DestinationRoutingNumber = ""
	r := NewReader(strings.NewReader(clh.String()))
	_, err := r.Read()
	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FieldError); ok {
			if !strings.Contains(e.Msg, msgFieldInclusion) {
				t.Errorf("%T: %s", e, e)
			}
		}
	} else {
		t.Errorf("%T: %s", err, err)
	}
}

// TestCashLetterHeaderDuplicate validates when two CashLetterHeader exists in a current CashLetter
func TestCashLetterHeaderDuplicate(t *testing.T) {
	// create a new CashLetter header string
	clh := mockCashLetterHeader()
	r := NewReader(strings.NewReader(clh.String()))
	// instantiate a CashLetter in the reader
	r.addCurrentCashLetter(NewCashLetter(clh))
	// read should fail because it is parsing a second CashLetter Header and there can only be one.
	_, err := r.Read()
	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FileError); ok {
			if e.Msg != msgFileCashLetterInside {
				t.Errorf("%T: %s", e, e)
			}
		}
	} else {
		t.Errorf("%T: %s", err, err)
	}
}

// TestBundleHeaderErr validates error flows back from the parser
func TestBundleHeaderErr(t *testing.T) {
	bh := mockBundleHeader()
	bh.DestinationRoutingNumber = ""
	r := NewReader(strings.NewReader(bh.String()))
	_, err := r.Read()
	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FieldError); ok {
			if !strings.Contains(e.Msg, msgFieldInclusion) {
				t.Errorf("%T: %s", e, e)
			}
		}
	} else {
		t.Errorf("%T: %s", err, err)
	}
}

// TestBundleHeaderDuplicate validates when two BundleHeader exists in a current Bundle
func TestBundleHeaderDuplicate(t *testing.T) {
	// create a new CashLetter header string
	bh := mockBundleHeader()
	r := NewReader(strings.NewReader(bh.String()))
	// instantiate a CashLetter in the reader
	clh := mockCashLetterHeader()
	r.addCurrentCashLetter(NewCashLetter(clh))
	bhTwo := mockBundleHeader()
	r.addCurrentBundle(NewBundle(bhTwo))
	// read should fail because it is parsing a second CashLetter Header and there can only be one.
	_, err := r.Read()
	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FileError); ok {
			if e.Msg != msgFileBundleInside {
				t.Errorf("%T: %s", e, e)
			}
		}
	} else {
		t.Errorf("%T: %s", err, err)
	}
}

// TestCheckDetailError validates error flows back from the parser
func TestCheckDetailError(t *testing.T) {
	cd := mockCheckDetail()
	cd.PayorBankRoutingNumber = ""
	r := NewReader(strings.NewReader(cd.String()))
	clh := mockCashLetterHeader()
	r.addCurrentCashLetter(NewCashLetter(clh))
	bh := mockBundleHeader()
	b := NewBundle(bh)
	r.currentCashLetter.AddBundle(b)
	r.addCurrentBundle(b)
	_, err := r.Read()
	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FieldError); ok {
			if !strings.Contains(e.Msg, msgFieldInclusion) {
				t.Errorf("%T: %s", e, e)
			}
		}
	} else {
		t.Errorf("%T: %s", err, err)
	}
}

// TestCheckDetailAddendumABundleError validates error flows back from the parser
func TestCheckDetailAddendumABundleError(t *testing.T) {
	cd := mockCheckDetail()
	cdaddendumA := mockCheckDetailAddendumA()
	cd.AddCheckDetailAddendumA(cdaddendumA)
	r := NewReader(strings.NewReader(cdaddendumA.String()))
	clh := mockCashLetterHeader()
	r.addCurrentCashLetter(NewCashLetter(clh))
	bh := mockBundleHeader()
	b := NewBundle(bh)
	r.currentCashLetter.AddBundle(b)
	r.addCurrentBundle(b)
	_, err := r.Read()
	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FileError); ok {
			if e.Msg != msgFileBundleOutside {
				t.Errorf("%T: %s", e, e)
			}
		}
	} else {
		t.Errorf("%T: %s", err, err)
	}
}

// TestCheckDetailAddendumBBundleError validates error flows back from the parser
func TestCheckDetailAddendumBBundleError(t *testing.T) {
	cd := mockCheckDetail()
	cdaddendumA := mockCheckDetailAddendumA()
	cd.AddCheckDetailAddendumA(cdaddendumA)
	cdaddendumB := mockCheckDetailAddendumB()
	cd.AddCheckDetailAddendumB(cdaddendumB)
	r := NewReader(strings.NewReader(cdaddendumB.String()))
	clh := mockCashLetterHeader()
	r.addCurrentCashLetter(NewCashLetter(clh))
	bh := mockBundleHeader()
	b := NewBundle(bh)
	r.currentCashLetter.AddBundle(b)
	r.addCurrentBundle(b)
	_, err := r.Read()
	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FileError); ok {
			if e.Msg != msgFileBundleOutside {
				t.Errorf("%T: %s", e, e)
			}
		}
	} else {
		t.Errorf("%T: %s", err, err)
	}
}

// TestCheckDetailAddendumCBundleError validates error flows back from the parser
func TestCheckDetailAddendumCBundleError(t *testing.T) {
	cd := mockCheckDetail()
	cdAddendumA := mockCheckDetailAddendumA()
	cd.AddCheckDetailAddendumA(cdAddendumA)
	cdAddendumB := mockCheckDetailAddendumB()
	cd.AddCheckDetailAddendumB(cdAddendumB)
	cdAddendumC := mockCheckDetailAddendumC()
	cd.AddCheckDetailAddendumC(cdAddendumC)
	r := NewReader(strings.NewReader(cdAddendumC.String()))
	clh := mockCashLetterHeader()
	r.addCurrentCashLetter(NewCashLetter(clh))
	bh := mockBundleHeader()
	b := NewBundle(bh)
	r.currentCashLetter.AddBundle(b)
	r.addCurrentBundle(b)
	_, err := r.Read()
	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FileError); ok {
			if e.Msg != msgFileBundleOutside {
				t.Errorf("%T: %s", e, e)
			}
		}
	} else {
		t.Errorf("%T: %s", err, err)
	}
}

// TestCheckDetailAddendumAError validates error flows back from the parser
func TestCheckDetailAddendumAError(t *testing.T) {
	cd := mockCheckDetail()
	cdAddendumA := mockCheckDetailAddendumA()
	cdAddendumA.ReturnLocationRoutingNumber = ""
	cd.AddCheckDetailAddendumA(cdAddendumA)
	r := NewReader(strings.NewReader(cdAddendumA.String()))
	clh := mockCashLetterHeader()
	r.addCurrentCashLetter(NewCashLetter(clh))
	bh := mockBundleHeader()
	b := NewBundle(bh)
	b.AddCheckDetail(cd)
	r.currentCashLetter.AddBundle(b)
	r.addCurrentBundle(b)
	_, err := r.Read()
	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FieldError); ok {
			if !strings.Contains(e.Msg, msgFieldInclusion) {
				t.Errorf("%T: %s", e, e)
			}
		}
	} else {
		t.Errorf("%T: %s", err, err)
	}
}

// TestCheckDetailAddendumBError validates error flows back from the parser
func TestCheckDetailAddendumBError(t *testing.T) {
	cd := mockCheckDetail()
	cdAddendumA := mockCheckDetailAddendumA()
	cd.AddCheckDetailAddendumA(cdAddendumA)
	cdAddendumB := mockCheckDetailAddendumB()
	cdAddendumB.MicrofilmArchiveSequenceNumber = "               "
	cd.AddCheckDetailAddendumB(cdAddendumB)
	r := NewReader(strings.NewReader(cdAddendumB.String()))
	clh := mockCashLetterHeader()
	r.addCurrentCashLetter(NewCashLetter(clh))
	bh := mockBundleHeader()
	b := NewBundle(bh)
	b.AddCheckDetail(cd)
	r.currentCashLetter.AddBundle(b)
	r.addCurrentBundle(b)
	_, err := r.Read()
	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FieldError); ok {
			if !strings.Contains(e.Msg, msgFieldInclusion) {
				t.Errorf("%T: %s", e, e)
			}
		}
	} else {
		t.Errorf("%T: %s", err, err)
	}
}

// TestCheckDetailAddendumCError validates error flows back from the parser
func TestCheckDetailAddendumCError(t *testing.T) {
	cd := mockCheckDetail()
	cdAddendumA := mockCheckDetailAddendumA()
	cd.AddCheckDetailAddendumA(cdAddendumA)
	cdAddendumB := mockCheckDetailAddendumB()
	cd.AddCheckDetailAddendumB(cdAddendumB)
	cdAddendumC := mockCheckDetailAddendumC()
	cdAddendumC.EndorsingBankRoutingNumber = ""
	cd.AddCheckDetailAddendumC(cdAddendumC)
	r := NewReader(strings.NewReader(cdAddendumC.String()))
	clh := mockCashLetterHeader()
	r.addCurrentCashLetter(NewCashLetter(clh))
	bh := mockBundleHeader()
	b := NewBundle(bh)
	b.AddCheckDetail(cd)
	r.currentCashLetter.AddBundle(b)
	r.addCurrentBundle(b)
	_, err := r.Read()
	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FieldError); ok {
			if !strings.Contains(e.Msg, msgFieldInclusion) {
				t.Errorf("%T: %s", e, e)
			}
		}
	} else {
		t.Errorf("%T: %s", err, err)
	}
}

// TestReturnDetailError validates error flows back from the parser
func TestReturnDetailError(t *testing.T) {
	rd := mockReturnDetail()
	rd.PayorBankRoutingNumber = ""
	r := NewReader(strings.NewReader(rd.String()))
	clh := mockCashLetterHeader()
	r.addCurrentCashLetter(NewCashLetter(clh))
	bh := mockBundleHeader()
	b := NewBundle(bh)
	r.currentCashLetter.AddBundle(b)
	r.addCurrentBundle(b)
	_, err := r.Read()
	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FieldError); ok {
			if !strings.Contains(e.Msg, msgFieldInclusion) {
				t.Errorf("%T: %s", e, e)
			}
		}
	} else {
		t.Errorf("%T: %s", err, err)
	}
}

// TestReturnDetailAddendumABundleError validates error flows back from the parser
func TestReturnDetailAddendumABundleError(t *testing.T) {
	rd := mockReturnDetail()
	rdAddendumA := mockReturnDetailAddendumA()
	rd.AddReturnDetailAddendumA(rdAddendumA)
	r := NewReader(strings.NewReader(rdAddendumA.String()))
	clh := mockCashLetterHeader()
	r.addCurrentCashLetter(NewCashLetter(clh))
	bh := mockBundleHeader()
	b := NewBundle(bh)
	r.currentCashLetter.AddBundle(b)
	r.addCurrentBundle(b)
	_, err := r.Read()
	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FileError); ok {
			if e.Msg != msgFileBundleOutside {
				t.Errorf("%T: %s", e, e)
			}
		}
	} else {
		t.Errorf("%T: %s", err, err)
	}
}

// TestReturnDetailAddendumBBundleError validates error flows back from the parser
func TestReturnDetailAddendumBBundleError(t *testing.T) {
	rd := mockReturnDetail()
	rdAddendumA := mockReturnDetailAddendumA()
	rd.AddReturnDetailAddendumA(rdAddendumA)
	rdAddendumB := mockReturnDetailAddendumB()
	rd.AddReturnDetailAddendumB(rdAddendumB)
	r := NewReader(strings.NewReader(rdAddendumB.String()))
	clh := mockCashLetterHeader()
	r.addCurrentCashLetter(NewCashLetter(clh))
	bh := mockBundleHeader()
	b := NewBundle(bh)
	r.currentCashLetter.AddBundle(b)
	r.addCurrentBundle(b)
	_, err := r.Read()
	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FileError); ok {
			if e.Msg != msgFileBundleOutside {
				t.Errorf("%T: %s", e, e)
			}
		}
	} else {
		t.Errorf("%T: %s", err, err)
	}
}

// TestReturnDetailAddendumCBundleError validates error flows back from the parser
func TestReturnDetailAddendumCBundleError(t *testing.T) {
	rd := mockReturnDetail()
	rdAddendumA := mockReturnDetailAddendumA()
	rd.AddReturnDetailAddendumA(rdAddendumA)
	rdAddendumB := mockReturnDetailAddendumB()
	rd.AddReturnDetailAddendumB(rdAddendumB)
	rdAddendumC := mockReturnDetailAddendumC()
	rd.AddReturnDetailAddendumC(rdAddendumC)
	r := NewReader(strings.NewReader(rdAddendumC.String()))
	clh := mockCashLetterHeader()
	r.addCurrentCashLetter(NewCashLetter(clh))
	bh := mockBundleHeader()
	b := NewBundle(bh)
	r.currentCashLetter.AddBundle(b)
	r.addCurrentBundle(b)
	_, err := r.Read()
	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FileError); ok {
			if e.Msg != msgFileBundleOutside {
				t.Errorf("%T: %s", e, e)
			}
		}
	} else {
		t.Errorf("%T: %s", err, err)
	}
}

// TestReturnDetailAddendumDBundleError validates error flows back from the parser
func TestReturnDetailAddendumDBundleError(t *testing.T) {
	rd := mockReturnDetail()
	rdAddendumA := mockReturnDetailAddendumA()
	rd.AddReturnDetailAddendumA(rdAddendumA)
	rdAddendumB := mockReturnDetailAddendumB()
	rd.AddReturnDetailAddendumB(rdAddendumB)
	rdAddendumC := mockReturnDetailAddendumC()
	rd.AddReturnDetailAddendumC(rdAddendumC)
	rdAddendumD := mockReturnDetailAddendumD()
	rd.AddReturnDetailAddendumD(rdAddendumD)
	r := NewReader(strings.NewReader(rdAddendumD.String()))
	clh := mockCashLetterHeader()
	r.addCurrentCashLetter(NewCashLetter(clh))
	bh := mockBundleHeader()
	b := NewBundle(bh)
	r.currentCashLetter.AddBundle(b)
	r.addCurrentBundle(b)
	_, err := r.Read()
	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FileError); ok {
			if e.Msg != msgFileBundleOutside {
				t.Errorf("%T: %s", e, e)
			}
		}
	} else {
		t.Errorf("%T: %s", err, err)
	}
}

// TestReturnDetailAddendumAError validates error flows back from the parser
func TestReturnDetailAddendumAError(t *testing.T) {
	rd := mockReturnDetail()
	rdAddendumA := mockReturnDetailAddendumA()
	rdAddendumA.ReturnLocationRoutingNumber = ""
	rd.AddReturnDetailAddendumA(rdAddendumA)
	r := NewReader(strings.NewReader(rdAddendumA.String()))
	clh := mockCashLetterHeader()
	r.addCurrentCashLetter(NewCashLetter(clh))
	bh := mockBundleHeader()
	b := NewBundle(bh)
	b.AddReturnDetail(rd)
	r.currentCashLetter.AddBundle(b)
	r.addCurrentBundle(b)
	_, err := r.Read()
	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FieldError); ok {
			if !strings.Contains(e.Msg, msgFieldInclusion) {
				t.Errorf("%T: %s", e, e)
			}
		}
	} else {
		t.Errorf("%T: %s", err, err)
	}
}

// TestReturnDetailAddendumBError validates error flows back from the parser
func TestReturnDetailAddendumBError(t *testing.T) {
	rd := mockReturnDetail()
	rdAddendumA := mockReturnDetailAddendumA()
	rd.AddReturnDetailAddendumA(rdAddendumA)
	rdAddendumB := mockReturnDetailAddendumB()
	rdAddendumB.PayorBankSequenceNumber = "               "
	rd.AddReturnDetailAddendumB(rdAddendumB)
	r := NewReader(strings.NewReader(rdAddendumB.String()))
	clh := mockCashLetterHeader()
	r.addCurrentCashLetter(NewCashLetter(clh))
	bh := mockBundleHeader()
	b := NewBundle(bh)
	b.AddReturnDetail(rd)
	r.currentCashLetter.AddBundle(b)
	r.addCurrentBundle(b)
	_, err := r.Read()
	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FieldError); ok {
			if !strings.Contains(e.Msg, msgFieldInclusion) {
				t.Errorf("%T: %s", e, e)
			}
		}
	} else {
		t.Errorf("%T: %s", err, err)
	}
}

// TestReturnDetailAddendumCError validates error flows back from the parser
func TestReturnDetailAddendumCError(t *testing.T) {
	rd := mockReturnDetail()
	rdAddendumA := mockReturnDetailAddendumA()
	rd.AddReturnDetailAddendumA(rdAddendumA)
	rdAddendumB := mockReturnDetailAddendumB()
	rd.AddReturnDetailAddendumB(rdAddendumB)
	rdAddendumC := mockReturnDetailAddendumC()
	rdAddendumC.MicrofilmArchiveSequenceNumber = "               "
	rd.AddReturnDetailAddendumC(rdAddendumC)
	r := NewReader(strings.NewReader(rdAddendumC.String()))
	clh := mockCashLetterHeader()
	r.addCurrentCashLetter(NewCashLetter(clh))
	bh := mockBundleHeader()
	b := NewBundle(bh)
	b.AddReturnDetail(rd)
	r.currentCashLetter.AddBundle(b)
	r.addCurrentBundle(b)
	_, err := r.Read()
	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FieldError); ok {
			if !strings.Contains(e.Msg, msgFieldInclusion) {
				t.Errorf("%T: %s", e, e)
			}
		}
	} else {
		t.Errorf("%T: %s", err, err)
	}
}

// TestReturnDetailAddendumDError validates error flows back from the parser
func TestReturnDetailAddendumDError(t *testing.T) {
	rd := mockReturnDetail()
	rdAddendumA := mockReturnDetailAddendumA()
	rd.AddReturnDetailAddendumA(rdAddendumA)
	rdAddendumB := mockReturnDetailAddendumB()
	rd.AddReturnDetailAddendumB(rdAddendumB)
	rdAddendumC := mockReturnDetailAddendumC()
	rd.AddReturnDetailAddendumC(rdAddendumC)
	rdAddendumD := mockReturnDetailAddendumD()
	rdAddendumD.EndorsingBankRoutingNumber = "000000000"
	rd.AddReturnDetailAddendumD(rdAddendumD)
	r := NewReader(strings.NewReader(rdAddendumD.String()))
	clh := mockCashLetterHeader()
	r.addCurrentCashLetter(NewCashLetter(clh))
	bh := mockBundleHeader()
	b := NewBundle(bh)
	b.AddReturnDetail(rd)
	r.currentCashLetter.AddBundle(b)
	r.addCurrentBundle(b)
	_, err := r.Read()
	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FieldError); ok {
			if !strings.Contains(e.Msg, msgFieldInclusion) {
				t.Errorf("%T: %s", e, e)
			}
		}
	} else {
		t.Errorf("%T: %s", err, err)
	}
}

// TestCheckDetailBundleError validates error flows back from the parser
func TestCheckDetailBundleError(t *testing.T) {
	cd := mockCheckDetail()
	r := NewReader(strings.NewReader(cd.String()))
	clh := mockCashLetterHeader()
	r.addCurrentCashLetter(NewCashLetter(clh))
	bh := mockBundleHeader()
	b := NewBundle(bh)
	r.currentCashLetter.AddBundle(b)
	_, err := r.Read()
	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FileError); ok {
			if e.Msg != msgFileBundleOutside {
				t.Errorf("%T: %s", e, e)
			}
		}
	} else {
		t.Errorf("%T: %s", err, err)
	}
}

// TestReturnDetailBundleError validates error flows back from the parser
func TestReturnDetailBundleError(t *testing.T) {
	rd := mockReturnDetail()
	r := NewReader(strings.NewReader(rd.String()))
	clh := mockCashLetterHeader()
	r.addCurrentCashLetter(NewCashLetter(clh))
	bh := mockBundleHeader()
	b := NewBundle(bh)
	r.currentCashLetter.AddBundle(b)
	_, err := r.Read()
	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FileError); ok {
			if e.Msg != msgFileBundleOutside {
				t.Errorf("%T: %s", e, e)
			}
		}
	} else {
		t.Errorf("%T: %s", err, err)
	}
}

// TestCheckDetailIVDetailError validates error flows back from the parser
func TestCheckDetailIVDetailError(t *testing.T) {
	cd := mockCheckDetail()
	cdAddendumA := mockCheckDetailAddendumA()
	cd.AddCheckDetailAddendumA(cdAddendumA)
	cdAddendumB := mockCheckDetailAddendumB()
	cd.AddCheckDetailAddendumB(cdAddendumB)
	ivDetail := mockImageViewDetail()
	ivDetail.ViewDescriptor = ""
	cd.AddImageViewDetail(ivDetail)
	r := NewReader(strings.NewReader(ivDetail.String()))
	clh := mockCashLetterHeader()
	r.addCurrentCashLetter(NewCashLetter(clh))
	bh := mockBundleHeader()
	b := NewBundle(bh)
	b.AddCheckDetail(cd)
	r.currentCashLetter.AddBundle(b)
	r.addCurrentBundle(b)
	_, err := r.Read()
	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FieldError); ok {
			if !strings.Contains(e.Msg, msgFieldInclusion) {
				t.Errorf("%T: %s", e, e)
			}
		}
	} else {
		t.Errorf("%T: %s", err, err)
	}
}

// TestCheckDetailIVDataError validates error flows back from the parser
func TestCheckDetailIVDataError(t *testing.T) {
	cd := mockCheckDetail()
	cdAddendumA := mockCheckDetailAddendumA()
	cd.AddCheckDetailAddendumA(cdAddendumA)
	cdAddendumB := mockCheckDetailAddendumB()
	cd.AddCheckDetailAddendumB(cdAddendumB)
	ivd := mockImageViewDetail()
	cd.AddImageViewDetail(ivd)
	ivData := mockImageViewData()
	ivData.EceInstitutionRoutingNumber = "000000000"
	cd.AddImageViewData(ivData)
	r := NewReader(strings.NewReader(ivData.String()))
	clh := mockCashLetterHeader()
	r.addCurrentCashLetter(NewCashLetter(clh))
	bh := mockBundleHeader()
	b := NewBundle(bh)
	b.AddCheckDetail(cd)
	r.currentCashLetter.AddBundle(b)
	r.addCurrentBundle(b)
	_, err := r.Read()
	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FieldError); ok {
			if !strings.Contains(e.Msg, msgFieldInclusion) {
				t.Errorf("%T: %s", e, e)
			}
		}
	} else {
		t.Errorf("%T: %s", err, err)
	}
}

// TestCheckDetailIVAnalysisError validates error flows back from the parser
func TestCheckDetailIVAnalysisError(t *testing.T) {
	cd := mockCheckDetail()
	cdAddendumA := mockCheckDetailAddendumA()
	cd.AddCheckDetailAddendumA(cdAddendumA)
	cdAddendumB := mockCheckDetailAddendumB()
	cd.AddCheckDetailAddendumB(cdAddendumB)
	ivd := mockImageViewDetail()
	cd.AddImageViewDetail(ivd)
	ivData := mockImageViewData()
	cd.AddImageViewData(ivData)
	ivAnalysis := mockImageViewAnalysis()
	ivAnalysis.GlobalImageQuality = 9
	cd.AddImageViewAnalysis(ivAnalysis)
	r := NewReader(strings.NewReader(ivAnalysis.String()))
	clh := mockCashLetterHeader()
	r.addCurrentCashLetter(NewCashLetter(clh))
	bh := mockBundleHeader()
	b := NewBundle(bh)
	b.AddCheckDetail(cd)
	r.currentCashLetter.AddBundle(b)
	r.addCurrentBundle(b)
	_, err := r.Read()
	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FieldError); ok {
			if e.FieldName != "GlobalImageQuality" {
				t.Errorf("%T: %s", e, e)
			}
		}
	} else {
		t.Errorf("%T: %s", err, err)
	}
}

// TestReturnDetailIVDetailError validates error flows back from the parser
func TestReturnDetailIVDetailError(t *testing.T) {
	rd := mockReturnDetail()
	rdAddendumA := mockReturnDetailAddendumA()
	rd.AddReturnDetailAddendumA(rdAddendumA)
	rdAddendumC := mockReturnDetailAddendumC()
	rd.AddReturnDetailAddendumC(rdAddendumC)
	ivDetail := mockImageViewDetail()
	ivDetail.ViewDescriptor = ""
	rd.AddImageViewDetail(ivDetail)
	r := NewReader(strings.NewReader(ivDetail.String()))
	clh := mockCashLetterHeader()
	r.addCurrentCashLetter(NewCashLetter(clh))
	bh := mockBundleHeader()
	b := NewBundle(bh)
	b.AddReturnDetail(rd)
	r.currentCashLetter.AddBundle(b)
	r.addCurrentBundle(b)
	_, err := r.Read()
	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FieldError); ok {
			if !strings.Contains(e.Msg, msgFieldInclusion) {
				t.Errorf("%T: %s", e, e)
			}
		}
	} else {
		t.Errorf("%T: %s", err, err)
	}
}

// TestReturnDetailIVDataError validates error flows back from the parser
func TestReturnDetailIVDataError(t *testing.T) {
	rd := mockReturnDetail()
	rdAddendumA := mockReturnDetailAddendumA()
	rd.AddReturnDetailAddendumA(rdAddendumA)
	rdAddendumC := mockReturnDetailAddendumC()
	rd.AddReturnDetailAddendumC(rdAddendumC)
	ivDetail := mockImageViewDetail()
	rd.AddImageViewDetail(ivDetail)
	ivData := mockImageViewData()
	ivData.EceInstitutionRoutingNumber = "000000000"
	r := NewReader(strings.NewReader(ivData.String()))
	clh := mockCashLetterHeader()
	r.addCurrentCashLetter(NewCashLetter(clh))
	bh := mockBundleHeader()
	b := NewBundle(bh)
	b.AddReturnDetail(rd)
	r.currentCashLetter.AddBundle(b)
	r.addCurrentBundle(b)
	_, err := r.Read()
	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FieldError); ok {
			if !strings.Contains(e.Msg, msgFieldInclusion) {
				t.Errorf("%T: %s", e, e)
			}
		}
	} else {
		t.Errorf("%T: %s", err, err)
	}
}

// TestReturnDetailIVAnalysisError validates error flows back from the parser
func TestReturnDetailIVAnalysisError(t *testing.T) {
	rd := mockReturnDetail()
	rdAddendumA := mockReturnDetailAddendumA()
	rd.AddReturnDetailAddendumA(rdAddendumA)
	rdAddendumC := mockReturnDetailAddendumC()
	rd.AddReturnDetailAddendumC(rdAddendumC)
	ivDetail := mockImageViewDetail()
	rd.AddImageViewDetail(ivDetail)
	ivData := mockImageViewData()
	rd.AddImageViewData(ivData)
	ivAnalysis := mockImageViewAnalysis()
	ivAnalysis.GlobalImageQuality = 9
	rd.AddImageViewAnalysis(ivAnalysis)
	r := NewReader(strings.NewReader(ivAnalysis.String()))
	clh := mockCashLetterHeader()
	r.addCurrentCashLetter(NewCashLetter(clh))
	bh := mockBundleHeader()
	b := NewBundle(bh)
	b.AddReturnDetail(rd)
	r.currentCashLetter.AddBundle(b)
	r.addCurrentBundle(b)
	_, err := r.Read()
	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FieldError); ok {
			if e.FieldName != "GlobalImageQuality" {
				t.Errorf("%T: %s", e, e)
			}
		}
	} else {
		t.Errorf("%T: %s", err, err)
	}
}

// TestIVDetailBundleError validates error flows back from the parser
func TestIVDetailBundleError(t *testing.T) {
	cd := mockCheckDetail()
	cdAddendumA := mockCheckDetailAddendumA()
	cd.AddCheckDetailAddendumA(cdAddendumA)
	cdAddendumB := mockCheckDetailAddendumB()
	cd.AddCheckDetailAddendumB(cdAddendumB)
	ivDetail := mockImageViewDetail()
	cd.AddImageViewDetail(ivDetail)
	r := NewReader(strings.NewReader(ivDetail.String()))
	clh := mockCashLetterHeader()
	r.addCurrentCashLetter(NewCashLetter(clh))
	bh := mockBundleHeader()
	b := NewBundle(bh)
	//b.AddCheckDetail(cd)
	r.currentCashLetter.AddBundle(b)
	r.addCurrentBundle(b)
	_, err := r.Read()
	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FileError); ok {
			if e.Msg != msgFileBundleOutside {
				t.Errorf("%T: %s", e, e)
			}
		}
	} else {
		t.Errorf("%T: %s", err, err)
	}
}

// TestIVDataBundleError validates error flows back from the parser
func TestIVDataBundleError(t *testing.T) {
	cd := mockCheckDetail()
	cdAddendumA := mockCheckDetailAddendumA()
	cd.AddCheckDetailAddendumA(cdAddendumA)
	cdAddendumB := mockCheckDetailAddendumB()
	cd.AddCheckDetailAddendumB(cdAddendumB)
	ivDetail := mockImageViewDetail()
	cd.AddImageViewDetail(ivDetail)
	ivData := mockImageViewData()
	cd.AddImageViewData(ivData)
	r := NewReader(strings.NewReader(ivData.String()))
	clh := mockCashLetterHeader()
	r.addCurrentCashLetter(NewCashLetter(clh))
	bh := mockBundleHeader()
	b := NewBundle(bh)

	r.currentCashLetter.AddBundle(b)
	r.addCurrentBundle(b)
	_, err := r.Read()
	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FileError); ok {
			if e.Msg != msgFileBundleOutside {
				t.Errorf("%T: %s", e, e)
			}
		}
	} else {
		t.Errorf("%T: %s", err, err)
	}
}

// TestIVAnalysisBundleError validates error flows back from the parser
func TestIVAnalysisBundleError(t *testing.T) {
	cd := mockCheckDetail()
	cdAddendumA := mockCheckDetailAddendumA()
	cd.AddCheckDetailAddendumA(cdAddendumA)
	cdAddendumB := mockCheckDetailAddendumB()
	cd.AddCheckDetailAddendumB(cdAddendumB)
	ivDetail := mockImageViewDetail()
	cd.AddImageViewDetail(ivDetail)
	ivData := mockImageViewData()
	cd.AddImageViewData(ivData)
	ivAnalysis := mockImageViewAnalysis()
	cd.AddImageViewAnalysis(ivAnalysis)
	r := NewReader(strings.NewReader(ivAnalysis.String()))
	clh := mockCashLetterHeader()
	r.addCurrentCashLetter(NewCashLetter(clh))
	bh := mockBundleHeader()
	b := NewBundle(bh)

	r.currentCashLetter.AddBundle(b)
	r.addCurrentBundle(b)
	_, err := r.Read()
	if p, ok := err.(*ParseError); ok {
		if e, ok := p.Err.(*FileError); ok {
			if e.Msg != msgFileBundleOutside {
				t.Errorf("%T: %s", e, e)
			}
		}
	} else {
		t.Errorf("%T: %s", err, err)
	}
}

// TestICLCreditItemFile validates reading an ICL file with a CreditItem
func TestICLCreditItemFile(t *testing.T) {
	fd, err := os.Open(filepath.Join("test", "testdata", "BNK20181010121042882-A.icl"))
	if err != nil {
		t.Fatalf("Can not open local file: %s: \n", err)
	}
	defer fd.Close()

	ICLFile, err := NewReader(fd).Read()
	if err != nil {
		t.Errorf("Issue reading file: %+v \n", err)
	}
	// ensure we have a validated file structure
	if ICLFile.Validate(); err != nil {
		t.Errorf("Could not validate entire read file: %v", err)
	}
}

func TestICLBase64ImageData(t *testing.T) {
	bs, err := ioutil.ReadFile(filepath.Join("test", "testdata", "base64-encoded-images.json"))
	if err != nil {
		t.Fatal(err)
	}

	file, err := FileFromJSON(bs)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := NewWriter(&buf).Write(file); err != nil {
		t.Fatal(err)
	}

	if !bytes.Contains(buf.Bytes(), []byte("hello, world")) {
		t.Fatalf("unexpected ICL file:\n%s", buf.String())
	}
}
