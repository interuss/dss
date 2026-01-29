#!/usr/bin/env bash
# shellcheck disable=SC2010

set -eo pipefail

OS=$(uname)
if [[ "$OS" == "Darwin" ]]; then
	# OSX uses BSD readlink
	SCRIPTDIR="$(dirname "$0")"
else
	SCRIPTDIR=$(readlink -e "$(dirname "$0")")
fi

BASEDIR="${SCRIPTDIR}/../.."
cd "${SCRIPTDIR}"

# Extract version
CRDB_RID=$(ls rid | grep -oE 'upto-v[0-9]+\.[0-9]+\.[0-9]+' | sed 's/upto-v//' | sort -V | tail -n 1)
CRDB_SCD=$(ls scd | grep -oE 'upto-v[0-9]+\.[0-9]+\.[0-9]+' | sed 's/upto-v//' | sort -V | tail -n 1)
YBDB_RID=$(ls yugabyte/rid | grep -oE 'upto-v[0-9]+\.[0-9]+\.[0-9]+' | sed 's/upto-v//' | sort -V | tail -n 1)
YBDB_SCD=$(ls yugabyte/scd | grep -oE 'upto-v[0-9]+\.[0-9]+\.[0-9]+' | sed 's/upto-v//' | sort -V | tail -n 1)
AUX=$(ls yugabyte/aux_ | grep -oE 'upto-v[0-9]+\.[0-9]+\.[0-9]+' | sed 's/upto-v//' | sort -V | tail -n 1)

# Replace terraform latests
sed -i -E "s/(rid_db_schema = var\.desired_rid_db_version == \"latest\" \? \(var\.datastore_type == \"cockroachdb\" \? ).*\)/\1\"$CRDB_RID\" : \"$YBDB_RID\")/" "${BASEDIR}/deploy/infrastructure/dependencies/terraform-commons-dss/default_latest.tf"
sed -i -E "s/(scd_db_schema = var\.desired_scd_db_version == \"latest\" \? \(var\.datastore_type == \"cockroachdb\" \? ).*\)/\1\"$CRDB_SCD\" : \"$YBDB_SCD\")/" "${BASEDIR}/deploy/infrastructure/dependencies/terraform-commons-dss/default_latest.tf"
sed -i -E "s/(aux_db_schema = var\.desired_aux_db_version == \"latest\" \? ).* :/\1\"$AUX\" :/" "${BASEDIR}/deploy/infrastructure/dependencies/terraform-commons-dss/default_latest.tf"

# Replace tanka examples / latests
sed -i -E "s/(desired_rid_db_version: ).*/\1'$YBDB_RID',/" "${BASEDIR}/deploy/services/tanka/examples/minikube/main.jsonnet"
sed -i -E "s/(desired_scd_db_version: ).*/\1'$YBDB_SCD',/" "${BASEDIR}/deploy/services/tanka/examples/minikube/main.jsonnet"
sed -i -E "s/(desired_aux_db_version: ).*/\1'$AUX',/" "${BASEDIR}/deploy/services/tanka/examples/minikube/main.jsonnet"

for file in "${BASEDIR}/deploy/services/tanka/metadata_base.libsonnet" "${BASEDIR}/deploy/services/tanka/examples/schema_manager/main.jsonnet" "${BASEDIR}/deploy/services/tanka/examples/minimum/main.jsonnet"; do
    sed -i -E "s/(desired_rid_db_version: ).*/\1'$CRDB_RID',/" "${file}"
    sed -i -E "s/(desired_scd_db_version: ).*/\1'$CRDB_SCD',/" "${file}"
    sed -i -E "s/(desired_aux_db_version: ).*/\1'$AUX',/" "${file}"
done

# Replace helm latests
sed -i -E -e "s/(\\\$schemas := dict ).*}}/\1\"rid\" \"$CRDB_RID\" \"scd\" \"$CRDB_SCD\" \"aux_\" \"$AUX\" }}/" "${BASEDIR}/deploy/services/helm-charts/dss/templates/schema-manager.yaml"
sed -i -E -e "s/(\\\$schemas = dict ).*}}/\1\"rid\" \"$YBDB_RID\" \"scd\" \"$YBDB_SCD\" \"aux_\" \"$AUX\" }}/" "${BASEDIR}/deploy/services/helm-charts/dss/templates/schema-manager.yaml"

# Generate libsonnet files with list of migrations
cat <<EOF > rid.libsonnet
{
  data:{
$(find rid -type f -print0 | sort -z | xargs -0 -I {} basename {} | awk '{print "    \"" $1 "\": importstr \"rid/" $1 "\","}')
  },
}
EOF

cat <<EOF > scd.libsonnet
{
  data:{
$(find scd -type f -print0 | sort -z | xargs -0 -I {} basename {} | awk '{print "    \"" $1 "\": importstr \"scd/" $1 "\","}')
  },
}
EOF

cat <<EOF > aux_.libsonnet
{
  data:{
$(find aux_ -type f -print0 | sort -z | xargs -0 -I {} basename {} | awk '{print "    \"" $1 "\": importstr \"aux_/" $1 "\","}')
  },
}
EOF

# Extract major versions
CRDB_RID_MAJOR=$(echo "$CRDB_RID" | cut -d. -f1)
CRDB_SCD_MAJOR=$(echo "$CRDB_SCD" | cut -d. -f1)
YBDB_RID_MAJOR=$(echo "$YBDB_RID" | cut -d. -f1)
YBDB_SCD_MAJOR=$(echo "$YBDB_SCD" | cut -d. -f1)
AUX_MAJOR=$(echo "$AUX" | cut -d. -f1)

# Replace major versions in datastore files
sed -i -E "s/(currentCrdbMajorSchemaVersion.*= )[0-9]/\1$CRDB_SCD_MAJOR/" "${BASEDIR}/pkg/scd/store/datastore/store.go"
sed -i -E "s/(currentYugabyteMajorSchemaVersion.*= )[0-9]/\1$YBDB_SCD_MAJOR/" "${BASEDIR}/pkg/scd/store/datastore/store.go"

sed -i -E "s/(currentCrdbMajorSchemaVersion.*= )[0-9]/\1$CRDB_RID_MAJOR/" "${BASEDIR}/pkg/rid/store/datastore/store.go"
sed -i -E "s/(currentYugabyteMajorSchemaVersion.*= )[0-9]/\1$YBDB_RID_MAJOR/" "${BASEDIR}/pkg/rid/store/datastore/store.go"

sed -i -E "s/(currentCrdbMajorSchemaVersion.*= )[0-9]/\1$AUX_MAJOR/" "${BASEDIR}/pkg/aux_/store/datastore/store.go"
sed -i -E "s/(currentYugabyteMajorSchemaVersion.*= )[0-9]/\1$AUX_MAJOR/" "${BASEDIR}/pkg/aux_/store/datastore/store.go"
