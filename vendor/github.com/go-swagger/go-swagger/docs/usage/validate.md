# Validate a swagger spec

The toolkit has a command to validate swagger specifications for you.
It includes a full json-schema validator and adds some extra validations to ensure the spec is valid.

<!--more-->

### Usage

To validate a specification:

```
swagger validate [http-url|filepath]
```

### Swagger 2.0 resources

* Specification Documentation: https://github.com/swagger-api/swagger-spec/blob/master/versions/2.0.md
* JSON Schema: https://github.com/swagger-api/swagger-spec/blob/master/schemas/v2.0/schema.json

### Semantic Validation

All the rules the validator tool supports:

*	validate against jsonschema
*	validate extra rules outlined [here](https://github.com/apigee-127/sway/blob/master/docs/versions/2.0.md#semantic-validation)

Rule | Severity
-----|---------
definition can't declare a property that's already defined by one of its ancestors | Error
definition's ancestor can't be a descendant of the same model | Error
each security reference should contain only unique scopes | Warning
each security scope in a security definition should be unique | Warning
path parameter declarations do not allow empty names _(`/path/{}` is not valid)_ | Error
each api path should be non-verbatim (account for path param names) unique per method | Error
each path parameter should correspond to a parameter placeholder and vice versa | Error
each referencable definition must have references | Warning
each definition property listed in the required array must be defined in the properties of the model | Error
each parameter should have a unique `name` and `in` combination | Error
each operation must have an unique `operationId` | Error
each operation should have only 1 parameter of type body | Error
each operation cannot have both a body parameter and a formData parameter | Error
each reference must point to a valid object | Error
every default value that is specified must validate against the schema for that property | Error
every example that is specified must validate against the schema for that property | Error
items property is required for all schemas/definitions of type `array` | Error
