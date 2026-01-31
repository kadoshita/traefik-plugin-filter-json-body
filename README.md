# traefik-plugin-filter-json-body
A Traefik plugin that filters requests based on JSON body content

## Overview
This plugin filters incoming requests based on JSON body content and rejects requests that match specified conditions.

## Configuration

### Parameters
- `rules`: Array of filtering rules (at least one rule is required)
  - `path`: Request path (exact match, required)
  - `method`: HTTP method (exact match, required)
  - `bodyPath`: Path to the target field in JSON body (XPath format, required)
  - `bodyValueCondition`: Value matching condition (regular expression, required)

### Behavior
- Multiple rules can be specified in the `rules` array
- Within each rule, all parameters (`path`, `method`, `bodyPath`, `bodyValueCondition`) must match for that rule to be considered matched (AND logic)
- If any one of the rules matches, the request returns 403 Forbidden (OR logic)
- Requests not matching any rules are passed to the next handler
- Filtering is skipped in the following cases:
  - Content-Type is not application/json or application/*+json
  - Body size exceeds 10MB
  - Body read error occurs
  - JSON parsing fails
  - Specified `bodyPath` does not exist in the JSON body

## Configuration Examples

### Example 1: Reject specific string value
```yaml
rules:
  - path: /api/test
    method: POST
    bodyPath: key
    bodyValueCondition: ^value$
```

### Example 2: Inspect value in nested object
```yaml
rules:
  - path: /api/test
    method: POST
    bodyPath: //nestedObject/innerString
    bodyValueCondition: ^inner$
```

### Example 3: Inspect value in object within array
```yaml
rules:
  - path: /api/test
    method: POST
    bodyPath: //arrayOfObjects/*/objString[text()='obj2']
    bodyValueCondition: ^obj2$
```
