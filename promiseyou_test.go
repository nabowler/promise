package promise_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/nabowler/promise"
)

// TestYou ensures expected behavior of promise.You in the happy path
// 1. the expected value and error are returned when ctx is not done
// 2. the expected value and error continue to be returned on all calls
// 3. subsequent calls to Complete do not block, and do not change the
// returned values of the Promise
func TestYou(t *testing.T) {
	for name, testcase := range testCases {
		tc := testcase
		t.Run(name, func(t *testing.T) {
			p, c := promise.You[string](context.Background())
			c(tc.val, tc.err)
			for i := 0; i < 10; i++ {
				c("something invalid", fmt.Errorf("invalid error"))
				av, ae := p()
				expect(t, tc.val, av)
				expect(t, tc.err, ae)
			}
		})
	}
}

// TestYouCancelled ensures expected behavior of promise.You on the non-happy path
// when the context is done
// 1. the default value of T and ctx.Err() are returned when ctx is done
// 2. the default value of T and ctx.Err() continue to be returned on all calls
// 3. subsequent calls to Complete do not block, and do not change the
// returned values of the Promise
func TestYouCancelled(t *testing.T) {
	for name, testcase := range testCases {
		tc := testcase
		var expectedVal string

		// cancel the context to ensure that it ctx.Done() will return
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		t.Run(name, func(t *testing.T) {
			p, c := promise.You[string](ctx)

			for i := 0; i < 10; i++ {
				av, ae := p()
				expect(t, expectedVal, av)
				expect(t, ctx.Err(), ae)
				c(tc.val, tc.err)
			}
		})
	}
}

// TestYouNoError ensures expected behavior of promise.YouNoError in the happy path
// 1. the expected value is returned when ctx is not done
// 2. the expected value continues to be returned on all calls
// 3. subsequent calls to Complete do not block, and do not change the
// returned value of the Promise
func TestYouNoError(t *testing.T) {
	for name, testcase := range noErrorTestCases {
		tc := testcase
		t.Run(name, func(t *testing.T) {
			p, c := promise.YouNoError[string](context.Background())
			c(tc)
			for i := 0; i < 10; i++ {
				c("something invalid")
				av := p()
				expect(t, tc, av)
			}
		})
	}
}

// TestYouNoErrorCancelled ensures expected behavior of promise.YouNoError on the non-happy path
// when the context is done
// 1. the default value of T is returned when ctx is done
// 2. the default value of T continues to be returned on all calls
// 3. subsequent calls to Complete do not block, and do not change the
// returned value of the Promise
func TestYouNoErrorCancelled(t *testing.T) {
	for name, testcase := range noErrorTestCases {
		tc := testcase
		var expectedVal string

		// cancel the context to ensure that it ctx.Done() will return
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		t.Run(name, func(t *testing.T) {
			p, c := promise.YouNoError[string](ctx)
			for i := 0; i < 10; i++ {
				av := p()
				expect(t, expectedVal, av)
				c(tc)
			}
		})
	}
}

// TestYouHTTPRequests tests a possible real-world use case of a promose.You
// by performing a single http.Reqest, and handing the results off to multiple consumers.
// Because this is fallible for reasons outside of our control, the test will only run
// when -short is not set and PROMISE_INT_TEST is configured to true in the environment.
func TestYouHTTPRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping due to short flag")
	}
	if !strings.EqualFold(os.Getenv("PROMISE_INT_TEST"), "true") {
		t.Skip("Skipping because PROMISE_INT_TEST is not set to true")
	}

	p, c := promise.You[int](context.Background())
	wg := sync.WaitGroup{}
	// provide the promise to multiple consumers
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

	// provide the complete to a producer
	go func() {
		time.Sleep(100 * time.Millisecond)

		resp, err := http.Get("https://example.org/")
		if err == nil {
			resp.Body.Close()
		}
		var status int
		if resp != nil {
			status = resp.StatusCode
		}

		// complete the promise
		c(status, err)
	}()

	// wait for consumers to finish
	wg.Wait()
}
