# Quiz 8: Security & Performance Assessment

## Section A: Concurrency Safety (5 questions)

1. What Go pattern prevents double-close panics on channels?
   a) sync.Mutex
   b) sync.Once
   c) atomic.Bool
   d) context.WithCancel

2. Why is `atomic.Bool` preferred over a plain `bool` for concurrent flags?
   a) It's faster for single-threaded code
   b) It prevents data races without requiring a mutex lock
   c) It automatically retries on failure
   d) It's required by the Go specification

3. What is the correct shutdown pattern for a goroutine tracked by WaitGroup?
   a) wg.Add(1) inside goroutine, wg.Done() outside
   b) wg.Add(1) before goroutine, defer wg.Done() inside, wg.Wait() in Stop()
   c) wg.Add(1) in Stop(), wg.Done() in goroutine
   d) No WaitGroup needed if using channels

4. Why add `defer recover()` to participant goroutines in debate service?
   a) To improve performance
   b) To prevent a panic from skipping wg.Done(), causing the WaitGroup to hang
   c) To automatically retry the operation
   d) To log the error to a file

5. What does the `-race` flag do in `go test`?
   a) Makes tests run faster
   b) Enables the data race detector that finds concurrent memory access bugs
   c) Runs tests in parallel
   d) Limits CPU usage

## Section B: Security Scanning (5 questions)

6. What TLS version should be the minimum for any HTTPS server?
   a) TLS 1.0
   b) TLS 1.1
   c) TLS 1.2
   d) TLS 1.3

7. When is `nosemgrep` annotation appropriate?
   a) Always, to suppress noisy warnings
   b) Only for verified false positives with documented justification
   c) Never — all findings must be fixed
   d) Only for test files

8. What type of vulnerability is `exec.Command` with dynamic input?
   a) SQL Injection
   b) Command Injection
   c) XSS
   d) CSRF

9. Why is deserializing into `interface{}` flagged by security scanners?
   a) It's slower than typed deserialization
   b) It allows arbitrary data structures that could enable type confusion attacks
   c) It causes memory leaks
   d) It's deprecated in Go

10. What is the difference between Snyk and SonarQube?
    a) Snyk scans dependencies (SCA), SonarQube scans source code (SAST)
    b) They do exactly the same thing
    c) Snyk is for Go, SonarQube is for Java only
    d) Snyk is free, SonarQube is paid

## Section C: Performance & Resource Limits (5 questions)

11. Why must all tests use GOMAXPROCS=2 in this project?
    a) Go only supports 2 CPUs
    b) The host runs mission-critical processes; exceeding 30-40% CPU caused system crashes
    c) Tests are faster with fewer cores
    d) It's a Go best practice

12. What does `nice -n 19` do?
    a) Sets the process to lowest CPU scheduling priority
    b) Limits memory usage
    c) Enables verbose logging
    d) Runs the process in a container

13. What is the purpose of `ionice -c 3` in test commands?
    a) Network I/O limiting
    b) Sets idle I/O scheduling class (only uses disk when nothing else needs it)
    c) CPU throttling
    d) Memory page locking

14. How do you detect goroutine leaks in a test?
    a) Check if tests pass
    b) Compare runtime.NumGoroutine() before and after the test
    c) Look for error messages
    d) Run with -v flag

15. What does `-p 1` do in `go test`?
    a) Runs only 1 test
    b) Limits to 1 package being tested at a time (sequential package execution)
    c) Uses 1 CPU core
    d) Sets timeout to 1 second

---
**Answer Key:** 1-b, 2-b, 3-b, 4-b, 5-b, 6-c, 7-b, 8-b, 9-b, 10-a, 11-b, 12-a, 13-b, 14-b, 15-b
