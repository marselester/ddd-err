# DDD Error Handling

This small project is an error handling example based on Ben Johnson's article.
Based on https://middlemost.com/failure-is-your-domain/ error can be:

- well-defined errors. These allow us to manage our application flow
  because we know what to expect and can work with them on a case-by-case basis.
- undefined errors. It can also occur when APIs we depend on add additional
  errors conditions after we've integrated our code with them.

Error consumers:

- app itself can recover from error states (requires error codes).
- end user requires human-readable message. API undefined errors must not be
  shown, e.g., Postgres error can reveal db schema.
- operator should be able to debug and see all errors including stack trace.
