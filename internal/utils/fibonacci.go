package utils

// Fibonacci calculates the nth Fibonacci number using iteration.
// Returns the Fibonacci number at position n (0-indexed).
func Fibonacci(n int) int {
	if n <= 1 {
		return n
	}

	a, b := 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	return b
}

// FibonacciSequence generates a slice of Fibonacci numbers up to n terms.
func FibonacciSequence(n int) []int {
	if n <= 0 {
		return []int{}
	}

	sequence := make([]int, n)
	for i := 0; i < n; i++ {
		sequence[i] = Fibonacci(i)
	}
	return sequence
}
