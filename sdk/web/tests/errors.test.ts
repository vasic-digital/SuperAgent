/**
 * Error Classes Tests
 */

import {
  HelixAgentError,
  AuthenticationError,
  RateLimitError,
  APIError,
  ValidationError,
  NetworkError,
  TimeoutError,
  ProviderError,
  DebateError,
} from '../src/errors';

describe('HelixAgentError', () => {
  it('should create error with message', () => {
    const error = new HelixAgentError('Test error');
    expect(error.message).toBe('Test error');
    expect(error.name).toBe('HelixAgentError');
  });

  it('should be instance of Error', () => {
    const error = new HelixAgentError('Test');
    expect(error).toBeInstanceOf(Error);
  });
});

describe('AuthenticationError', () => {
  it('should create with default message', () => {
    const error = new AuthenticationError();
    expect(error.message).toBe('Authentication failed');
    expect(error.name).toBe('AuthenticationError');
  });

  it('should create with custom message', () => {
    const error = new AuthenticationError('Invalid token');
    expect(error.message).toBe('Invalid token');
  });

  it('should extend HelixAgentError', () => {
    const error = new AuthenticationError();
    expect(error).toBeInstanceOf(HelixAgentError);
  });
});

describe('RateLimitError', () => {
  it('should create with default message and no retryAfter', () => {
    const error = new RateLimitError();
    expect(error.message).toBe('Rate limit exceeded');
    expect(error.retryAfter).toBeNull();
    expect(error.name).toBe('RateLimitError');
  });

  it('should create with custom message and retryAfter', () => {
    const error = new RateLimitError('Too many requests', 60);
    expect(error.message).toBe('Too many requests');
    expect(error.retryAfter).toBe(60);
  });

  it('should extend HelixAgentError', () => {
    const error = new RateLimitError();
    expect(error).toBeInstanceOf(HelixAgentError);
  });
});

describe('APIError', () => {
  it('should create with required parameters', () => {
    const error = new APIError('Server error', 500);
    expect(error.message).toBe('Server error');
    expect(error.statusCode).toBe(500);
    expect(error.name).toBe('APIError');
  });

  it('should create with all parameters', () => {
    const error = new APIError('Bad request', 400, {
      type: 'invalid_request',
      param: 'model',
      code: 'model_not_found',
    });
    expect(error.type).toBe('invalid_request');
    expect(error.param).toBe('model');
    expect(error.code).toBe('model_not_found');
  });

  it('should extend HelixAgentError', () => {
    const error = new APIError('Error', 500);
    expect(error).toBeInstanceOf(HelixAgentError);
  });
});

describe('ValidationError', () => {
  it('should create with default message', () => {
    const error = new ValidationError();
    expect(error.message).toBe('Validation failed');
    expect(error.name).toBe('ValidationError');
  });

  it('should create with custom message', () => {
    const error = new ValidationError('Invalid input');
    expect(error.message).toBe('Invalid input');
  });

  it('should extend HelixAgentError', () => {
    const error = new ValidationError();
    expect(error).toBeInstanceOf(HelixAgentError);
  });
});

describe('NetworkError', () => {
  it('should create with default message', () => {
    const error = new NetworkError();
    expect(error.message).toBe('Network error occurred');
    expect(error.name).toBe('NetworkError');
  });

  it('should create with custom message', () => {
    const error = new NetworkError('Connection refused');
    expect(error.message).toBe('Connection refused');
  });

  it('should extend HelixAgentError', () => {
    const error = new NetworkError();
    expect(error).toBeInstanceOf(HelixAgentError);
  });
});

describe('TimeoutError', () => {
  it('should create with default message', () => {
    const error = new TimeoutError();
    expect(error.message).toBe('Request timeout');
    expect(error.name).toBe('TimeoutError');
  });

  it('should create with custom message', () => {
    const error = new TimeoutError('Request timed out after 30s');
    expect(error.message).toBe('Request timed out after 30s');
  });

  it('should extend HelixAgentError', () => {
    const error = new TimeoutError();
    expect(error).toBeInstanceOf(HelixAgentError);
  });
});

describe('ProviderError', () => {
  it('should create with message and provider', () => {
    const error = new ProviderError('Provider unavailable', 'openai');
    expect(error.message).toBe('Provider unavailable');
    expect(error.provider).toBe('openai');
    expect(error.name).toBe('ProviderError');
  });

  it('should extend HelixAgentError', () => {
    const error = new ProviderError('Error', 'test');
    expect(error).toBeInstanceOf(HelixAgentError);
  });
});

describe('DebateError', () => {
  it('should create with message and debateId', () => {
    const error = new DebateError('Debate failed', 'debate-123');
    expect(error.message).toBe('Debate failed');
    expect(error.debateId).toBe('debate-123');
    expect(error.name).toBe('DebateError');
  });

  it('should extend HelixAgentError', () => {
    const error = new DebateError('Error', 'test');
    expect(error).toBeInstanceOf(HelixAgentError);
  });
});
