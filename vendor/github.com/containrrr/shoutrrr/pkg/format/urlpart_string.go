// Code generated by "stringer -type=URLPart -trimprefix URL"; DO NOT EDIT.

package format

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[URLQuery-0]
	_ = x[URLUser-1]
	_ = x[URLPassword-2]
	_ = x[URLHost-3]
	_ = x[URLPort-4]
	_ = x[URLPath-5]
}

const _URLPart_name = "QueryUserPasswordHostPortPath"

var _URLPart_index = [...]uint8{0, 5, 9, 17, 21, 25, 29}

func (i URLPart) String() string {
	if i < 0 || i >= URLPart(len(_URLPart_index)-1) {
		return "URLPart(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _URLPart_name[_URLPart_index[i]:_URLPart_index[i+1]]
}
