/**
 * ErrorExplanations - Maps Facebook API error codes to human-readable explanations.
 *
 * Used by the Log system to provide contextual help when FB API errors are
 * encountered. Each entry includes a message, suggestion, and link to docs.
 */
var ErrorExplanations = {
  /**
   * Look up an error explanation by Facebook API error response object.
   * @param {Object} error - FB API error response (may contain error.error_code or error.error.code)
   * @returns {Object|null} Explanation object with message, suggestion, docs fields, or null
   */
  lookup: function(error) {
    var code = null;
    if (error && typeof error === 'object') {
      code = error.error_code || (error.error && error.error.code);
    }
    if (code && ErrorExplanations.codes[code]) {
      return ErrorExplanations.codes[code];
    }
    return null;
  },

  codes: {
    1: {
      message: "Unknown error occurred.",
      suggestion: "Try again. If the problem persists, check the Graph API status page.",
      docs: "https://developers.facebook.com/docs/graph-api/guides/error-handling"
    },
    2: {
      message: "Service temporarily unavailable.",
      suggestion: "Wait a moment and retry the request.",
      docs: "https://developers.facebook.com/docs/graph-api/guides/error-handling"
    },
    4: {
      message: "Application request limit reached.",
      suggestion: "Reduce the frequency of API calls or request a rate limit increase.",
      docs: "https://developers.facebook.com/docs/graph-api/overview/rate-limiting"
    },
    10: {
      message: "Permission denied for this API call.",
      suggestion: "Check that your app has the required permissions. You may need to request additional permissions via FB.login().",
      docs: "https://developers.facebook.com/docs/permissions"
    },
    100: {
      message: "Invalid parameter.",
      suggestion: "Check that all API parameters are correct and properly formatted.",
      docs: "https://developers.facebook.com/docs/graph-api/reference"
    },
    102: {
      message: "Session expired or invalid.",
      suggestion: "Re-authenticate by calling FB.login() to get a fresh session.",
      docs: "https://developers.facebook.com/docs/facebook-login/access-tokens"
    },
    104: {
      message: "Access token required.",
      suggestion: "Ensure the user is logged in. Call FB.login() first.",
      docs: "https://developers.facebook.com/docs/facebook-login"
    },
    190: {
      message: "Access token has expired or is invalid.",
      suggestion: "Re-authenticate with FB.login() to get a new access token.",
      docs: "https://developers.facebook.com/docs/facebook-login/access-tokens/debugging-and-error-handling"
    },
    200: {
      message: "Permission not granted by user.",
      suggestion: "Request the necessary permission scope via FB.login({scope: '...'}).",
      docs: "https://developers.facebook.com/docs/permissions"
    },
    341: {
      message: "Too many post requests.",
      suggestion: "Reduce posting frequency. Rate limits apply per user and per app.",
      docs: "https://developers.facebook.com/docs/graph-api/overview/rate-limiting"
    },
    506: {
      message: "Duplicate post detected.",
      suggestion: "Change the content of your post, as Facebook blocks identical repeated posts.",
      docs: "https://developers.facebook.com/docs/graph-api/guides/error-handling"
    },
    2500: {
      message: "Extended permission required.",
      suggestion: "This API call requires specific extended permissions. Use FB.login with the right scope.",
      docs: "https://developers.facebook.com/docs/permissions"
    }
  }
};
