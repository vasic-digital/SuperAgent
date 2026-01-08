"""Tests for HelixAgent SDK exceptions."""

import unittest

from helixagent.exceptions import (
    HelixAgentError,
    AuthenticationError,
    RateLimitError,
    APIError,
    ConnectionError,
    ValidationError,
    TimeoutError,
    raise_for_status,
)


class TestHelixAgentError(unittest.TestCase):
    """Test HelixAgentError base exception."""

    def test_basic_error(self):
        """Test basic error."""
        error = HelixAgentError("Something went wrong")
        self.assertEqual(error.message, "Something went wrong")
        self.assertIsNone(error.status_code)
        self.assertIsNone(error.response)

    def test_error_with_status_code(self):
        """Test error with status code."""
        error = HelixAgentError("Error", status_code=500)
        self.assertEqual(error.status_code, 500)

    def test_error_with_response(self):
        """Test error with response."""
        response = {"error": {"message": "Details"}}
        error = HelixAgentError("Error", response=response)
        self.assertEqual(error.response, response)

    def test_str_with_status_code(self):
        """Test string representation with status code."""
        error = HelixAgentError("Not found", status_code=404)
        self.assertEqual(str(error), "[404] Not found")

    def test_str_without_status_code(self):
        """Test string representation without status code."""
        error = HelixAgentError("Some error")
        self.assertEqual(str(error), "Some error")


class TestAuthenticationError(unittest.TestCase):
    """Test AuthenticationError exception."""

    def test_inheritance(self):
        """Test inheritance from HelixAgentError."""
        error = AuthenticationError("Invalid API key")
        self.assertIsInstance(error, HelixAgentError)

    def test_with_status_code(self):
        """Test with status code."""
        error = AuthenticationError("Unauthorized", status_code=401)
        self.assertEqual(str(error), "[401] Unauthorized")


class TestRateLimitError(unittest.TestCase):
    """Test RateLimitError exception."""

    def test_inheritance(self):
        """Test inheritance from HelixAgentError."""
        error = RateLimitError("Rate limit exceeded")
        self.assertIsInstance(error, HelixAgentError)

    def test_with_retry_after(self):
        """Test with retry_after."""
        error = RateLimitError("Rate limited", retry_after=30)
        self.assertEqual(error.retry_after, 30)

    def test_without_retry_after(self):
        """Test without retry_after."""
        error = RateLimitError("Rate limited")
        self.assertIsNone(error.retry_after)

    def test_with_all_params(self):
        """Test with all parameters."""
        error = RateLimitError(
            "Too many requests",
            retry_after=60,
            status_code=429,
            response={"error": "rate_limit"},
        )
        self.assertEqual(error.retry_after, 60)
        self.assertEqual(error.status_code, 429)


class TestAPIError(unittest.TestCase):
    """Test APIError exception."""

    def test_inheritance(self):
        """Test inheritance from HelixAgentError."""
        error = APIError("API error")
        self.assertIsInstance(error, HelixAgentError)


class TestConnectionError(unittest.TestCase):
    """Test ConnectionError exception."""

    def test_inheritance(self):
        """Test inheritance from HelixAgentError."""
        error = ConnectionError("Connection failed")
        self.assertIsInstance(error, HelixAgentError)


class TestValidationError(unittest.TestCase):
    """Test ValidationError exception."""

    def test_inheritance(self):
        """Test inheritance from HelixAgentError."""
        error = ValidationError("Invalid input")
        self.assertIsInstance(error, HelixAgentError)


class TestTimeoutError(unittest.TestCase):
    """Test TimeoutError exception."""

    def test_inheritance(self):
        """Test inheritance from HelixAgentError."""
        error = TimeoutError("Request timed out")
        self.assertIsInstance(error, HelixAgentError)


class TestRaiseForStatus(unittest.TestCase):
    """Test raise_for_status function."""

    def test_401_raises_authentication_error(self):
        """Test 401 raises AuthenticationError."""
        with self.assertRaises(AuthenticationError) as ctx:
            raise_for_status(401, {"error": {"message": "Invalid token"}})
        self.assertEqual(ctx.exception.status_code, 401)
        self.assertEqual(ctx.exception.message, "Invalid token")

    def test_429_raises_rate_limit_error(self):
        """Test 429 raises RateLimitError."""
        with self.assertRaises(RateLimitError) as ctx:
            raise_for_status(429, {"error": {"message": "Too fast"}, "retry_after": 30})
        self.assertEqual(ctx.exception.status_code, 429)
        self.assertEqual(ctx.exception.retry_after, 30)

    def test_400_raises_validation_error(self):
        """Test 400 raises ValidationError."""
        with self.assertRaises(ValidationError) as ctx:
            raise_for_status(400, {"error": {"message": "Bad request"}})
        self.assertEqual(ctx.exception.status_code, 400)

    def test_404_raises_validation_error(self):
        """Test 404 raises ValidationError."""
        with self.assertRaises(ValidationError) as ctx:
            raise_for_status(404, {"error": {"message": "Not found"}})
        self.assertEqual(ctx.exception.status_code, 404)

    def test_500_raises_api_error(self):
        """Test 500 raises APIError."""
        with self.assertRaises(APIError) as ctx:
            raise_for_status(500, {"error": {"message": "Internal error"}})
        self.assertEqual(ctx.exception.status_code, 500)

    def test_502_raises_api_error(self):
        """Test 502 raises APIError."""
        with self.assertRaises(APIError) as ctx:
            raise_for_status(502, {"error": {"message": "Bad gateway"}})
        self.assertEqual(ctx.exception.status_code, 502)

    def test_error_as_string(self):
        """Test error message as string."""
        with self.assertRaises(ValidationError) as ctx:
            raise_for_status(400, {"error": "Simple error message"})
        self.assertEqual(ctx.exception.message, "Simple error message")

    def test_unknown_error(self):
        """Test unknown error format."""
        with self.assertRaises(ValidationError) as ctx:
            raise_for_status(400, {})
        self.assertEqual(ctx.exception.message, "Unknown error")


if __name__ == "__main__":
    unittest.main()
