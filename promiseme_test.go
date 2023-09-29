package promise_test

import (
	"context"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/nabowler/promise"
)

// TestMe ensures expected behavior of promise.Me in the happy path
// 1. the expected value and error are returned when ctx is not done
// 2. the expected value and error continue to be returned on all calls
func TestMe(t *testing.T) {
	for name, testcase := range testCases {
		tc := testcase
		t.Run(name, func(t *testing.T) {
			p := promise.Me(context.Background(), func() (string, error) {
				return tc.val, tc.err
			})
			for i := 0; i < 10; i++ {
				av, ae := p()
				expect(t, tc.val, av)
				expect(t, tc.err, ae)
			}
		})
	}
}

// TestMeCancelled ensures expected behavior of promise.Me on the non-happy path
// when the context is done
// 1. the default value of T and ctx.Err() are returned when ctx is done
// 2. the default value of T and ctx.Err() continue to be returned on all calls
func TestMeCancelled(t *testing.T) {
	for name, testcase := range testCases {
		tc := testcase
		var expectedVal string

		// cancel the context to ensure that it ctx.Done() will return
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		t.Run(name, func(t *testing.T) {
			p := promise.Me(ctx, func() (string, error) {
				// 3 seconds should be long enough for the runtime to
				// choose to select <- ctx.Done()
				time.Sleep(3 * time.Second)
				return tc.val, tc.err
			})

			for i := 0; i < 10; i++ {
				av, ae := p()
				expect(t, expectedVal, av)
				expect(t, ctx.Err(), ae)
			}
		})
	}
}

// TestMeNoError ensures expected behavior of promise.MeNoError in the happy path
// 1. the expected value is returned when ctx is not done
// 2. the expected value continues to be returned on all calls
func TestMeNoError(t *testing.T) {
	for name, testcase := range noErrorTestCases {
		tc := testcase
		t.Run(name, func(t *testing.T) {
			p := promise.MeNoError(context.Background(), func() string {
				return tc
			})
			for i := 0; i < 10; i++ {
				av := p()
				expect(t, tc, av)
			}
		})
	}
}

// TestMeNoErrorCancelled ensures expected behavior of promise.MeNoError on the non-happy path
// when the context is done
// 1. the default value of T is returned when ctx is done
// 2. the default value of T continues to be returned on all calls
func TestMeNoErrorCancelled(t *testing.T) {
	for name, testcase := range noErrorTestCases {
		tc := testcase
		var expectedVal string

		// cancel the context to ensure that it ctx.Done() will return
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		t.Run(name, func(t *testing.T) {
			p := promise.MeNoError(ctx, func() string {
				// 3 seconds should be long enough for the runtime to
				// choose to select <- ctx.Done()
				time.Sleep(3 * time.Second)
				return tc
			})
			for i := 0; i < 10; i++ {
				av := p()
				expect(t, expectedVal, av)
			}
		})
	}
}

// TestMeHTTPRequests tests a possible real-world use case of a promose.Me
// by performing a single http.Reqest, and handing the results off to multiple consumers.
// Because this is fallible for reasons outside of our control, the test will only run
// when -short is not set and PROMISE_INT_TEST is configured to true in the environment.
func TestMeHTTPRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping due to short flag")
	}
	if !strings.EqualFold(os.Getenv("PROMISE_INT_TEST"), "true") {
		t.Skip("Skipping because PROMISE_INT_TEST is not set to true")
	}

	fn := func() (int, error) {
		resp, err := http.Get("https://example.org/")
		if err == nil {
			resp.Body.Close()
		}
		time.Sleep(100 * time.Millisecond)
		if resp == nil {
			return 0, nil
		}
		return resp.StatusCode, err
	}

	p := promise.Me(context.Background(), fn)
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			status, err := p()
			expect(t, nil, err)
			expect(t, 200, status)
			t.Logf("%d: %d %v", idx, status, err)
			wg.Done()
		}(i)
	}

	wg.Wait()
}
