# USS Qualifier webapp automation interface

This interface defines how to instruct
[uss_qualifier](../../monitoring/uss_qualifier)'s
[webapp](../../monitoring/uss_qualifier/webapp) to execute the test suite and
how to retrieve the results.

## Viewing locally
To view this YAML files locally:

```shell script
docker run -it --rm -p 8080:8080 \
  -v $(pwd)/uss_qualifier.yaml:/usr/share/nginx/html/swagger.yaml \
  -e PORT=8080 -e SPEC_URL=swagger.yaml redocly/redoc
```

...then visit [http://localhost:8080/](http://localhost:8080/) in a browser.
