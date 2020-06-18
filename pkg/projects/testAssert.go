package projects

import (
	"context"
	"testing"
	"time"
)

type AssertFunc func() bool

func TimeWait(t *testing.T, timeout time.Duration) AssertFunc {
	t.Logf("timeWait %s formed\n", timeout)
	return func() bool {
		t.Logf("timeWait %s started\n", timeout)
		time.Sleep(timeout)
		t.Logf("timeWait %s finished\n", timeout)
		return true
	}
}

//createAssertSuccess method to return immediately and simply register a
//function with standard boolean return value handling.
func (tc *TestContext) CreateAssertSuccess(assert AssertFunc) {
	tc.asserts_sucess = append(tc.asserts_sucess, assert)
}

//createAssertSuccess method to return immediately and simply register a
//function with inversed boolean return value handling.
func (tc *TestContext) CreateAssertFail(assert AssertFunc) {
	tc.asserts_fail = append(tc.asserts_fail, assert)
}

//WaitForAsserts blocking execution until the time elapses or asserts fires
func (tc *TestContext) WaitForAsserts(t *testing.T, secs int) {
	timeout := time.Duration(secs) * time.Second
	t.Log("WaitForAsserts timewait: ", timeout)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cancel when we are finished tasks

	tc.CreateAssertFail(TimeWait(t, timeout))

	result := make(chan bool)

	for _, s := range tc.asserts_sucess {
		go func(afn AssertFunc) {
			t.Logf("AssertFunc+ %q", afn)
			res := afn()
			result <- res
			cancel()
			t.Logf("AssertFunc+ -> %t", res)
		}(s)
	}

	for _, f := range tc.asserts_fail {
		go func(afn AssertFunc) {
			t.Logf("AssertFunc- %q", afn)
			res := afn()
			result <- !res
			cancel()
			t.Logf("AssertFunc- -> %t", res)
		}(f)
	}

	select {
	case res := <-result:
		t.Log("assertFunc result: ", res)
		if !res {
			t.Fatal("assertFunc failed")
		}
	case <-time.After(timeout):
		t.Fatal("assertFunc terminated by timewat", timeout)
	case <-ctx.Done():
		t.Log("assertFunc finished")
	}
}
