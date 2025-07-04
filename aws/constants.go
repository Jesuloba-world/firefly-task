package aws

// Instance states
const (
	// InstanceStateRunning represents a running EC2 instance
	InstanceStateRunning = "running"
	// InstanceStateStopped represents a stopped EC2 instance
	InstanceStateStopped = "stopped"
	// InstanceStatePending represents a pending EC2 instance
	InstanceStatePending = "pending"
	// InstanceStateStopping represents a stopping EC2 instance
	InstanceStateStopping = "stopping"
	// InstanceStateTerminated represents a terminated EC2 instance
	InstanceStateTerminated = "terminated"
)

// AWS API error patterns
const (
	// ErrorPatternInvalidInstanceID is the error pattern for invalid instance IDs
	ErrorPatternInvalidInstanceID = "InvalidInstanceID"
	// ErrorPatternInstanceNotFound is the error pattern for instance not found
	ErrorPatternInstanceNotFound = "InvalidInstanceID.NotFound"
	// ErrorPatternThrottling is the error pattern for throttling errors
	ErrorPatternThrottling = "Throttling"
	// ErrorPatternRequestLimitExceeded is the error pattern for request limit exceeded
	ErrorPatternRequestLimitExceeded = "RequestLimitExceeded"
	// ErrorPatternServiceUnavailable is the error pattern for service unavailable
	ErrorPatternServiceUnavailable = "ServiceUnavailable"
	// ErrorPatternInternalError is the error pattern for internal errors
	ErrorPatternInternalError = "InternalError"
	// ErrorPatternConnectionReset is the error pattern for connection reset
	ErrorPatternConnectionReset = "connection reset"
	// ErrorPatternTimeout is the error pattern for timeout errors
	ErrorPatternTimeout = "timeout"
	// ErrorPatternTemporaryFailure is the error pattern for temporary failures
	ErrorPatternTemporaryFailure = "temporary failure"
	// ErrorPatternInvalidParameter is the error pattern for invalid parameters
	ErrorPatternInvalidParameter = "InvalidParameter"
	// ErrorPatternUnauthorizedOperation is the error pattern for unauthorized operations
	ErrorPatternUnauthorizedOperation = "UnauthorizedOperation"
)
