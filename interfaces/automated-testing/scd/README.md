# Strategic Deconfliction testing interfaces

These interfaces define how the Strategic Deconfliction test driver interact with the system under test.

## Viewing locally

To view these YAML files locally:

```shell script
docker run -it --rm -p 8080:8080 \
  -v $(pwd)/observation.yaml:/usr/share/nginx/html/swagger.yaml \
  -e PORT=8080 -e SPEC_URL=swagger.yaml redocly/redoc
```

OR

```shell script
docker run -it --rm -p 8080:8080 \
  -v $(pwd)/injection.yaml:/usr/share/nginx/html/swagger.yaml \
  -e PORT=8080 -e SPEC_URL=swagger.yaml redocly/redoc
```

...then visit [http://localhost:8080/](http://localhost:8080/) in a browser.

## System Diagram

TBC