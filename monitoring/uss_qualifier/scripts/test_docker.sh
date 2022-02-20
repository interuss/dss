#!/usr/bin/env bash

set -eo pipefail

cd "$(dirname "$0")"

project_name=qualifier_sandbox

echo "Ensure the environment is clean"
echo "============="
docker-compose -f docker-compose_qualifier_mocks.yaml -p $project_name down

if [ "$CI" = "true" ]; then
    echo "Rebuild images without cache"
    docker-compose -f docker-compose_qualifier_mocks.yaml -p $project_name build --no-cache
fi

echo "Start mocks"
echo "============="
docker-compose -f docker-compose_qualifier_mocks.yaml -p $project_name up --remove-orphans -d

echo "Wait for mocks to be healthy"
echo "============="

services=( "$(docker-compose -f docker-compose_qualifier_mocks.yaml config --services)" )
for service_name in "${services[@]}"; do
    container_name="${project_name}_${service_name}_1"
    retry=0
    max_retry=3
    until [ "$(docker inspect -f \{\{.State.Health.Status\}\} "${container_name}")" == "healthy" ]; do
        if [ "$retry" -gt "$max_retry" ]; then
            echo "$container_name logs:"
            docker logs "$container_name"
            echo "$container_name didn't properly start. Exit." && exit 1
        fi
        sleep 1
        retry=$((retry+1))
    done
done

echo "Generate simulation data and run the qualifier."
echo "============="
pushd ..
./bin/generate_rid_test_definition.sh
./run_locally.sh
popd

echo "Clean up"
echo "============="
docker-compose -f docker-compose_qualifier_mocks.yaml -p $project_name down
