package stringx

import "testing"

func TestSplit(t *testing.T) {
	strArr := make([]string, 0)
	strArr = SplitStr("", ",")
	if len(strArr) != 0 {
		t.Error()
	}

	strArr = SplitStr("厉害了我的国0", "")
	if len(strArr) != 7 {
		t.Error()
	}

	strArr = SplitStr("厉害了,我的国", ",")
	if len(strArr) != 2 {
		t.Error()
	}

	strArr = SplitStr("厉,害,了,我,的,国,", ",")
	if len(strArr) != 7 {
		t.Error()
	}

	strArr = SplitStr("|||a||b", "|")
	if len(strArr) != 6 {
		t.Error()
	}

	strArr = SplitStr("|||a||b|", "|")
	if len(strArr) != 7 {
		t.Error()
	}
}
