package purpleair

type ApiError int64

const (
	Undefined               ApiError = iota
	ApiKeyMissingError               //403 No API key was found in the request.
	ApiKeyTypeMismatchError          //403 The provided key was of the wrong type (READ or WRITE).
	ApiKeyInvalidError               //403 The provided key was not valid.
	ApiKeyRestrictedError            //403 The provided API key is restricted to certain hosts or referrers.
	ApiDisabledError                 //403 API calls to this endpoint have been restricted for your key. Please try again later or contact PurpleAir for more information.
	InvalidTokenError                //403 The provided token was not valid.
)

func (e ApiError) String() string {
	switch e {
	case ApiKeyMissingError:
		return "ApiKeyMissingError"
	case ApiKeyTypeMismatchError:
		return "ApiKeyTypeMismatchError"
	case ApiKeyInvalidError:
		return "ApiKeyInvalidError"
	case ApiKeyRestrictedError:
		return "ApiKeyRestrictedError"
	case ApiDisabledError:
		return "ApiDisabledError"
	case InvalidTokenError:
		return "InvalidTokenError"
	}
	return "unknown"
}
