package outer

import "testing"

func Test(t *testing.T) {
	var tp = &transportImpl{IP: "2.2.2.2",
		TargetHost: "1.1.1.1",
		TargetPort: 9000}
	transportMng.add(tp)
	t.Log(transportMng.tl.search(tp))
}
