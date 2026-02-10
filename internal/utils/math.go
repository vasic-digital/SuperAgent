package utils

import "errors"

// Factorial calculates the factorial of a non-negative integer.
// Returns an error if n is negative or if the result would overflow int64.
// Maximum input: 20 (20! = 2432902008176640000)
func Factorial(n int) (int64, error) {
	if n < 0 {
		return 0, errors.New("factorial is not defined for negative numbers")
	}
	if n > 20 {
		return 0, errors.New("factorial result would overflow int64")
	}

	result := int64(1)
	for i := 2; i <= n; i++ {
		result *= int64(i)
	}
	return result, nil
}

// FactorialRecursive calculates factorial recursively.
// Less efficient than iterative version but demonstrates recursive approach.
func FactorialRecursive(n int) (int64, error) {
	if n < 0 {
		return 0, errors.New("factorial is not defined for negative numbers")
	}
	if n > 20 {
		return 0, errors.New("factorial result would overflow int64")
	}
	if n == 0 || n == 1 {
		return 1, nil
	}

	prev, err := FactorialRecursive(n - 1)
	if err != nil {
		return 0, err
	}
	return int64(n) * prev, nil
}
